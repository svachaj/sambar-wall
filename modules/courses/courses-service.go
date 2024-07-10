package courses

import (
	"fmt"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
	"github.com/svachaj/sambar-wall/db/types"
	"github.com/svachaj/sambar-wall/utils"

	"encoding/base64"

	qrcode "github.com/skip2/go-qrcode"
)

type ICoursesService interface {
	GetCoursesList() ([]types.CourseType, error)
	CheckApplicationFormExists(courseId int, personalId string) (bool, error)
	GetOrCreateParticipant(firstName string, lastName string, birthYear int, parentUserId int) (int, error)
	CheckCourseCapacity(courseId int) (bool, error)
	CreateApplicationForm(courseId int, participantId int, personalId, parentName, phone, email string, userId int, healthState string) (int, error)
	SendApplicationFormEmail(applicationFormId int, email string, courseId int, firstName, lastName, parentName, phone, birthYear string) error
	GetApplicationsByUserId(userId int) ([]types.ApplicationForm, error)
	GetCourseInfo(id int) types.Course
	GetAllApplicationForms(searchText string) ([]types.ApplicationForm, error)
}

type CoursesService struct {
	db                  *sqlx.DB
	emailService        utils.IEmailService
	emailCopyAddress    string
	accountIBAN         string
	accountNumber       string
	generatePaymentInfo bool
}

func NewCoursesService(db *sqlx.DB, emailService utils.IEmailService, emailCopyAddress string, accountIBAN string, accountNumber string, genratePaymentInfo bool) ICoursesService {
	return &CoursesService{db: db, emailService: emailService, emailCopyAddress: emailCopyAddress, accountIBAN: accountIBAN, accountNumber: accountNumber, generatePaymentInfo: genratePaymentInfo}
}

func (s *CoursesService) GetCoursesList() ([]types.CourseType, error) {
	courses := []types.CourseType{}

	err := s.db.Select(&courses, `
	SELECT
		tct.ID as id ,
		tct.Name1 as name, 
		tct.Description1 as description 
	FROM t_course tc 
	LEFT JOIN t_course_type tct on tc.ID_typeOfCourse = tct.ID 
	WHERE tc.ValidFrom <= getdate() and tc.ValidTo >= getdate() and tc.IsActive = 1
	GROUP BY tct.Name1, tct.Description1 , tct.ID, tct.Code 
	ORDER BY tct.Code`)

	if err != nil {
		return nil, err
	}

	for i, courseType := range courses {
		coursesList := []types.Course{}
		err := s.db.Select(&coursesList, `
		SELECT 
		tc.ID as id,  			
		tc.TimeFrom as timeFrom , 
		tc.TimeTo  as timeTo,
		tcd.Name1 as days, 
		tcag.Name1 as ageGroup, 
		tc.Capacity as capacity,
		count(tcaf.id) as applicationsCount,
		tc.PartipicatnsCount as partipicatnsCount,
		tc.Price as price,
		tc.DurationMin as durationMin,
		CONCAT(tc.DurationMin , ' min, počet dětí v kurzu ', tc.PartipicatnsCount) as name
		FROM t_course tc 		
		LEFT JOIN t_course_day tcd on tc.ID_dayOfCourse = tcd.ID 
		LEFT JOIN t_course_age_group tcag on tc.ID_ageGroup = tcag.ID 
		LEFT JOIN t_course_application_form tcaf on tc.ID = tcaf.ID_course
		WHERE tc.ValidFrom <= getdate() and tc.ValidTo >= getdate() AND tc.IsActive = 1 AND tc.ID_typeOfCourse = @p1
		GROUP BY tc.ID, tc.TimeFrom ,tc.TimeTo , tcd.Name1 , tcag.Name1, tc.Capacity, tcd.Code, tc.PartipicatnsCount ,tc.Price , tc.DurationMin
		ORDER BY tcd.Code, tc.TimeFrom ;
		`, courseType.ID)

		if err != nil {
			return nil, err
		}

		courses[i].Courses = coursesList
	}

	return courses, nil
}

