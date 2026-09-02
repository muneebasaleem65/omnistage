package storage

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
)

type Storage interface {
	CreateUser(email, passwordHash, name string) (int64, error)
}
