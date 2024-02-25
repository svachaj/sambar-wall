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
}

type AgreementService struct {
	db           *sqlx.DB
	emailService *utils.EmailService
}

func NewAgreementService(db *sqlx.DB, emailService *utils.EmailService) IAgreementService {
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

	query := fmt.Sprintf("INSERT INTO t_system_registration_code (email, code, createdate) VALUES ('%v', '%v', '%v')", email, code, time.Now().Format("2006-01-02 15:04:05"))
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
