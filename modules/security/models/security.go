package models

import (
	baseTypes "github.com/svachaj/sambar-wall/modules/types"
)

const LOGIN_FORM_CONFIRMATION_CODE = "confirmationCode"
const LOGIN_FORM_EMAIL = "email"

const LOGIN_FORM_STEP1 = "loginStep1Form"
const LOGIN_FORM_STEP2 = "loginStep2Form"

func SignInStep1InitModel() baseTypes.Form {
	return baseTypes.Form{
		FormFields: map[string]baseTypes.FormField{
			LOGIN_FORM_EMAIL: {
				ID:          LOGIN_FORM_EMAIL,
				Label:       "Zadej email pro přihlášení",
				Placeholder: "Email",
				FieldType:   "text",
				FormId:      LOGIN_FORM_STEP1,
				Validations: baseTypes.Validations(baseTypes.Required(), baseTypes.Email())},
		},
	}
}

func SignInStep2InitModel() baseTypes.Form {
	return baseTypes.Form{
		FormFields: map[string]baseTypes.FormField{
			LOGIN_FORM_CONFIRMATION_CODE: {
				ID:          LOGIN_FORM_CONFIRMATION_CODE,
				Label:       "Ověřovací kód z emailu",
				FieldType:   "number",
				Validations: baseTypes.Validations(baseTypes.Required()), FormId: LOGIN_FORM_STEP2},
		},
	}
}
