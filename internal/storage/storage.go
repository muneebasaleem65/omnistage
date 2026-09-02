package storage

type Storage interface {
	CreateUser(name string, email string, age int) (int64, error)
}