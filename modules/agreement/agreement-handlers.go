package agreement

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
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
	service IAgreementService
}

func NewAgreementHandlers(svc IAgreementService) IAgreementHandlers {
	return &AgreementHandlers{service: svc}
}

func (h *AgreementHandlers) AgreementStartPage(c echo.Context) error {

	step1Page := Step1Page()

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

	// check if email exists
	existEmail, err := h.service.EmailExists(email)
	if err != nil {
		log.Error().Msgf("CheckEmail error: %v", err)
		step1WithToast := agreementTemplates.Step1Form(types.AgreementFormStep1InitModel, toasts.ServerErrorToast())
		return utils.HTML(c, step1WithToast)
	}

	if existEmail {
		step1WithToast := agreementTemplates.Step1Form(types.AgreementFormStep1InitModel, toasts.WarnToast(fmt.Sprintf("Email %v je již pro souhlas s provozním řádem na naší stěně použitý. Přejme příjemnou zábavu.", email)))
		return utils.HTML(c, step1WithToast)
	}

	// generate and save verification code
	code := h.service.GenerateVerificationCode()
	err = h.service.SaveVerificationCode(email, code)
	if err != nil {
		log.Error().Msgf("Save verification code error: %v", err)
		step1WithToast := agreementTemplates.Step1Form(types.AgreementFormStep1InitModel, toasts.ServerErrorToast())
		return utils.HTML(c, step1WithToast)
	}

	// send verification code
	err = h.service.SendVerificationCode(email, code)
	if err != nil {
		log.Error().Msgf("Send verification code error: %v", err)
		step1WithToast := agreementTemplates.Step1Form(types.AgreementFormStep1InitModel, toasts.ServerErrorToast())
		return utils.HTML(c, step1WithToast)
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

func Step1Page() templ.Component {
	step1Page := agreementTemplates.AgreementPage()

	return step1Page
}
