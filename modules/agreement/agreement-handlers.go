package agreement

import (
	"fmt"
	"strings"

	"github.com/a-h/templ"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/svachaj/sambar-wall/middlewares"
	"github.com/svachaj/sambar-wall/utils"

	agreementTemplates "github.com/svachaj/sambar-wall/modules/agreement/templates"
	toasts "github.com/svachaj/sambar-wall/modules/toasts"
)

type IAgreementHandlers interface {
	AgreementStartPage(c echo.Context) error
	CheckEmail(c echo.Context) error
	Finalize(c echo.Context) error
}

type AgreementHandlers struct {
	db *sqlx.DB
}

func NewAgreementHandlers(db *sqlx.DB) IAgreementHandlers {
	return &AgreementHandlers{db: db}
}

func (h *AgreementHandlers) AgreementStartPage(c echo.Context) error {

	isAuthenticated, _ := middlewares.IsAuthenticated(&c)

	step1Page := Step1Page(h.db, isAuthenticated)

	return utils.HTML(c, step1Page)
}

func (h *AgreementHandlers) CheckEmail(c echo.Context) error {

	email := c.FormValue("email")

	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM t_system_wall_user WHERE isenabled = 'true' AND lower(email) = '%v'", strings.ToLower(email))
	err := h.db.Get(&count, query)

	if err != nil {
		log.Error().Msgf("CheckEmail error: %v", err)
		step1WithToast := agreementTemplates.Step1Form(toasts.ErrorToast("Na serveru došlo k chybě. Zkuste akci opakovat, prosím."))
		return utils.HTML(c, step1WithToast)
	}

	// if count > 0 then user is already registered
	if count > 0 {
		step1WithToast := agreementTemplates.Step1Form(toasts.WarnToast(fmt.Sprintf("Email %v je již pro souhlas s provozním řádem na naší stěně použitý. Přejme příjemnou zábavu.", email)))
		return utils.HTML(c, step1WithToast)
	}

	step2 := agreementTemplates.Step2Form(email, toasts.InfoToast(fmt.Sprintf("Na email %v byl odeslán ověřovací kód.", email)))
	return utils.HTML(c, step2)
}

func (h *AgreementHandlers) Finalize(c echo.Context) error {

	//email := c.FormValue("email")

	step1WithToast := agreementTemplates.Step1Form(toasts.SuccessToast("Souhlas s provozním řádem byl úspěšně dokončen."))
	return utils.HTML(c, step1WithToast)

}

func Step1Page(db *sqlx.DB, isAuthenticated bool) templ.Component {
	step1Page := agreementTemplates.AgreementPage()

	return step1Page
}
