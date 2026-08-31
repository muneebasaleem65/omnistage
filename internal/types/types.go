package types

type User struct {
	ID    int
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
	Age   int    `validate:"required"`
}
