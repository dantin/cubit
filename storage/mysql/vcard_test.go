package mysql

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/dantin/cubit/xmpp"
	"github.com/stretchr/testify/require"
)

func newVCardMock() (*mySQLVCard, sqlmock.Sqlmock) {
	s, sqlMock := newStorageMock()
	return &mySQLVCard{
		mySQLStorage: s,
	}, sqlMock
}

func TestMySQLStorage_InsertVCard(t *testing.T) {
	vCard := xmpp.NewElementName("vCard")
	rawXML := vCard.String()

	s, mock := newVCardMock()
	mock.ExpectExec("INSERT INTO vcards (.+) ON DUPLICATE KEY UPDATE (.+)").
		WithArgs("alice", rawXML, rawXML).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := s.UpsertVCard(context.Background(), vCard, "alice")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)

	// error case
	s, mock = newVCardMock()
	mock.ExpectExec("INSERT INTO vcards (.+) ON DUPLICATE KEY UPDATE (.+)").
		WithArgs("alice", rawXML, rawXML).
		WillReturnError(errMySQLStorage)

	err = s.UpsertVCard(context.Background(), vCard, "alice")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)
}

func TestMySQLStorage_FetchVCard(t *testing.T) {
	var cols = []string{"vcard"}

	s, mock := newVCardMock()
	mock.ExpectQuery("SELECT (.+) FROM vcards (.+)").
		WithArgs("alice").
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow("<vCard><FN>Alice</FN></vCard>"))

	vCard, err := s.FetchVCard(context.Background(), "alice")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
	require.NotNil(t, vCard)

	// empty
	s, mock = newVCardMock()
	mock.ExpectQuery("SELECT (.+) FROM vcards (.+)").
		WithArgs("alice").
		WillReturnRows(sqlmock.NewRows(cols))

	vCard, err = s.FetchVCard(context.Background(), "alice")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
	require.Nil(t, vCard)

	// error case
	s, mock = newVCardMock()
	mock.ExpectQuery("SELECT (.+) FROM vcards (.+)").
		WithArgs("alice").
		WillReturnError(errMySQLStorage)

	vCard, err = s.FetchVCard(context.Background(), "alice")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)
}
