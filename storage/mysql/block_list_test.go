package mysql

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/dantin/cubit/model"
	"github.com/stretchr/testify/require"
)

func newBlockListMock() (*mySQLBlockList, sqlmock.Sqlmock) {
	s, sqlMock := newStorageMock()
	return &mySQLBlockList{
		mySQLStorage: s,
	}, sqlMock
}

func TestMySQLStorage_InsertBlockListItems(t *testing.T) {
	s, mock := newBlockListMock()
	mock.ExpectExec("INSERT IGNORE INTO blocklist_items (.+)").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := s.InsertBlockListItem(context.Background(), &model.BlockListItem{Username: "username", JID: "demo@example.org"})
	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)

	s, mock = newBlockListMock()
	mock.ExpectExec("INSERT IGNORE INTO blocklist_items (.+)").WillReturnError(errMySQLStorage)

	err = s.InsertBlockListItem(context.Background(), &model.BlockListItem{Username: "username", JID: "demo@example.org"})
	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)

}

func TestMySQLStorage_FetchBlockListItems(t *testing.T) {
	var blocklistColumns = []string{"username", "jid"}
	s, mock := newBlockListMock()
	mock.ExpectQuery("SELECT (.+) FROM blocklist_items (.*)").
		WithArgs("demo").
		WillReturnRows(sqlmock.NewRows(blocklistColumns).AddRow("demo", "demo@example.org"))

	_, err := s.FetchBlockListItems(context.Background(), "demo")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)

	s, mock = newBlockListMock()
	mock.ExpectQuery("SELECT (.+) FROM blocklist_items (.*)").
		WithArgs("demo").
		WillReturnError(errMySQLStorage)
	_, err = s.FetchBlockListItems(context.Background(), "demo")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)
}

func TestMySQLStorage_DeleteBlockListItems(t *testing.T) {
	s, mock := newBlockListMock()
	mock.ExpectExec("DELETE FROM blocklist_items (.*)").
		WithArgs("demo", "demo@example.org").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := s.DeleteBlockListItem(context.Background(), &model.BlockListItem{Username: "demo", JID: "demo@example.org"})
	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)

	s, mock = newBlockListMock()
	mock.ExpectExec("DELETE FROM blocklist_items (.*)").
		WillReturnError(errMySQLStorage)

	err = s.DeleteBlockListItem(context.Background(), &model.BlockListItem{Username: "demo", JID: "demo@example.org"})
	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)
}
