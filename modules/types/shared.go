package types

type FormField struct {
	ID         string   `json:"id"`
	Label      string   `json:"label"`
	Value      string   `json:"value"`
	Errors     []string `json:"errors"`
	FieldType  string   `json:"inputType"` // text | number | date | boolean | email | password
	Validation string
}

type FormValue interface {
	string | float64
}

type BaseModel struct {
	Errors []string `json:"errors"`
	WasOk  bool     `json:"wasOk"`
}