func (s *CoursesService) CheckApplicationFormExists(courseId int, personalId string) (bool, error) {
	var count int
	err := s.db.Get(&count, `
	SELECT count(*) 
	FROM t_course_application_form 
	WHERE ID_course = @p1 AND PersonalId = @p2
	`, courseId, personalId)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *CoursesService) GetOrCreateParticipant(firstName string, lastName string, birthYear int, parentUserId int) (int, error) {
	var participantId int
	err := s.db.Get(&participantId, `
	IF EXISTS (SELECT ID FROM t_system_user_participant WHERE FirstName = @p1 AND LastName = @p2 AND BirthYear = @p3 AND ID_ParentUser = @p4)
	BEGIN
		SELECT ID FROM t_system_user_participant WHERE FirstName = @p1 AND LastName = @p2 AND BirthYear = @p3 AND ID_ParentUser = @p4
	END
	ELSE
	BEGIN
		INSERT INTO t_system_user_participant(
		FirstName, 
		LastName,
		BirthYear, 
		ID_ParentUser,
		UpdatedDate,CreatedDate,ID_UpdatedBy,ID_CreatedBy,GID,ID_experiences,IsActive)
		VALUES (
		 @p1,
		 @p2, 
		 @p3, 
		 @p4,
		 GETDATE(),GETDATE(),@p4,@p4,NEWID(),1,1)

		SELECT SCOPE_IDENTITY()
	END
	`, firstName, lastName, birthYear, parentUserId)

	if err != nil {
		return 0, err
	}

	return participantId, nil
}

func (s *CoursesService) CheckCourseCapacity(courseId int) (bool, error) {
	var capacity int
	err := s.db.Get(&capacity, `
	SELECT tc.Capacity - count(tcaf.id) as capacity
	FROM t_course tc 
	LEFT JOIN t_course_application_form tcaf on tc.ID = tcaf.ID_course
	WHERE tc.ID = @p1
	GROUP BY tc.Capacity;
	`, courseId)

	if err != nil {
		return false, err
	}

	return capacity > 0, nil
}

func (s *CoursesService) CreateApplicationForm(courseId int, participantId int, personalId, parentName, phone, email string, userId int, healthState string) (int, error) {
	var applicationFormId int
	err := s.db.Get(&applicationFormId, `
	INSERT INTO t_course_application_form(
	ID_course, 
	ID_participant,
	PersonalId,
	ParentName,
	Phone,
	Email,
	HealthState,
	GDPR_confirmed,
	Rules_confirmed,
	UpdatedDate,
	CreatedDate,
	ID_UpdatedBy,
	ID_CreatedBy,
	GID,
	EmailSent,Paid)
	VALUES (
	@p1,
	@p2,
	@p3,
	@p4,
	@p5,
	@p6,
	@p8,
	1,
	1,
	GETDATE(),
	GETDATE(),
	@p7,
	@p7,
	NEWID(),
	0,0)
	SELECT SCOPE_IDENTITY()
	`, courseId, participantId, personalId, parentName, phone, email, userId, healthState)

	if err != nil {
		return 0, err
	}

	return applicationFormId, nil
}

