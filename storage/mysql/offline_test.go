package mysql

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/dantin/cubit/util/pool"
	"github.com/dantin/cubit/xmpp"
	"github.com/dantin/cubit/xmpp/jid"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"
)

func newOfflineMock() (*mySQLOffline, sqlmock.Sqlmock) {
	s, sqlMock := newStorageMock()
	return &mySQLOffline{
		mySQLStorage: s,
		pool:         pool.NewBufferPool(),
	}, sqlMock
}

func TestMySQLStorage_InsertOfflineMessages(t *testing.T) {
	j, _ := jid.NewWithString("demo@example.org/desktop", false)
	message := xmpp.NewElementName("message")
	message.SetID(uuid.New())
	message.AppendElement(xmpp.NewElementName("body"))
	m, _ := xmpp.NewMessageFromElement(message, j, j)
	messageXML := m.String()

	s, mock := newOfflineMock()
	mock.ExpectExec("INSERT INTO offline_messages (.+)").
		WithArgs("demo", messageXML).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := s.InsertOfflineMessage(context.Background(), m, "demo")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)

	s, mock = newOfflineMock()
	mock.ExpectExec("INSERT INTO offline_messages (.+)").
		WithArgs("demo", messageXML).
		WillReturnError(errMySQLStorage)

	err = s.InsertOfflineMessage(context.Background(), m, "demo")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)
}

func TestMySQLStorage_CountOfflineMessages(t *testing.T) {
	countColumns := []string{"count"}

	s, mock := newOfflineMock()
	mock.ExpectQuery("SELECT COUNT(.+) FROM offline_messages (.+)").
		WithArgs("demo").
		WillReturnRows(sqlmock.NewRows(countColumns).AddRow(1))

	cnt, _ := s.CountOfflineMessages(context.Background(), "demo")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, 1, cnt)

	s, mock = newOfflineMock()
	mock.ExpectQuery("SELECT COUNT(.+) FROM offline_messages (.+)").
		WithArgs("demo").
		WillReturnRows(sqlmock.NewRows(countColumns))

	cnt, _ = s.CountOfflineMessages(context.Background(), "demo")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, 0, cnt)

	s, mock = newOfflineMock()
	mock.ExpectQuery("SELECT COUNT(.+) FROM offline_messages (.+)").
		WithArgs("demo").
		WillReturnError(errMySQLStorage)

	_, err := s.CountOfflineMessages(context.Background(), "demo")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)
}

func TestMySQLStorage_FetchOfflineMessages(t *testing.T) {
	var offlineMessagesColumns = []string{"data"}

	s, mock := newOfflineMock()
	mock.ExpectQuery("SELECT (.+) FROM offline_messages (.+)").
		WithArgs("demo").
		WillReturnRows(sqlmock.NewRows(offlineMessagesColumns).AddRow("<message id='1'><body>Hi</body></message>"))

	msgs, _ := s.FetchOfflineMessages(context.Background(), "demo")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, 1, len(msgs))

	s, mock = newOfflineMock()
	mock.ExpectQuery("SELECT (.+) FROM offline_messages (.+)").
		WithArgs("demo").
		WillReturnRows(sqlmock.NewRows(offlineMessagesColumns))

	msgs, _ = s.FetchOfflineMessages(context.Background(), "demo")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, 0, len(msgs))

	s, mock = newOfflineMock()
	mock.ExpectQuery("SELECT (.+) FROM offline_messages (.+)").
		WithArgs("demo").
		WillReturnRows(sqlmock.NewRows(offlineMessagesColumns).AddRow("<message id='1'><body>Hi"))

	_, err := s.FetchOfflineMessages(context.Background(), "demo")
	require.Nil(t, mock.ExpectationsWereMet())
	require.NotNil(t, err)

	s, mock = newOfflineMock()
	mock.ExpectQuery("SELECT (.+) FROM offline_messages (.+)").
		WithArgs("demo").
		WillReturnError(errMySQLStorage)

	_, err = s.FetchOfflineMessages(context.Background(), "demo")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)
}

func TestMySQLStorage_DeleteOfflineMessages(t *testing.T) {
	s, mock := newOfflineMock()
	mock.ExpectExec("DELETE FROM offline_messages (.+)").
		WithArgs("demo").WillReturnResult(sqlmock.NewResult(0, 1))

	err := s.DeleteOfflineMessages(context.Background(), "demo")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)

	s, mock = newOfflineMock()
	mock.ExpectExec("DELETE FROM offline_messages (.+)").
		WithArgs("demo").WillReturnError(errMySQLStorage)

	err = s.DeleteOfflineMessages(context.Background(), "demo")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)
}
