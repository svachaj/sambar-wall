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
	SendVisitorCardEmail(email, firstName, lastName string) error
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

func (s *AgreementService) SendVisitorCardEmail(email, firstName, lastName string) error {
	subject := "Sambar Lezecká Stěna - potvrzení registrace"
	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; color: #111827; max-width: 640px; margin: 0 auto; line-height: 1.6;">
			<div style="background: linear-gradient(90deg, #ec4899, #db2777); color: white; padding: 18px 24px; border-radius: 10px 10px 0 0;">
				<h2 style="margin: 0; font-size: 22px;">Sambar Lezecká Stěna</h2>
				<p style="margin: 6px 0 0 0; font-size: 14px;">Potvrzení registrace návštěvníka</p>
			</div>
			<div style="border: 1px solid #e5e7eb; border-top: none; border-radius: 0 0 10px 10px; padding: 24px;">
				<p style="margin-top: 0;">Dobrý den %v %v,</p>
				<p>vaše registrace byla úspěšně dokončena a byla vám vytvořena návštěvnická karta.</p>
				<p><strong>Účel:</strong> karta slouží pro ověření vstupu na recepci lezecké stěny.</p>
				<p><strong>Platnost:</strong> karta je platná 12 měsíců od dokončení registrace.</p>
				<p>Pokud by systém kartu při vstupu nenašel, recepce vás může dohledat ručně podle jména a e-mailu.</p>
				<p style="margin-bottom: 0; color: #6b7280; font-size: 13px;">Tento e-mail je informační, odpovědi na něj nejsou monitorovány.</p>
			</div>
		</div>
	`, firstName, lastName)

	return s.emailService.SendEmail(subject, body, email)
}

func (s *AgreementService) FinalizeAgreement(email, firstName, lastName, birthDate, confirmationCode string, commercialAgreement bool) error {

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
		query = fmt.Sprintf("INSERT INTO t_system_wall_user (id, email, firstname, lastname, birthdate, isenabled, createdate, GDPR_confirmed, Rules_confirmed, commercial_confirmed) VALUES ((select max(id)+1 from t_system_wall_user), '%v', '%v', '%v', '%v', 'true', '%v', 'true', 'true', '%v')", email, firstName, lastName, utils.NormalizeDate(birthDate), time.Now().Format("2006-01-02 15:04:05"), commercialAgreement)
	} else {
		query = fmt.Sprintf("INSERT INTO t_system_wall_user (email, firstname, lastname, birthdate, isenabled, createdate, GDPR_confirmed, Rules_confirmed, commercial_confirmed) VALUES ('%v', '%v', '%v', '%v', 'true', '%v', 'true', 'true', '%v')", email, firstName, lastName, utils.NormalizeDate(birthDate), time.Now().Format("2006-01-02 15:04:05"), commercialAgreement)
	}
	_, err = s.db.Exec(query)

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
	normalizedSearch := strings.ToLower(searchQuery)
	normalizedSearch = strings.ReplaceAll(normalizedSearch, "'", "''") // Escape single quotes

	query := `
		SELECT firstname, lastname, email, createdate 
		FROM t_system_wall_user 
		WHERE isenabled = 1 
		  AND GDPR_confirmed = 1 
		  AND Rules_confirmed = 1
	`

	if searchQuery != "" {
		query += fmt.Sprintf(`
		  AND (
			LOWER(firstname) COLLATE Latin1_General_CI_AI LIKE '%%%v%%' OR
			LOWER(lastname) COLLATE Latin1_General_CI_AI LIKE '%%%v%%' OR
			LOWER(email) COLLATE Latin1_General_CI_AI LIKE '%%%v%%'
		  )
		`, normalizedSearch, normalizedSearch, normalizedSearch)
	}

	query += " ORDER BY createdate DESC"

	var visitors []models.WallVisitor
	err := s.db.Select(&visitors, query)

	if err != nil {
		return nil, err
	}

	return visitors, nil
}
