package home

import (
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/svachaj/sambar-wall/db/types"
	"github.com/svachaj/sambar-wall/utils"

	"github.com/svachaj/sambar-wall/modules/constants"
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

	courses := []types.Course{}
	err := h.db.Select(&courses, `SELECT tc.id as id, tct.Name1 as name , tc.ValidFrom as valid_from, tc.ValidTo as valid_to FROM t_course tc
	inner join t_course_type tct on tc.ID_typeOfCourse = tct.ID
	ORDER BY tc.ValidTo desc, tct.Name1 `)
	if err != nil {
		log.Error().Err(err).Msg("Error getting courses")
		courses = []types.Course{}
		courses = append(courses, types.Course{Name: "Chyba při načítání kurzů"})
	}

	authSession, _ := session.Get(constants.AUTH_SESSION_NAME, c)

	if authSession != nil {

	}

	homeComponent := homeTemplates.HomeComponent(courses)
	homePage := homeTemplates.HomePage(homeComponent, false)

	return utils.HTML(c, homePage)
}
