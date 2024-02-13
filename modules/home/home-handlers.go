package home

import (
	"github.com/a-h/templ"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/svachaj/sambar-wall/db/types"
	"github.com/svachaj/sambar-wall/middlewares"
	"github.com/svachaj/sambar-wall/utils"

	homeTemplates "github.com/svachaj/sambar-wall/modules/home/templates"
)

type IHomeHandlers interface {
	Home(c echo.Context) error
}

type HomeHandlers struct {
	db *sqlx.DB
}

func NewHomeHandlers(db *sqlx.DB) IHomeHandlers {
	return &HomeHandlers{db: db}
}

func (h *HomeHandlers) Home(c echo.Context) error {

	isAuthenticated, _ := middlewares.IsAuthenticated(&c)

	homePage := HomePage(h.db, isAuthenticated)

	return utils.HTML(c, homePage)
}

func HomePage(db *sqlx.DB, isAuthenticated bool) templ.Component {
	courses := []types.Course{}
	err := db.Select(&courses, `SELECT tc.id as id, tct.Name1 as name , tc.ValidFrom as valid_from, tc.ValidTo as valid_to FROM t_course tc
	inner join t_course_type tct on tc.ID_typeOfCourse = tct.ID
	ORDER BY tc.ValidTo desc, tct.Name1 `)
	if err != nil {
		log.Error().Err(err).Msg("Error getting courses")
		courses = []types.Course{}
		courses = append(courses, types.Course{Name: "Chyba při načítání kurzů"})
	}

	homeComponent := homeTemplates.HomeComponent(courses)
	homePage := homeTemplates.HomePage(homeComponent, isAuthenticated)

	return homePage
}
