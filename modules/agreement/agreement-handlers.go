package agreement

import (
	"fmt"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"github.com/svachaj/sambar-wall/utils"

	"github.com/svachaj/sambar-wall/modules/agreement/models"
	agreementTemplates "github.com/svachaj/sambar-wall/modules/agreement/templates"
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

	// validate form
	step1Form := models.AgreementFormStep1InitModel()
	params, _ := c.FormParams()
	isValid := step1Form.ValidateFields(params)

	if !isValid {
		step1 := agreementTemplates.Step1Form(step1Form, nil)
		return utils.HTML(c, step1)
	}

	email := step1Form.FormFields[models.AGREEMENT_FORM_EMAIL].Value

	// check if email exists
	existEmail, err := h.service.EmailExists(email)
	if err != nil {
		log.Error().Msgf("CheckEmail error: %v", err)
		step1WithToast := agreementTemplates.Step1Form(step1Form, toasts.ServerErrorToast())
		return utils.HTML(c, step1WithToast)
	}

	if existEmail {
		step1WithToast := agreementTemplates.Step1Form(step1Form, toasts.WarnToast(fmt.Sprintf("Email %v je již pro souhlas s provozním řádem na naší stěně použitý. Přejme příjemnou zábavu.", email)))
		return utils.HTML(c, step1WithToast)
	}

	// generate and save verification code
	code := h.service.GenerateVerificationCode()
	err = h.service.SaveVerificationCode(email, code)
	if err != nil {
		log.Error().Msgf("Save verification code error: %v", err)
		step1WithToast := agreementTemplates.Step1Form(step1Form, toasts.ServerErrorToast())
		return utils.HTML(c, step1WithToast)
	}

	// send verification code
	err = h.service.SendVerificationCode(email, code)
	if err != nil {
		log.Error().Msgf("Send verification code error: %v", err)
		step1WithToast := agreementTemplates.Step1Form(step1Form, toasts.ServerErrorToast())
		return utils.HTML(c, step1WithToast)
	}

	agreementForm := models.AgreementFormInitModel()
	if val, ok := agreementForm.FormFields[models.AGREEMENT_FORM_EMAIL]; ok {
		val.Value = email
		agreementForm.FormFields[models.AGREEMENT_FORM_EMAIL] = val
	}
	step2 := agreementTemplates.Step2Form(agreementForm, toasts.InfoToast(fmt.Sprintf("Na zadaný email %v byl odeslán ověřovací kód.", email)))
	return utils.HTML(c, step2)
}

func (h *AgreementHandlers) Finalize(c echo.Context) error {

	// validate form
	agreementForm := models.AgreementFormInitModel()
	params, _ := c.FormParams()

	isValid := agreementForm.ValidateFields(params)

	if !isValid {
		step2 := agreementTemplates.Step2Form(agreementForm, nil)
		return utils.HTML(c, step2)
	}

	step1WithToast := agreementTemplates.Step1Form(models.AgreementFormStep1InitModel(), toasts.SuccessToast("Souhlas s provozním řádem byl úspěšně dokončen."))
	return utils.HTML(c, step1WithToast)

}

func Step1Page() templ.Component {
	step1Page := agreementTemplates.AgreementPage()

	return step1Page
}
