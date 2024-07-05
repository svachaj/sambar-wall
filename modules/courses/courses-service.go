package courses

import (
	"github.com/jmoiron/sqlx"
	"github.com/svachaj/sambar-wall/db/types"
	"github.com/svachaj/sambar-wall/utils"
)

type ICoursesService interface {
	GetCoursesList() ([]types.CourseType, error)
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
