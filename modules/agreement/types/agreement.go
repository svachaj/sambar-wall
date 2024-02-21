package types

import (
	baseTypes "github.com/svachaj/sambar-wall/modules/types"
)

type AgreementForm struct {
	baseTypes.BaseModel
	FirstName        baseTypes.FormField `json:"firstName"`
	LastName         baseTypes.FormField `json:"lastName"`
	Email            baseTypes.FormField `json:"email"`
	BirthDate        baseTypes.FormField `json:"birthDate"`
	ConfirmationCode baseTypes.FormField `json:"confirmationCode"`
}

var AgreementFormInitModel AgreementForm = AgreementForm{
	FirstName:        baseTypes.FormField{ID: "firstName", Label: "Jméno", FieldType: "text"},
	LastName:         baseTypes.FormField{ID: "lastName", Label: "Příjmení", FieldType: "text", Validation: "required"},
	Email:            baseTypes.FormField{ID: "email", Label: "Email", FieldType: "email", Validation: "required"},
	BirthDate:        baseTypes.FormField{ID: "birthDate", Label: "Datum narození", FieldType: "date", Validation: "required"},
	ConfirmationCode: baseTypes.FormField{ID: "confirmationCode", Label: "Ověřovací kód z emailu", FieldType: "text", Validation: "required"},
}

type AgreementFormStep1 struct {
	baseTypes.BaseModel
	Email baseTypes.FormField `json:"email"`
}

var AgreementFormStep1InitModel AgreementFormStep1 = AgreementFormStep1{
	Email: baseTypes.FormField{ID: "email", Label: "Nejprve zadejte váš email, prosím", FieldType: "email", Validation: "required"},
}

// const ERROR_LOGIN = "Chybné uživatelské jméno nebo heslo"
