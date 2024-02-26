package models

import (
	baseTypes "github.com/svachaj/sambar-wall/modules/types"
)

func AgreementFormInitModel() baseTypes.Form {
	return baseTypes.Form{
		FormFields: map[string]baseTypes.FormField{
			"firstName":        {ID: "firstName", Label: "Jméno", FieldType: "text", Validations: baseTypes.Validations(baseTypes.Required())},
			"lastName":         {ID: "lastName", Label: "Příjmení", FieldType: "text", Validations: baseTypes.Validations(baseTypes.Required())},
			"email":            {ID: "email", Label: "Email", FieldType: "text", Validations: baseTypes.Validations(baseTypes.Required(), baseTypes.Email())},
			"birthDate":        {ID: "birthDate", Label: "Datum narození", FieldType: "date", Validations: baseTypes.Validations(baseTypes.Required())},
			"confirmationCode": {ID: "confirmationCode", Label: "Ověřovací kód z emailu", FieldType: "text", Validations: baseTypes.Validations(baseTypes.Required())},
		},
	}
}

func AgreementFormStep1InitModel() baseTypes.Form {
	return baseTypes.Form{
		FormFields: map[string]baseTypes.FormField{
			"email": {ID: "email",
				Label:       "Nejprve zadejte svůj email, prosím",
				FieldType:   "text",
				Validations: baseTypes.Validations(baseTypes.Required(), baseTypes.Email())},
		},
	}
}

// const ERROR_LOGIN = "Chybné uživatelské jméno nebo heslo"
