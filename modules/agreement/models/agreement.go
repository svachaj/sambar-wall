package models

import (
	baseTypes "github.com/svachaj/sambar-wall/modules/types"
)

const AGREEMENT_FORM_FIRST_NAME = "firstName"
const AGREEMENT_FORM_LAST_NAME = "lastName"
const AGREEMENT_FORM_EMAIL = "email"
const AGREEMENT_FORM_BIRTH_DATE = "birthDate"
const AGREEMENT_FORM_CONFIRMATION_CODE = "confirmationCode"
const AGREEMENT_FORM_RULES_AGREEMENT = "rulesAgreement"
const AGREEMENT_FORM_GDPR_AGREEMENT = "gdprAgreement"
const AGREEMENT_FORM_COMMERCIAL_COMMUNICATIONS = "commercialAgreement"

const AGREEMENT_FORM_STEP1 = "agreementStep1Form"
const AGREEMENT_FORM_STEP2 = "agreementStep2Form"

func AgreementFormInitModel() baseTypes.Form {
	return baseTypes.Form{
		FormFields: map[string]baseTypes.FormField{
			AGREEMENT_FORM_FIRST_NAME: {
				ID:          AGREEMENT_FORM_FIRST_NAME,
				Label:       "Jméno",
				FieldType:   "text",
				Validations: baseTypes.Validations(baseTypes.Required()), FormId: AGREEMENT_FORM_STEP2},
			AGREEMENT_FORM_LAST_NAME: {
				ID:          AGREEMENT_FORM_LAST_NAME,
				Label:       "Příjmení",
				FieldType:   "text",
				Validations: baseTypes.Validations(baseTypes.Required()), FormId: AGREEMENT_FORM_STEP2},
			AGREEMENT_FORM_EMAIL: {
				ID:          AGREEMENT_FORM_EMAIL,
				Label:       "Email",
				Disabled:    true,
				FieldType:   "text",
				Validations: baseTypes.Validations(baseTypes.Required(), baseTypes.Email()), FormId: AGREEMENT_FORM_STEP2},
			AGREEMENT_FORM_BIRTH_DATE: {
				ID:          AGREEMENT_FORM_BIRTH_DATE,
				Label:       "Datum narození",
				FieldType:   "text",
				Placeholder: "den.měsíc.rok",
				Validations: baseTypes.Validations(baseTypes.Required(), baseTypes.Date()), FormId: AGREEMENT_FORM_STEP2},
			AGREEMENT_FORM_CONFIRMATION_CODE: {
				ID:          AGREEMENT_FORM_CONFIRMATION_CODE,
				Label:       "Ověřovací kód z emailu",
				FieldType:   "number",
				Validations: baseTypes.Validations(baseTypes.Required(), baseTypes.MinLength(4)), FormId: AGREEMENT_FORM_STEP2},
			AGREEMENT_FORM_RULES_AGREEMENT: {
				ID:          AGREEMENT_FORM_RULES_AGREEMENT,
				Label:       "Souhlasím s provozním řádem stěny",
				FieldType:   "checkbox",
				Link:        "https://www.stenakladno.cz/provozni-rad",
				FormId:      AGREEMENT_FORM_STEP2,
				Validations: baseTypes.Validations(baseTypes.RequiredMsg("Musíte souhlasit s provozním řádem stěny"))},
			AGREEMENT_FORM_GDPR_AGREEMENT: {
				ID:          AGREEMENT_FORM_GDPR_AGREEMENT,
				Label:       "Souhlasím se zpracováním osobních údajů",
				FieldType:   "checkbox",
				Link:        "/static/files/gdpr.pdf",
				FormId:      AGREEMENT_FORM_STEP2,
				Validations: baseTypes.Validations(baseTypes.RequiredMsg("Musíte souhlasit se zpracováním osobních údajů"))},
			AGREEMENT_FORM_COMMERCIAL_COMMUNICATIONS: {
				ID:        AGREEMENT_FORM_COMMERCIAL_COMMUNICATIONS,
				Label:     "Chci být v obraze – novinky ze stěny i světa lezení",
				FieldType: "checkbox",
				Link:      "/static/files/informace_o_souhlasu_zasilani_obchodni_sdeleni.docx",
				FormId:    AGREEMENT_FORM_STEP2},
		},
	}
}

func AgreementFormStep1InitModel() baseTypes.Form {
	return baseTypes.Form{
		FormFields: map[string]baseTypes.FormField{
			AGREEMENT_FORM_EMAIL: {
				ID:          AGREEMENT_FORM_EMAIL,
				Label:       "Nejprve zadej svůj email, prosím",
				Placeholder: "Email",
				FieldType:   "text",
				FormId:      AGREEMENT_FORM_STEP1,
				Validations: baseTypes.Validations(baseTypes.Required(), baseTypes.Email())},
		},
	}
}

// const ERROR_LOGIN = "Chybné uživatelské jméno nebo heslo"
