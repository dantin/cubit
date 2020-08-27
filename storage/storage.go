package storage

import (
	"fmt"

	memorystorage "github.com/dantin/cubit/storage/memory"
	"github.com/dantin/cubit/storage/mysql"
	"github.com/dantin/cubit/storage/repository"
)

// New initializes configured storage type and returns associated container.
func New(config *Config) (repository.Container, error) {
	switch config.Type {
	case MySQL:
		return mysql.New(config.MySQL)
	case Memory:
		return memorystorage.New()
	default:
		return nil, fmt.Errorf("storage: unrecognized storage type: %d", config.Type)
	}
}
