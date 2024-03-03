package types

import (
	"regexp"
	"strings"
)

type FormField struct {
	ID          string
	Label       string
	Placeholder string
	Value       string
	Errors      []string
	FieldType   string
	Validations []ValidationRule
	Link        string
	Disabled    bool
	FormId      string
	Pattern     string
}
type IForm interface {
	ValidateFields(data map[string][]string) bool
}
type Form struct {
	FormFields map[string]FormField
	Errors     []string
}

func NewForm(formFields map[string]FormField, errors []string) IForm {
	return Form{FormFields: formFields, Errors: errors}
}

func (model Form) ValidateFields(data map[string][]string) bool {
	isValid := true
	for k, v := range model.FormFields {
		if val, ok := data[k]; ok {
			v.Value = val[0]
			for _, rule := range v.Validations {
				if !rule.ValidateFunc(v.Value) {
					v.Errors = append(v.Errors, rule.MessageFunc())
					isValid = false
				}
			}
			model.FormFields[k] = v
		} else if v.FieldType == "checkbox" {
			v.Value = ""
			for _, rule := range v.Validations {
				if !rule.ValidateFunc(v.Value) {
					v.Errors = append(v.Errors, rule.MessageFunc())
					isValid = false
				}
			}
			model.FormFields[k] = v
		}
	}
	return isValid
}

type ValidationRule struct {
	MessageFunc  func() string
	ValidateFunc func(value string) bool
}

type ValidationFunc func() ValidationRule

func Required() ValidationFunc {
	return func() ValidationRule {
		return ValidationRule{
			MessageFunc: func() string {
				return "Povinné pole"
			},
			ValidateFunc: func(value string) bool {
				return value != ""
			},
		}
	}
}

func RequiredMsg(customMessage string) ValidationFunc {
	return func() ValidationRule {
		return ValidationRule{
			MessageFunc: func() string {
				return customMessage
			},
			ValidateFunc: func(value string) bool {
				return value != ""
			},
		}
	}
}

func Email() ValidationFunc {
	return func() ValidationRule {
		return ValidationRule{
			MessageFunc: func() string {
				return "Neplatný email"
			},
			ValidateFunc: func(value string) bool {
				return emailRegex.MatchString(value)
			},
		}
	}
}

func Date() ValidationFunc {
	return func() ValidationRule {
		return ValidationRule{
			MessageFunc: func() string {
				return "Neplatný datum - akceptovaný formát je: den.měsíc.rok"
			},
			ValidateFunc: func(value string) bool {
				dateParts := strings.Split(value, ".")
				return len(dateParts) == 3
			},
		}
	}
}

func Validations(validations ...ValidationFunc) []ValidationRule {
	rules := make([]ValidationRule, len(validations))
	for i, v := range validations {
		rules[i] = v()
	}
	return rules
}

var (
	emailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	urlRegex   = regexp.MustCompile(`^(http(s)?://)?([\da-z\.-]+)\.([a-z\.]{2,6})([/\w \.-]*)*/?$`)
)
