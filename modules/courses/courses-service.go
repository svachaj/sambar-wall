package courses

import (
	"github.com/jmoiron/sqlx"
	"github.com/svachaj/sambar-wall/db/types"
	"github.com/svachaj/sambar-wall/utils"
)

type ICoursesService interface {
	GetCoursesList() ([]types.CourseType, error)
	CheckApplicationFormExists(courseId int, personalId int) (bool, error)
	GetOrCreateParticipant(firstName string, lastName string, birthYear int, parentUserId int) (int, error)
	CheckCourseCapacity(courseId int) (bool, error)
	CreateApplicationForm(courseId int, participantId int, personalId int, parentName, phone, email string) (int, error)
	SendApplicationFormEmail(email string, courseId int, firstName, lastName, parentName, phone, birthYear string) error
}

type CoursesService struct {
	db           *sqlx.DB
	emailService utils.IEmailService
}

func NewCoursesService(db *sqlx.DB, emailService utils.IEmailService) ICoursesService {
	return &CoursesService{db: db, emailService: emailService}
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

func (s *CoursesService) CheckApplicationFormExists(courseId int, personalId int) (bool, error) {
	var count int
	err := s.db.Get(&count, `
	SELECT count(*) 
	FROM t_course_application_form 
	WHERE ID_course = @p1 AND PersonalIdNumber = @p2
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

func (s *CoursesService) CreateApplicationForm(courseId int, participantId int, personalId int, parentName, phone, email string) (int, error) {
	var applicationFormId int
	err := s.db.Get(&applicationFormId, `
	INSERT INTO t_course_application_form(
	ID_course, 
	ID_participant,
	PersonalIdNumber,
	ParentName,
	Phone,
	Email,
	UpdatedDate,
	CreatedDate,
	ID_UpdatedBy,
	ID_CreatedBy,
	GID,
	IsActive)
	VALUES (
	@p1,
	@p2,
	@p3,
	@p4,
	@p5,
	@p6,
	GETDATE(),
	GETDATE(),
	@p7,
	@p8,
	NEWID(),
	1
	);

	SELECT SCOPE_IDENTITY();
	`, courseId, participantId, personalId, parentName, phone, email, email, email)

	if err != nil {
		return 0, err
	}

	return applicationFormId, nil
}

func (s *CoursesService) SendApplicationFormEmail(email string, courseId int, firstName, lastName, parentName, phone, birthYear string) error {
	course := types.Course{}
	err := s.db.Get(&course, `
	SELECT 
		tct.Name1 as name, 
		tct.Description1 as description 
	FROM t_course tc 
	LEFT JOIN t_course_type tct on tc.ID_typeOfCourse = tct.ID 
	WHERE tc.ID = @p1
	`, courseId)

	if err != nil {
		return err
	}

	subject := "Přihláška na kurz " + course.Name
	body := "Dobrý den,\n\n"
	body += "Děkujeme za Vaši přihlášku na kurz " + course.Name + ".\n\n"
	body += "Níže naleznete informace o přihlášce:\n\n"
	body += "Jméno: " + firstName + " " + lastName + "\n"
	body += "Jméno rodiče: " + parentName + "\n"
	body += "Telefon: " + phone + "\n"
	body += "Rok narození: " + birthYear + "\n\n"
	body += "Těšíme se na setkání s Vámi.\n\n"
	body += "S pozdravem\n"
	body += "Tým Sambar Lezecká Stěna"

	return s.emailService.SendEmail(subject, body, email)
}