func (s *CoursesService) SendApplicationFormEmail(applicationFormId int, email string, courseId int, firstName, lastName, parentName, phone, birthYear string) error {
	course := types.Course{}
	err := s.db.Get(&course, `
	SELECT 
		tct.Name1 as name, 
		tct.Description1 as description ,
		tc.Price as price,
		tc.TimeFrom as timeFrom,
		tc.TimeTo as timeTo,
		tcd.Name1 as days,
		tcag.Name1 as ageGroup
	FROM t_course tc 
	LEFT JOIN t_course_type tct on tc.ID_typeOfCourse = tct.ID 
	LEFT JOIN t_course_day tcd on tc.ID_dayOfCourse = tcd.ID
	LEFT JOIN t_course_age_group tcag on tc.ID_ageGroup = tcag.ID
	WHERE tc.ID = @p1
	`, courseId)

	if err != nil {
		return err
	}

	price := strconv.FormatFloat(course.Price, 'f', 2, 64)

	subject := "Přihláška na kurz: " + course.Name
	body := "<div style=\"width: 100%; max-width: 600px;line-heigth:1.5rem; margin: 0 auto; padding: 20px; border: 1px solid #ccc; border-radius: 10px;\">\n"
	body += "<p style=\"font-size: 20px; margin-bottom: 20px;\">Dobrý den,</p>\n\n"
	body += "<p style=\"margin-bottom: 20px;\">Děkujeme za Vaši přihlášku na kurz:<br> <strong>" + course.Name + "</strong>.</p>\n\n"
	body += "<p style=\"margin-bottom: 20px;\">Níže naleznete informace o přihlášce:</p>\n\n"
	body += "<p style=\"margin-bottom: 20px;\">\n"
	body += "<strong>Kdy:</strong> " + course.Days + "<br>\n"
	body += "<strong>V čase:</strong>  od " + course.TimeFrom.Format("15:04") + " do " + course.TimeTo.Format("15:04") + "<br>\n"
	body += "<strong>Věková skupina:</strong> " + course.AgeGroup + "<br>\n"
	body += "<strong>Jméno:</strong> " + firstName + " " + lastName + "<br>\n"
	body += "<strong>Rok narození:</strong> " + birthYear + "<br><br>\n"
	body += "<strong>Jméno rodiče:</strong> " + parentName + "<br>\n"
	body += "<strong>Telefon:</strong> " + phone + "<br>\n"
	body += "<strong>Email:</strong> " + email + "<br>\n"
	body += "<strong>Číslo přihlášky:</strong> " + strconv.Itoa(applicationFormId) + "<br><br>\n"
	body += "<strong>Cena kurzu:</strong> " + price + " Kč\n"
	body += "</p>\n\n"
	if s.generatePaymentInfo {
		// The data to encode as a QR code (e.g., payment information)
		paymentInfo := fmt.Sprintf("SPD*1.0*ACC:%v*AM:%v*CC:CZK*RF:%v*X-VS:%v*PT:IP*MSG:Platba za kurz - %v %v", s.accountIBAN, price, applicationFormId, applicationFormId, firstName+" "+lastName, birthYear)
		// Generate the QR code
		var png []byte
		png, err = qrcode.Encode(paymentInfo, qrcode.Medium, 256)
		if err != nil {
			log.Error().Err(err).Msg("Failed to generate QR code")
			return err
		}

		// Encode the QR code image to base64
		base64QRCode := base64.StdEncoding.EncodeToString(png)
		body += "<p style=\"margin-bottom: 20px;\">Pro okamžitou platbu kurzu můžete použít QR kód nebo můžete zaplatit převodem na účet:</p>\n\n"
		body += "<img src=\"data:image/png;base64," + base64QRCode + "\" style=\"margin-bottom: 20px;\"/>\n\n"
		body += "<p style=\"margin-bottom: 20px;\">Číslo účtu: " + s.accountNumber + "</p>\n\n"
		body += "<p style=\"margin-bottom: 20px;\">Variabilní symbol: " + strconv.Itoa(applicationFormId) + "</p>\n\n"
		body += "<br><br>"
	}
	body += "<p style=\"margin-bottom: 20px;\">Děkujeme Vám za přihlášení na akci pořádanou Lezeckou stěnou Kladno. Během několika pracovních dní Vám zašleme podrobné informace k akci (tábory, lezení na skalách atd.)</p>\n\n"
	body += "<p style=\"margin-bottom: 20px;\">V případě jakýchkoliv dotazů nás neváhejte kontaktovat na emailu: anna@stenakladno.cz"

	body += "<p>Často kladené otázky - <a target=\"_blank\" href=\"http://www.stenakladno.cz/clanek-1533038613-casto-kladene-otazky-cs/\">zde</a></p>"
	body += "<p>Najdete nás na adrese: Huťská, 272 01 Kladno (vjezd do areálu u svářečské školy)</p>"
	body += "<p>Nejbližší autobusová zastávka: Poldi (autobusy číslo: 7, 8, 9, 11, 12, 13, 14, 18)</p>"
	body += " <p><a href=\"https://www.google.cz/maps/place/Lezecká+stěna+Kladno/@50.150291,14.119982,17z/data=!3m1!4b1!4m2!3m1!1s0x470bc81df4294531:0xba79ce925bcfb29\" target=\"_blank\">Mapa stěny</a>"
	body += "<br>"
	body += "<a href=\"https://maps.google.cz/maps/ms?msid=215522111162202071644.0004bacd0e0ab317a4e99&msa=0&ll=50.143246,14.153481&spn=0.119914,0.338173&dg=feature\" target=\"_blank\">Mapa příjezdových tras ke stěně</a>"
	body += "</p>"
	body += "<p> GPS: 50°9'0.613\"N, 14°7'11.908\"E </p> <p> Telefon recepce: 730 827 898 </p>"
	body += "<p style=\"margin-top: 20px; font-size: 14px; color: #555;\">S pozdravem,<br>\n"
	body += "Lezecká Stěna Kladno</p>\n"
	body += "</div>\n"

	err = s.emailService.SendEmail(subject, body, email)

	if err == nil {
		// send the email also to the admin and then set the emailSent flag to true on the application form
		adminEmail := s.emailCopyAddress
		err = s.emailService.SendEmail(subject, body, adminEmail)
		if err != nil {
			log.Error().Err(err).Msg("Failed to send application form email to the admin")
		}

		_, err = s.db.Exec(`
		UPDATE t_course_application_form 
		SET EmailSent = 1
		WHERE ID = @p1
		`, applicationFormId)

		if err != nil {
			log.Error().Err(err).Msg("Failed to update application form emailSent flag")
		}
	}

	return err
}

