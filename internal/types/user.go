package types

type User struct {
	ID           int64
	Email        string
	PasswordHash string
	Role         string
	Name		 string
}
