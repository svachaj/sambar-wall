package agreement

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/a-h/templ"
	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/svachaj/sambar-wall/middlewares"
	"github.com/svachaj/sambar-wall/utils"

	agreementTemplates "github.com/svachaj/sambar-wall/modules/agreement/templates"
	"github.com/svachaj/sambar-wall/modules/agreement/types"
	toasts "github.com/svachaj/sambar-wall/modules/toasts"
)

type IAgreementHandlers interface {
	AgreementStartPage(c echo.Context) error
	CheckEmail(c echo.Context) error
	Finalize(c echo.Context) error
}

type AgreementHandlers struct {
	db           *sqlx.DB
	emailService *utils.EmailService
}

func NewAgreementHandlers(db *sqlx.DB, emailService *utils.EmailService) IAgreementHandlers {
	return &AgreementHandlers{db: db, emailService: emailService}
}

func (h *AgreementHandlers) AgreementStartPage(c echo.Context) error {

	isAuthenticated, _ := middlewares.IsAuthenticated(&c)

	step1Page := Step1Page(h.db, isAuthenticated)

	return utils.HTML(c, step1Page)
}

func (h *AgreementHandlers) CheckEmail(c echo.Context) error {

	email := c.FormValue("email")

	// validate email
	if email == "" {
		model := types.AgreementFormStep1InitModel
		model.Email.Errors = append(model.Email.Errors, "Email je povinný údaj.")
		step1WithToast := agreementTemplates.Step1Form(model, nil)
		return utils.HTML(c, step1WithToast)
	}

	var count int
	query := fmt.Sprintf("SELECT COUNT(*) FROM t_system_wall_user WHERE isenabled = 'true' AND lower(email) = '%v'", strings.ToLower(email))
	err := h.db.Get(&count, query)

	if err != nil {
		log.Error().Msgf("CheckEmail error: %v", err)
		step1WithToast := agreementTemplates.Step1Form(types.AgreementFormStep1InitModel, toasts.ServerErrorToast())
		return utils.HTML(c, step1WithToast)
	}

	// if count > 0 then user is already registered
	if count > 0 {
		step1WithToast := agreementTemplates.Step1Form(types.AgreementFormStep1InitModel, toasts.WarnToast(fmt.Sprintf("Email %v je již pro souhlas s provozním řádem na naší stěně použitý. Přejme příjemnou zábavu.", email)))
		return utils.HTML(c, step1WithToast)
	}

	if email != "" {
		// send email with verification code
		// generate verification code, random 4 digit number
		code := rand.Int31n(10000)

		err := h.emailService.SendEmail("Ověření emailu pro souhlas s provozním řádem", fmt.Sprintf("Ověřovací kód: %v", code), email)
		if err != nil {
			log.Error().Msgf("CheckEmail error: %v", err)
			step1WithToast := agreementTemplates.Step1Form(types.AgreementFormStep1InitModel, toasts.ServerErrorToast())
			return utils.HTML(c, step1WithToast)
		}
	}

	agreementForm := types.AgreementFormInitModel
	agreementForm.Email.Value = email
	step2 := agreementTemplates.Step2Form(agreementForm, toasts.InfoToast(fmt.Sprintf("Na zadaný email %v byl odeslán ověřovací kód.", email)))
	return utils.HTML(c, step2)
}

func (h *AgreementHandlers) Finalize(c echo.Context) error {

	//email := c.FormValue("email")

	step1WithToast := agreementTemplates.Step1Form(types.AgreementFormStep1InitModel, toasts.SuccessToast("Souhlas s provozním řádem byl úspěšně dokončen."))
	return utils.HTML(c, step1WithToast)

}

func Step1Page(db *sqlx.DB, isAuthenticated bool) templ.Component {
	step1Page := agreementTemplates.AgreementPage()

	return step1Page
}
