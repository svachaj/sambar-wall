package types

type FormField[V FormValue] struct {
	Label  string   `json:"label"`
	Value  V        `json:"value"`
	Errors []string `json:"errors"`
}

type FormValue interface {
	string | float64
}

type BaseModel struct {
	Errors []string `json:"errors"`
}
