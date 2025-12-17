package models

import "time"

type WallVisitor struct {
	FirstName  string    `json:"firstname" db:"firstname"`
	LastName   string    `json:"lastname" db:"lastname"`
	Email      string    `json:"email" db:"email"`
	CreateDate time.Time `json:"createdate" db:"createdate"`
}
