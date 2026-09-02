package user

type User struct {
	ID    int64
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
	Age   int    `validate:"required"`
}
