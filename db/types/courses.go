package types

import "time"

type Course struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	ValidFrom time.Time `json:"valid_from" db:"valid_from"`
	ValidTo   time.Time `json:"valid_to" db:"valid_to"`
}
