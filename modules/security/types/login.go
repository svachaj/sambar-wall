package types

type LoginFormModel struct {
	UserName FormField[string] `json:"username"`
	Password FormField[string] `json:"password"`
}

type FormField[V FormValue] struct {
	Value   V        `json:"value"`
	IsValid bool     `json:"isValid"`
	Errors  []string `json:"errors"`
}

type FormValue interface {
	string | float64
}
