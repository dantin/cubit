package mysql

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/dantin/cubit/util/pool"
	"github.com/dantin/cubit/xmpp"
	"github.com/stretchr/testify/require"
)

func newPrivateMock() (*mySQLPrivate, sqlmock.Sqlmock) {
	s, sqlMock := newStorageMock()
	return &mySQLPrivate{
		mySQLStorage: s,
		pool:         pool.NewBufferPool(),
	}, sqlMock
}

func TestMySQLStorage_InsertPrivateXML(t *testing.T) {
	private := xmpp.NewElementNamespace("node", "node:ns")
	rawXML := private.String()

	s, mock := newPrivateMock()
	mock.ExpectExec("INSERT INTO private_storage (.+) ON DUPLICATE KEY UPDATE (.)").
		WithArgs("demo", "node:ns", rawXML, rawXML).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := s.UpsertPrivateXML(context.Background(), []xmpp.XElement{private}, "node:ns", "demo")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)

	// error
	s, mock = newPrivateMock()
	mock.ExpectExec("INSERT INTO private_storage (.+) ON DUPLICATE KEY UPDATE (.)").
		WithArgs("demo", "node:ns", rawXML, rawXML).
		WillReturnError(errMySQLStorage)

	err = s.UpsertPrivateXML(context.Background(), []xmpp.XElement{private}, "node:ns", "demo")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)
}

func TestMySQLStorage_FetchPrivateXML(t *testing.T) {
	var columns = []string{"data"}

	s, mock := newPrivateMock()
	mock.ExpectQuery("SELECT (.+) FROM private_storage (.+)").
		WithArgs("demo", "node:ns").
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow("<node xmlns='node:ns'><stuff/></node>"))

	elems, err := s.FetchPrivateXML(context.Background(), "node:ns", "demo")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
	require.Len(t, elems, 1)

	s, mock = newPrivateMock()

	mock.ExpectQuery("SELECT (.+) FROM private_storage (.+)").
		WithArgs("demo", "node:ns").
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow("<node xmlns='node:ns'><stuff/>"))

	elems, err = s.FetchPrivateXML(context.Background(), "node:ns", "demo")

	require.Nil(t, mock.ExpectationsWereMet())
	require.NotNil(t, err)
	require.Len(t, elems, 0)

	s, mock = newPrivateMock()

	mock.ExpectQuery("SELECT (.+) FROM private_storage (.+)").
		WithArgs("demo", "node:ns").
		WillReturnRows(sqlmock.NewRows(columns))

	elems, err = s.FetchPrivateXML(context.Background(), "node:ns", "demo")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
	require.Len(t, elems, 0)

	s, mock = newPrivateMock()

	mock.ExpectQuery("SELECT (.+) FROM private_storage (.+)").
		WithArgs("demo", "node:ns").
		WillReturnError(errMySQLStorage)

	elems, err = s.FetchPrivateXML(context.Background(), "node:ns", "demo")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)
	require.Len(t, elems, 0)
}
