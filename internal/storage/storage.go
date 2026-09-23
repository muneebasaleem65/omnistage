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
	GetUserByID(id int64) (types.User, error)
	CreateVenue(name, city string, capacity int) (int64, error)
	//ListVenues(no parameters as we need to fetch everything) (a slice of venues as it returns a collection)
	ListVenues() ([]types.Venue, error)
}
