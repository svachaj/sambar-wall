package types

type User struct {
	ID           int    `json:"id" db:"id"`
	UserName     string `json:"username" db:"username"`
	PasswordHash string `json:"-" db:"password_hash"`
}
