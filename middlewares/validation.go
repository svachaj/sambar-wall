package middlewares

import (
	"net/url"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	formComponents "github.com/svachaj/sambar-wall/modules/components/forms"
	toasts "github.com/svachaj/sambar-wall/modules/toasts"
	"github.com/svachaj/sambar-wall/modules/types"
	"github.com/svachaj/sambar-wall/utils"
)

// ServerHeader middleware adds a `Server` header to the response.
func ValidateFormField(c echo.Context) error {

	var body = make(map[string]interface{})

	err := c.Bind(&body)
	if err != nil {
		log.Error().Err(err).Msg("Error binding request body")
		errToast := toasts.ErrorToast("Něco se pokazilo, zkuste to prosím znovu.")
		return utils.HTMLWithStatus(c, 500, errToast)
	} else {
		fieldName := c.Request().Header.Get("HX-Trigger-Name")
		fieldValue := body[fieldName].(string)
		fieldValidation := c.Request().Header.Get("Field-Validation")
		fieldType := c.Request().Header.Get("Field-Type")
		fieldLabel, _ := url.QueryUnescape(c.Request().Header.Get("Field-Label"))

		formField := types.FormField{ID: fieldName, Label: fieldLabel, FieldType: fieldType, Value: fieldValue}

		if fieldValidation == "required" && fieldValue == "" {
			formField.Errors = []string{"Toto pole je povinné!"}

		}
		return utils.HTML(c, formComponents.FormField(formField))

	}
}
