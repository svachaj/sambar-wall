package agreement

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/svachaj/sambar-wall/modules/agreement/models"
	"github.com/svachaj/sambar-wall/utils"
)

type IAgreementService interface {
	EmailExists(email string) (bool, error)
	GenerateVerificationCode() string
	SaveVerificationCode(email string, code string) error
	SendVerificationCode(email string, code string) error
	FinalizeAgreement(email, firstName, lastName, birthDate, confirmationCode string, commercialAgreement bool) error
	ExportEmailsConfirmedForCommercialCommunication() (string, error)
	GetWallVisitors(searchQuery string) ([]models.WallVisitor, error)
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
	err := s.db.Get(&count, "SELECT COUNT(*) FROM t_system_wall_user WHERE isenabled = 'true' AND lower(email) = @p1", strings.ToLower(email))

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

	// createdate comes from the database clock, the same one FinalizeAgreement compares against
	_, err := s.db.Exec(
		"INSERT INTO t_system_registration_code (email, code, createdate) VALUES (@p1, @p2, SYSDATETIME())",
		email, code,
	)

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

func (s *AgreementService) FinalizeAgreement(email, firstName, lastName, birthDate, confirmationCode string, commercialAgreement bool) error {

	// check confirmation code (valid 10 minutes); the limit is computed in SQL, see SecurityService.FinalizeLogin
	var count int
	err := s.db.Get(&count,
		"SELECT COUNT(*) FROM t_system_registration_code WHERE email = @p1 AND code = @p2 AND createdate > DATEADD(minute, -10, SYSDATETIME())",
		email, confirmationCode,
	)

	if err != nil {
		return err
	}

	if count == 0 {
		return fmt.Errorf(AGREEMENT_ERROR_BAD_CONFIRMATION_CODE)
	}

	// finalize agreement
	_, err = s.db.Exec(
		"INSERT INTO t_system_wall_user (email, firstname, lastname, birthdate, isenabled, createdate, GDPR_confirmed, Rules_confirmed, commercial_confirmed) VALUES (@p1, @p2, @p3, @p4, 'true', @p5, 'true', 'true', @p6)",
		email, firstName, lastName, utils.NormalizeDate(birthDate), time.Now(), commercialAgreement,
	)

	if err != nil {
		return err
	}

	return nil
}

const AGREEMENT_ERROR_BAD_CONFIRMATION_CODE = "Neplatný ověřovací kód"

func (s *AgreementService) ExportEmailsConfirmedForCommercialCommunication() (string, error) {

	rows, err := s.db.Query("SELECT DISTINCT email FROM t_system_wall_user WHERE IsEnabled = 1 and commercial_confirmed = 1 and email IS NOT NULL AND email != ''")

	if err != nil {
		return "", err
	}
	defer rows.Close()

	var emails []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return "", err
		}
		emails = append(emails, email)
	}

	// Join with semicolon
	result := strings.Join(emails, ";")
	return result, nil
}

func (s *AgreementService) GetWallVisitors(searchQuery string) ([]models.WallVisitor, error) {

	// if search query is empty, return nothing
	if strings.TrimSpace(searchQuery) == "" {
		return []models.WallVisitor{}, nil
	}

	// Normalize search query for diacritic-insensitive search
	pattern := "%" + strings.ToLower(searchQuery) + "%"

	query := `
		SELECT firstname, lastname, email, birthdate, createdate 
		FROM t_system_wall_user 
		WHERE isenabled = 1 
		  AND GDPR_confirmed = 1 
		  AND Rules_confirmed = 1
		  AND (
			LOWER(firstname) COLLATE Latin1_General_CI_AI LIKE @p1 OR
			LOWER(lastname) COLLATE Latin1_General_CI_AI LIKE @p1 OR
			LOWER(email) COLLATE Latin1_General_CI_AI LIKE @p1
		  )
		ORDER BY createdate DESC`

	var visitors []models.WallVisitor
	err := s.db.Select(&visitors, query, pattern)

	if err != nil {
		return nil, err
	}

	return visitors, nil
}