func (s *CoursesService) GetApplicationsByUserId(userId int) ([]types.ApplicationForm, error) {
	applicationForms := []types.ApplicationForm{}

	err := s.db.Select(&applicationForms, `
	SELECT 
tcaf.ID as id, 
tcaf.Paid as paid,
tcaf.PersonalId as personalId,
tc.ID as courseId,
tct.Name1 as courseName,
tcd.Name1 as courseDays,
tc.TimeFrom as courseTimeFrom,
tc.TimeTo as courseTimeTo,
tcag.Name1 as courseAgeGroup,
tc.Price as coursePrice,
tsup.FirstName as firstName,
tsup.LastName as lastName
FROM t_course_application_form tcaf
LEFT JOIN t_course tc on tc.ID = tcaf.ID_course
LEFT join t_course_type tct on tct.ID = tc.ID_typeOfCourse
LEFT JOIN t_system_user_participant tsup on tcaf.ID_participant = tsup.ID
LEFT JOIN t_system_user tsu on tsu.ID = tsup.ID_ParentUser
LEFT JOIN t_course_day tcd on tc.ID_dayOfCourse = tcd.ID 
LEFT JOIN t_course_age_group tcag on tc.ID_ageGroup = tcag.ID 
WHERE tsu.ID = @p1
ORDER BY tcaf.CreatedDate DESC;
	`, userId)

	if err != nil {
		return nil, err
	}

	return applicationForms, nil
}

func (s *CoursesService) GetCourseInfo(id int) types.Course {
	course := types.Course{}

	err := s.db.Get(&course, `
	SELECT TOP 1
tc.ID as id, 
tct.Name1 as name,
tct.Description1 as description,
tcd.Name1 as days,
tc.TimeFrom as timeFrom,
tc.TimeTo as timeTo,
tcag.Name1 as ageGroup,
tc.Price as price
FROM t_course_application_form tcaf
LEFT JOIN t_course tc on tc.ID = tcaf.ID_course
LEFT join t_course_type tct on tct.ID = tc.ID_typeOfCourse
LEFT JOIN t_system_user_participant tsup on tcaf.ID_participant = tsup.ID
LEFT JOIN t_system_user tsu on tsu.ID = tsup.ID_ParentUser
LEFT JOIN t_course_day tcd on tc.ID_dayOfCourse = tcd.ID 
LEFT JOIN t_course_age_group tcag on tc.ID_ageGroup = tcag.ID 
WHERE tc.ID = @p1
GROUP by tc.ID, tct.ID, tct.Name1, tct.Description1, tc.TimeFrom,tc.TimeTo, tcag.Name1, tc.Price, tcd.Name1;
	`, id)

	if err != nil {
		return types.Course{}
	}

	return course
}

func (s *CoursesService) GetAllApplicationForms(searchText string) ([]types.ApplicationForm, error) {
	applicationForms := []types.ApplicationForm{}

	err := s.db.Select(&applicationForms, `
	SELECT
tcaf.ID as id,
tcaf.Paid as paid,
tcaf.PersonalId as personalId,
tcaf.HealthState as healthState,
tsup.BirthYear as birthYear,
tc.ID as courseId,
tct.Name1 as courseName,
tcd.Name1 as courseDays,
tc.TimeFrom as courseTimeFrom,
tc.TimeTo as courseTimeTo,
tcag.Name1 as courseAgeGroup,
tc.Price as coursePrice,
tsup.FirstName as firstName,
tsup.LastName as lastName,
tsu.Email as email,
tcaf.CreatedDate as createdDate
FROM t_course_application_form tcaf
LEFT JOIN t_course tc on tc.ID = tcaf.ID_course
LEFT join t_course_type tct on tct.ID = tc.ID_typeOfCourse
LEFT JOIN t_system_user_participant tsup on tcaf.ID_participant = tsup.ID
LEFT JOIN t_system_user tsu on tsu.ID = tsup.ID_ParentUser
LEFT JOIN t_course_day tcd on tc.ID_dayOfCourse = tcd.ID
LEFT JOIN t_course_age_group tcag on tc.ID_ageGroup = tcag.ID
WHERE NULL is NULL OR (tsup.FirstName LIKE @p1 OR tsup.LastName LIKE @p1 OR tcaf.PersonalId LIKE @p1)
ORDER BY tcaf.CreatedDate DESC;
	`, "%"+searchText+"%")

	if err != nil {
		return nil, err
	}

	return applicationForms, nil
}
