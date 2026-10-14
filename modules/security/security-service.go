package security

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	mathRand "math/rand"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/svachaj/sambar-wall/config"
	"github.com/svachaj/sambar-wall/utils"
)

type ISecurityService interface {
	GenerateVerificationCode() string
	SaveVerificationCode(email string, code string) error
	SendVerificationCode(email string, code string, host string) error
	FinalizeLogin(email, confirmationCode string) (userId int, roles []string, err error)
	GetConfig() *config.Config
}

type SecurityService struct {
	db           *sqlx.DB
	emailService utils.IEmailService
	_config      *config.Config
}

func NewSecurityService(db *sqlx.DB, emailService utils.IEmailService, config *config.Config) ISecurityService {
	return &SecurityService{db: db, emailService: emailService, _config: config}
}

func (s *SecurityService) GetConfig() *config.Config {
	return s._config
}

func (s *SecurityService) GenerateVerificationCode() string {
	b := make([]byte, 4)
	_, err := rand.Read(b)
	if err != nil {
		return fmt.Sprintf("%05d", 10000+mathRand.Intn(90000))
	}
	code := binary.BigEndian.Uint32(b) % 100000
	if code < 10000 {
		code += 10000
	}

	return fmt.Sprintf("%d", code)
}

func (s *SecurityService) SaveVerificationCode(email string, code string) error {
	lowerEmail := strings.ToLower(email)

	// createdate comes from the database clock, the same one FinalizeLogin compares against
	_, err := s.db.Exec(
		"INSERT INTO t_system_registration_code (email, code, createdate) VALUES (@p1, @p2, SYSDATETIME())",
		lowerEmail, code,
	)
	return err
}

func (s *SecurityService) SendVerificationCode(email string, code string, host string) error {

	subject := "Lezecká stěna Kladno - přihlašovací kód"
	// crypt email and code as query string
	queryString := fmt.Sprintf("%v;%v", email, code)
	queryStringEncoded := utils.Encrypt(queryString, s.GetConfig().AppCryptoKey)

	body := "<p><strong>Lezecká stěna Kladno – přihlášení / registrace</strong></p>"
	body += fmt.Sprintf("<p style='letter-spacing: 0.75px;'>Váš jednorázový přihlašovací kód je: <a target='_blank' href='%v/sign-me-in?c=%v' style='color: rgb(219 39 119);' ><span style='font-size:20px;letter-spacing: 2px;'>%v</span></a></p>", host, queryStringEncoded, code)
	body += "<p style='letter-spacing: 0.75px;'>Kliknutím na kód je možné se rovnou přihlásit.</p>"
	body += "<p style='font-size:13px;color: #f40d0d;letter-spacing: 0.5px;'>Tento kód je platný pouze 10 minut.</p>"
	body += "<p style='font-size:13px;color: #4d4d4d;letter-spacing: 0.5px;'>Pokud jste o tento kód nepožádali, ignorujte tento email.</p>"

	return s.emailService.SendEmail(subject, body, email)
}

func (s *SecurityService) FinalizeLogin(email, confirmationCode string) (userId int, roles []string, err error) {
	lowerEmail := strings.ToLower(email)

	// check confirmation code exists and is within 10 minutes; the limit is computed in SQL because
	// go-mssqldb sends time.Time as datetimeoffset and comparing it with the datetime2 createdate
	// treats createdate as UTC, which stretched the validity by the UTC offset (e.g. to 2h10m)
	var count int
	err = s.db.Get(&count,
		"SELECT COUNT(*) FROM t_system_registration_code WHERE email = @p1 AND code = @p2 AND createdate > DATEADD(minute, -10, SYSDATETIME())",
		lowerEmail, confirmationCode)

	if err != nil {
		return -1, roles, err
	}
	if count == 0 {
		return -1, roles, fmt.Errorf(AGREEMENT_ERROR_BAD_CONFIRMATION_CODE)
	}

	// create user if not exists
	var exists int
	err = s.db.Get(&exists,
		"SELECT COUNT(*) FROM t_system_user WHERE UserName = @p1",
		lowerEmail)
	if err != nil {
		return -1, roles, err
	}
	if exists == 0 {
		_, err = s.db.Exec(
			"INSERT INTO t_system_user (email, username, CreateDate, IsActivated, IsDeleted, IsEnabled) VALUES (@p1, @p1, GETDATE(), 1, 0, 1)",
			lowerEmail)
		if err != nil {
			return -1, roles, err
		}
	}

	// get user id
	err = s.db.Get(&userId,
		"SELECT ID FROM t_system_user WHERE UserName = @p1",
		lowerEmail)
	if err != nil {
		return -1, roles, err
	}

	// get user roles (INNER JOIN so a user with no assigned role yields an empty
	// slice instead of a row with a NULL tsr.Code, which fails to scan into a string)
	var userRoles []string
	err = s.db.Select(&userRoles,
		`SELECT tsr.Code
		FROM t_system_user tsu
		INNER JOIN t_system_role_user tsru ON tsu.ID = tsru.UserID
		INNER JOIN t_system_role tsr ON tsr.ID = tsru.RoleId
		WHERE tsu.ID = @p1`, userId)
	if err != nil {
		return -1, roles, err
	}
	roles = userRoles

	// update last logon date
	_, err = s.db.Exec(
		"UPDATE t_system_user SET LastLogonDate = GETDATE() WHERE UserName = @p1",
		lowerEmail)
	if err != nil {
		return -1, roles, err
	}

	// delete the confirmation code (one-time use) - only the code that was actually redeemed,
	// so other still-valid outstanding codes for the same email aren't invalidated
	_, err = s.db.Exec(
		"DELETE FROM t_system_registration_code WHERE email = @p1 AND code = @p2",
		lowerEmail, confirmationCode)
	if err != nil {
		return -1, roles, err
	}

	return userId, roles, nil
}

const AGREEMENT_ERROR_BAD_CONFIRMATION_CODE = "Neplatný přihlašovací kód"
