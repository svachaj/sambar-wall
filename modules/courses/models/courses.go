package models

import (
	baseTypes "github.com/svachaj/sambar-wall/modules/types"
)

const APPLICATION_FORM = "applicationForm"

const APPLICATION_FORM_COURSE_ID = "courseId"
const APPLICATION_FORM_FIRST_NAME = "firstName"
const APPLICATION_FORM_LAST_NAME = "lastName"
const APPLICATION_FORM_PERSONAL_ID = "personalId"
const APPLICATION_FORM_PHONE = "phone"
const APPLICATION_FORM_PARENT_NAME = "parentName"
const APPLICATION_FORM_GDPR = "gdpr"
const APPLICATION_FORM_RULES = "rules"

func ApplicationFormModel(courseId string) baseTypes.Form {
	return baseTypes.Form{
		FormFields: map[string]baseTypes.FormField{
			APPLICATION_FORM_COURSE_ID: {
				ID:        APPLICATION_FORM_COURSE_ID,
				Label:     "ID kurzu",
				FieldType: "hidden",
				FormId:    APPLICATION_FORM,
				Value:     courseId},
			APPLICATION_FORM_FIRST_NAME: {
				ID:          APPLICATION_FORM_FIRST_NAME,
				Label:       "Jméno (koho přihlašuji)",
				FieldType:   "text",
				Validations: baseTypes.Validations(baseTypes.Required()),
				FormId:      APPLICATION_FORM},
			APPLICATION_FORM_LAST_NAME: {
				ID:          APPLICATION_FORM_LAST_NAME,
				Label:       "Příjmení (koho přihlašuji)",
				FieldType:   "text",
				FormId:      APPLICATION_FORM,
				Validations: baseTypes.Validations(baseTypes.Required())},
			APPLICATION_FORM_PERSONAL_ID: {
				ID:          APPLICATION_FORM_PERSONAL_ID,
				Label:       "Rodné číslo (koho přihlašuji)",
				Placeholder: "10 čísel bez lomítka (9 pro 1953 a starší)",
				FieldType:   "number",
				FormId:      APPLICATION_FORM,
				Validations: baseTypes.Validations(baseTypes.Required(), baseTypes.MinLength(9), baseTypes.MaxLength(10))},
			APPLICATION_FORM_PHONE: {
				ID:          APPLICATION_FORM_PHONE,
				Label:       "Telefonní číslo zákonného zástupce",
				FieldType:   "text",
				FormId:      APPLICATION_FORM,
				Validations: baseTypes.Validations(baseTypes.Required())},
			APPLICATION_FORM_PARENT_NAME: {
				ID:          APPLICATION_FORM_PARENT_NAME,
				Label:       "Jméno a příjmení zákonného zástupce",
				FieldType:   "text",
				FormId:      APPLICATION_FORM,
				Validations: baseTypes.Validations(baseTypes.Required())},
			APPLICATION_FORM_GDPR: {
				ID:          APPLICATION_FORM_GDPR,
				Label:       "Souhlasím se zpracováním osobních údajů",
				FieldType:   "checkbox",
				Link:        "/static/files/gdpr.pdf",
				FormId:      APPLICATION_FORM,
				Validations: baseTypes.Validations(baseTypes.Required())},
			APPLICATION_FORM_RULES: {
				ID:          APPLICATION_FORM_RULES,
				Label:       "Souhlasím s provozním řádem stěny",
				FieldType:   "checkbox",
				Link:        "https://www.stenakladno.cz/provozni-rad",
				FormId:      APPLICATION_FORM,
				Validations: baseTypes.Validations(baseTypes.Required())},
		},
	}
}
