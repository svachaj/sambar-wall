package types

import (
	baseTypes "github.com/svachaj/sambar-wall/modules/types"
)

type AgreementForm struct {
	baseTypes.BaseModel
	FirstName baseTypes.FormField[string] `json:"firstName"`
	LastName  baseTypes.FormField[string] `json:"lastName"`
	Email     baseTypes.FormField[string] `json:"email"`
}

var AgreementFormInitModel AgreementForm = AgreementForm{
	FirstName: baseTypes.FormField[string]{Label: "Jméno", FieldType: "text", Required: "false"},
	LastName:  baseTypes.FormField[string]{Label: "Příjmení", FieldType: "text", Required: "required"},
	Email:     baseTypes.FormField[string]{Label: "Email", FieldType: "email", Required: "required"},
}

// const ERROR_LOGIN = "Chybné uživatelské jméno nebo heslo"
