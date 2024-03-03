package agreement

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/svachaj/sambar-wall/utils"
)

type IAgreementService interface {
	EmailExists(email string) (bool, error)
	GenerateVerificationCode() string
	SaveVerificationCode(email string, code string) error
	SendVerificationCode(email string, code string) error
	FinalizeAgreement(email, firstName, lastName, birthDate, confirmationCode string) error
}

type AgreementService struct {
	db           *sqlx.DB
	emailService utils.IEmailService
}

func NewAgreementService(db *sqlx.DB, emailService utils.IEmailService) IAgreementService {
	return &AgreementService{db: db, emailService: emailService}
}

func (s *AgreementService) EmailExists(email string) (bool, error) {

	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM t_system_wall_user WHERE isenabled = 'true' AND lower(email) = '%v'", strings.ToLower(email))
	err := s.db.Get(&count, query)

	if err != nil {
		return false, err
	}

	return count > 0, nil

}

func (s *AgreementService) GenerateVerificationCode() string {
	code := rand.Intn(10000)
	if code < 1000 {
		code += 1000
	}

	return fmt.Sprintf("%v", code)
}

func (s *AgreementService) SaveVerificationCode(email string, code string) error {

	var query string
	if s.db.DriverName() == "postgres" {
		query = fmt.Sprintf("INSERT INTO t_system_registration_code (id, email, code, createdate) VALUES ((select max(id)+1 from t_system_registration_code), '%v', '1234', '%v')", email, time.Now().Format("2006-01-02 15:04:05"))
	} else {
		query = fmt.Sprintf("INSERT INTO t_system_registration_code (email, code, createdate) VALUES ('%v', '%v', '%v')", email, code, time.Now().Format("2006-01-02 15:04:05"))
	}

	_, err := s.db.Exec(query)

	if err != nil {
		return err
	}

	return nil
}

func (s *AgreementService) SendVerificationCode(email string, code string) error {
	err := s.emailService.SendEmail("Sambar Lezecká Stěna - Ověření emailu", fmt.Sprintf("Ověřovací kód: %v", code), email)
	if err != nil {
		return err
	}
	return nil
}

func (s *AgreementService) FinalizeAgreement(email, firstName, lastName, birthDate, confirmationCode string) error {

	// check confirmation code
	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM t_system_registration_code WHERE email = '%v' AND code = '%v' AND createdate > '%v'", email, confirmationCode, time.Now().Add(-time.Minute*10).Format("2006-01-02 15:04:05"))
	err := s.db.Get(&count, query)

	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf(AGREEMENT_ERROR_BAD_CONFIRMATION_CODE)
	}

	// finalize agreement
	if s.db.DriverName() == "postgres" {
		query = fmt.Sprintf("INSERT INTO t_system_wall_user (id, email, firstname, lastname, birthdate, isenabled, createdate, GDPR_confirmed, Rules_confirmed) VALUES ((select max(id)+1 from t_system_wall_user), '%v', '%v', '%v', '%v', 'true', '%v', 'true', 'true')", email, firstName, lastName, utils.NormalizeDate(birthDate), time.Now().Format("2006-01-02 15:04:05"))
	} else {
		query = fmt.Sprintf("INSERT INTO t_system_wall_user (email, firstname, lastname, birthdate, isenabled, createdate, GDPR_confirmed, Rules_confirmed) VALUES ('%v', '%v', '%v', '%v', 'true', '%v', 'true', 'true')", email, firstName, lastName, utils.NormalizeDate(birthDate), time.Now().Format("2006-01-02 15:04:05"))
	}
	_, err = s.db.Exec(query)

	if err != nil {
		return err
	}

	return nil
}

const AGREEMENT_ERROR_BAD_CONFIRMATION_CODE = "Neplatný ověřovací kód"
