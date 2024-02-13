package agreement

import (
	"github.com/a-h/templ"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/svachaj/sambar-wall/middlewares"
	"github.com/svachaj/sambar-wall/utils"

	agreementTemplates "github.com/svachaj/sambar-wall/modules/agreement/templates"
)

type IAgreementHandlers interface {
	Step1(c echo.Context) error
}

type AgreementHandlers struct {
	db *sqlx.DB
}

func NewAgreementHandlers(db *sqlx.DB) IAgreementHandlers {
	return &AgreementHandlers{db: db}
}

func (h *AgreementHandlers) Step1(c echo.Context) error {

	isAuthenticated, _ := middlewares.IsAuthenticated(&c)

	step1Page := Step1Page(h.db, isAuthenticated)

	return utils.HTML(c, step1Page)
}

func Step1Page(db *sqlx.DB, isAuthenticated bool) templ.Component {
	step1Page := agreementTemplates.Step1Page()

	return step1Page
}
