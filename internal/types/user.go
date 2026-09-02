package types

type User struct {
	ID           int64
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=8,max=72"`
	PasswordHash string `json:"-"`
	Name         string `json:"name" validate:"required"`
}
