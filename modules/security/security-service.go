package security

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/svachaj/sambar-wall/utils"
)

type ISecurityService interface {
	GenerateVerificationCode() string
	SaveVerificationCode(email string, code string) error
	SendVerificationCode(email string, code string, host string) error
	FinalizeLogin(email, confirmationCode string) (userId int, err error)
}

type SecurityService struct {
	db           *sqlx.DB
	emailService utils.IEmailService
}

func NewSecurityService(db *sqlx.DB, emailService utils.IEmailService) ISecurityService {
	return &SecurityService{db: db, emailService: emailService}
}

func (s *SecurityService) GenerateVerificationCode() string {
	code := rand.Intn(100000)
	if code < 10000 {
		code += 10000
	}

	return fmt.Sprintf("%v", code)
}

func (s *SecurityService) SaveVerificationCode(email string, code string) error {

	var query string
	if s.db.DriverName() == "postgres" {
		query = fmt.Sprintf("INSERT INTO t_system_registration_code (id, email, code, createdate) VALUES ((select max(id)+1 from t_system_registration_code), '%v', '1234', '%v')", email, time.Now().Format("2006-01-02 15:04:05"))
	} else {
		query = fmt.Sprintf("INSERT INTO t_system_registration_code (email, code, createdate) VALUES ('%v', '%v', '%v')", email, code, time.Now().Format("2006-01-02 15:04:05"))
	}

	_, err := s.db.Exec(query)

	return err
}

func (s *SecurityService) SendVerificationCode(email string, code string, host string) error {

	subject := "Sambar Lezecká Stěna - přihlašovací kód"
	// crypt email and code as query string
	queryString := fmt.Sprintf("%v;%v", email, code)
	queryStringEncoded := utils.Encrypt(queryString)

	body := fmt.Sprintf("<span style='letter-spacing: 0.75px;'>Tvůj jednorázový přihlašovací kód je: <a target='_blank' href='%v/sign-me-in?c=%v' style='color: rgb(219 39 119);' ><span style='font-size:20px;letter-spacing: 2px;'>%v</span></a>", host, queryStringEncoded, code)
	body += "<br><br>"
	body += "<span style='letter-spacing: 0.75px;'>Kliknutím na kód je možné se rovnou přihlásit.</span>"
	body += "<br><br>"
	body += "<span style='font-size:13px;color: #f40d0d;letter-spacing: 0.5px;'>Tento kód je platný pouze 10 minut.</span>"
	body += "<br>"
	body += "<span style='font-size:13px;color: #4d4d4d;letter-spacing: 0.5px;'>Pokud jste o tento kód nepožádali, ignorujte tento email.</span>"

	return s.emailService.SendEmail(subject, body, email)
}

func (s *SecurityService) FinalizeLogin(email, confirmationCode string) (userId int, err error) {

	// check confirmation code
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM t_system_registration_code WHERE email = '%v' AND code = '%v' AND createdate > '%v'", email, confirmationCode, time.Now().Add(-time.Minute*10).Format("2006-01-02 15:04:05"))
	err = s.db.Get(&count, query)

	if err != nil {
		return -1, err
	}

	if count == 0 {
		return -1, fmt.Errorf(AGREEMENT_ERROR_BAD_CONFIRMATION_CODE)
	}

	// we want to create a new user with the email if it does not exist
	query = fmt.Sprintf(`IF NOT EXISTS (SELECT 1 FROM t_system_user WHERE UserName = '%[1]v')
	BEGIN
		INSERT INTO t_system_user (email, username, CreateDate, IsActivated, IsDeleted, IsEnabled) VALUES ('%[1]v', '%[1]v', getdate(), 1, 0 ,1);
	END;`, email)
	_, _ = s.db.Exec(query)

	// get user id
	query = fmt.Sprintf("SELECT ID FROM t_system_user WHERE UserName = '%v'", email)
	err = s.db.Get(&userId, query)

	if err != nil {
		return -1, err
	}

	// set last logon date
	query = fmt.Sprintf("UPDATE t_system_user SET LastLogonDate = getdate() WHERE UserName = '%v'", email)
	_, _ = s.db.Exec(query)

	// if everything is ok, delete the confirmation code
	query = fmt.Sprintf("DELETE FROM t_system_registration_code WHERE email = '%v'", email)
	_, _ = s.db.Exec(query)

	return userId, nil
}

const AGREEMENT_ERROR_BAD_CONFIRMATION_CODE = "Neplatný přihlašovací kód"
