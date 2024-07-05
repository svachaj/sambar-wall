package types

import (
	baseTypes "github.com/svachaj/sambar-wall/modules/types"
)

type LoginFormResponse struct {
	baseTypes.Form
	UserName baseTypes.FormField `json:"username"`
	Password baseTypes.FormField `json:"password"`
}

var LoginFormInitModel LoginFormResponse = LoginFormResponse{
	UserName: baseTypes.FormField{Label: "Uživatelské jméno"},
	Password: baseTypes.FormField{Label: "Heslo"},
}

const ERROR_LOGIN = "Neplatný kód"
