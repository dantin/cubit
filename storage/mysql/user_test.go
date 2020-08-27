package mysql

import (
	"context"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/dantin/cubit/model"
	"github.com/dantin/cubit/util/pool"
	"github.com/dantin/cubit/xmpp"
	"github.com/dantin/cubit/xmpp/jid"
	"github.com/stretchr/testify/require"
)

func newUserMock() (*mySQLUser, sqlmock.Sqlmock) {
	s, sqlMock := newStorageMock()
	return &mySQLUser{
		mySQLStorage: s,
		pool:         pool.NewBufferPool(),
	}, sqlMock
}

func TestMySQLStorage_InsertUser(t *testing.T) {
	from, _ := jid.NewWithString("alice@example.org/desktop", true)
	to, _ := jid.NewWithString("alice@example.org", true)
	p := xmpp.NewPresence(from, to, xmpp.UnavailableType)

	user := model.User{
		Username:     "alice",
		Password:     "passwd",
		LastPresence: p,
	}

	s, mock := newUserMock()
	mock.ExpectExec("INSERT INTO users (.+) ON DUPLICATE KEY UPDATE (.+)").
		WithArgs("alice", "passwd", p.String(), "passwd", p.String()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := s.UpsertUser(context.Background(), &user)

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)

	// err case
	s, mock = newUserMock()
	mock.ExpectExec("INSERT INTO users (.+) ON DUPLICATE KEY UPDATE (.+)").
		WithArgs("alice", "passwd", p.String(), "passwd", p.String()).
		WillReturnError(errMySQLStorage)

	err = s.UpsertUser(context.Background(), &user)

	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)
}

func TestMySQLStorage_DeleteUser(t *testing.T) {
	s, mock := newUserMock()
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM offline_messages (.+)").
		WithArgs("alice").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM roster_items (.+)").
		WithArgs("alice").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM roster_versions (.+)").
		WithArgs("alice").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM private_storage (.+)").
		WithArgs("alice").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM vcards (.+)").
		WithArgs("alice").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM users (.+)").
		WithArgs("alice").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := s.DeleteUser(context.Background(), "alice")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)

	// error case
	s, mock = newUserMock()
	mock.ExpectBegin()
	mock.ExpectExec("DELETE FROM offline_messages (.+)").
		WithArgs("alice").
		WillReturnError(errMySQLStorage)
	mock.ExpectRollback()

	err = s.DeleteUser(context.Background(), "alice")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)
}

func TestMySQLStorage_FetchUser(t *testing.T) {
	from, _ := jid.NewWithString("alice@example.org/desktop", true)
	to, _ := jid.NewWithString("alice@example.org", true)
	p := xmpp.NewPresence(from, to, xmpp.UnavailableType)

	var cols = []string{"username", "password", "last_presence", "last_presence_at"}

	s, mock := newUserMock()
	mock.ExpectQuery("SELECT (.+) FROM users (.+)").
		WithArgs("alice").
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow("alice", "passwd", p.String(), time.Now()))

	usr, err := s.FetchUser(context.Background(), "alice")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
	require.NotNil(t, usr)

	// empty
	s, mock = newUserMock()
	mock.ExpectQuery("SELECT (.+) FROM users (.+)").
		WithArgs("alice").
		WillReturnRows(sqlmock.NewRows(cols))

	_, err = s.FetchUser(context.Background(), "alice")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)

	// error case
	s, mock = newUserMock()
	mock.ExpectQuery("SELECT (.+) FROM users (.+)").
		WithArgs("alice").
		WillReturnError(errMocked)

	_, err = s.FetchUser(context.Background(), "alice")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMocked, err)
}

func TestMySQLStorage_UserExists(t *testing.T) {
	cols := []string{"count"}

	s, mock := newUserMock()
	mock.ExpectQuery("SELECT COUNT(.+) FROM users (.+)").
		WithArgs("alice").
		WillReturnRows(sqlmock.NewRows(cols).AddRow(1))

	ok, err := s.UserExists(context.Background(), "alice")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
	require.True(t, ok)

	// error case
	s, mock = newUserMock()
	mock.ExpectQuery("SELECT COUNT(.+) FROM users (.+)").
		WithArgs("alice").
		WillReturnError(errMySQLStorage)

	_, err = s.UserExists(context.Background(), "alice")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)
}
