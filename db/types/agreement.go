package types

import "time"

type Agreement struct {
	ID                  int       `json:"id" db:"id"`
	BirthDate           time.Time `json:"birthdate" db:"birthdate"`
	CreateDate          time.Time `json:"createdate" db:"createdate"`
	Email               string    `json:"email" db:"email"`
	FirstName           string    `json:"firstname" db:"firstname"`
	LastName            string    `json:"lastname" db:"lastname"`
	GdprConfirmed       bool      `json:"gdprconfirmed" db:"gdpr_confirmed"`
	RulesConfirmed      bool      `json:"rulesconfirmed" db:"rules_confirmed"`
	CommercialConfirmed bool      `json:"commercialConfirmed" db:"commercial_confirmed"`
	IsEnabled           bool      `json:"isenabled" db:"isenabled"`
}
