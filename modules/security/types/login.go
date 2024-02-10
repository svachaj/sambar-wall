package types

import (
	baseTypes "github.com/svachaj/sambar-wall/modules/types"
)

type LoginFormResponse struct {
	baseTypes.BaseModel
	UserName baseTypes.FormField[string] `json:"username"`
	Password baseTypes.FormField[string] `json:"password"`
}

var LoginFormInitModel LoginFormResponse = LoginFormResponse{
	UserName: baseTypes.FormField[string]{Label: "Uživatelské jméno"},
	Password: baseTypes.FormField[string]{Label: "Heslo"},
}

const ERROR_LOGIN = "Chybné uživatelské jméno nebo heslo"
