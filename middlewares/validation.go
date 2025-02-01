package middlewares

import (
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	dbTypes "github.com/svachaj/sambar-wall/db/types"
	agreementModels "github.com/svachaj/sambar-wall/modules/agreement/models"
	formComponents "github.com/svachaj/sambar-wall/modules/components/forms"
	coursesModels "github.com/svachaj/sambar-wall/modules/courses/models"
	securityModels "github.com/svachaj/sambar-wall/modules/security/models"
	toasts "github.com/svachaj/sambar-wall/modules/toasts"
	"github.com/svachaj/sambar-wall/modules/types"
	"github.com/svachaj/sambar-wall/utils"
)

// ValidateFormField validates a form field based on the provided rules and returns the validation result.
func ValidateFormField(c echo.Context) error {

	var body = make(map[string]interface{})

	err := c.Bind(&body)
	if err != nil {
		log.Error().Err(err).Msg("Error binding request body")
		errToast := toasts.ErrorToast("Něco se pokazilo, zkuste to prosím znovu.")
		return utils.HTMLWithStatus(c, 500, errToast)
	} else {
		fieldName := c.Request().Header.Get("HX-Trigger-Name")
		fieldValue := ""
		fieldVal := body[fieldName]
		if fieldVal != nil {
			fieldValue = fieldVal.([]string)[0]
		}

		formId := c.Request().Header.Get("Form-Id")

		form := Forms[formId]

		formField := form.FormFields[fieldName]
		formField.Value = fieldValue

		for _, rule := range formField.Validations {
			if !rule.ValidateFunc(fieldValue) {
				formField.Errors = append(formField.Errors, rule.MessageFunc())
			}
		}

		return utils.HTML(c, formComponents.FormField(formField))

	}
}

// Forms is a map of form IDs to their corresponding form models.
var Forms map[string]types.Form = map[string]types.Form{
	agreementModels.AGREEMENT_FORM_STEP1: agreementModels.AgreementFormStep1InitModel(),
	agreementModels.AGREEMENT_FORM_STEP2: agreementModels.AgreementFormInitModel(),
	securityModels.LOGIN_FORM_STEP1:      securityModels.SignInStep1InitModel(),
	securityModels.LOGIN_FORM_STEP2:      securityModels.SignInStep2InitModel(),
	coursesModels.APPLICATION_FORM:       coursesModels.ApplicationFormModel("0"),
	coursesModels.APPLICATION_FORM_EDIT:  coursesModels.ApplicationFormEditModel(dbTypes.ApplicationForm{}),
}
