package storage

import (
	"errors"

	"github.com/muneebasaleem65/omnistage/internal/types"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
)

type Storage interface {
	CreateUser(email, passwordHash, name string) (int64, error)
	GetUserByEmail(email string) (types.User, error)
}
