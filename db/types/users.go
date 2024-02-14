package types

import "time"

type User struct {
	ID           int    `json:"id" db:"id"`
	UserName     string `json:"username" db:"username"`
	PasswordHash string `json:"-" db:"passwordhash"`
}

type RegistrationCode struct {
	ID         int       `json:"id" db:"id"`
	Code       string    `json:"code" db:"code"`
	CreateDate time.Time `json:"createdate" db:"createdate"`
	Email      string    `json:"email" db:"email"`
}
