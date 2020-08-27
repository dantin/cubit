package mysql

import (
	"log"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

// newStorageMock returns a mocked MySQL storage instance.
func newStorageMock() (*mySQLStorage, sqlmock.Sqlmock) {
	db, sqlMock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("%v", err)
	}
	return &mySQLStorage{db: db}, sqlMock
}
