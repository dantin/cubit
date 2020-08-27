package mysql

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	rostermodel "github.com/dantin/cubit/model/roster"
	"github.com/dantin/cubit/xmpp"
	"github.com/stretchr/testify/require"
)

func newRosterMock() (*mySQLRoster, sqlmock.Sqlmock) {
	s, sqlMock := newStorageMock()
	return &mySQLRoster{
		mySQLStorage: s,
	}, sqlMock
}

func TestMySQLStorage_InsertRosterItem(t *testing.T) {
	groups := []string{"Work", "Home"}
	ri := rostermodel.Item{
		Username:     "user",
		JID:          "user@example.org",
		Name:         "username",
		Subscription: "detail",
		Ask:          false,
		Ver:          1,
		Groups:       groups,
	}

	groupsBytes, _ := json.Marshal(groups)
	args := []driver.Value{
		ri.Username,
		ri.JID,
		ri.Name,
		ri.Subscription,
		groupsBytes,
		ri.Ask,
		ri.Username,
		ri.Name,
		ri.Subscription,
		groupsBytes,
		ri.Ask,
	}

	s, mock := newRosterMock()
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO roster_versions (.+) ON DUPLICATE KEY UPDATE (.+)").
		WithArgs("user").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO roster_items (.+) ON DUPLICATE KEY UPDATE (.+)").
		WithArgs(args...).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM roster_groups (.+)").
		WithArgs("user", "user@example.org").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec("INSERT INTO roster_groups (.+)").
		WithArgs("user", "user@example.org", "Work").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO roster_groups (.+)").
		WithArgs("user", "user@example.org", "Home").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT (.+) FROM roster_versions (.+)").
		WithArgs("user").
		WillReturnRows(sqlmock.NewRows([]string{"ver", "deletionVer"}).AddRow(1, 0))
	mock.ExpectCommit()

	_, err := s.UpsertRosterItem(context.Background(), &ri)

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
}

func TestMySQLStorage_DeleteRosterItem(t *testing.T) {
	s, mock := newRosterMock()
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO roster_versions (.+) ON DUPLICATE KEY UPDATE (.+)").
		WithArgs("user").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM roster_groups (.+)").
		WithArgs("user", "contact").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM roster_items (.+)").
		WithArgs("user", "contact").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT (.+) FROM roster_versions (.+)").
		WithArgs("user").
		WillReturnRows(sqlmock.NewRows([]string{"ver", "deletionVer"}).AddRow(1, 0))
	mock.ExpectCommit()

	_, err := s.DeleteRosterItem(context.Background(), "user", "contact")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)

	// error case
	s, mock = newRosterMock()
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO roster_versions (.+) ON DUPLICATE KEY UPDATE (.+)").
		WithArgs("user").
		WillReturnError(errMySQLStorage)
	mock.ExpectRollback()

	_, err = s.DeleteRosterItem(context.Background(), "user", "contact")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)
}

func TestMySQLStorage_FetchRosterItem(t *testing.T) {
	var cols = []string{"user", "contact", "name", "subscription", "`groups`", "ask", "ver"}

	s, mock := newRosterMock()
	mock.ExpectQuery("SELECT (.+) FROM roster_items (.+)").
		WithArgs("alice").
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow("alice", "bob", "Bob", "both", "", false, 0))
	mock.ExpectQuery("SELECT (.+) FROM roster_versions (.+)").
		WithArgs("alice").
		WillReturnRows(sqlmock.NewRows([]string{"ver", "deletionVer"}).AddRow(0, 0))

	rosterItem, _, err := s.FetchRosterItems(context.Background(), "alice")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
	require.Len(t, rosterItem, 1)

	// error case
	s, mock = newRosterMock()
	mock.ExpectQuery("SELECT (.+) FROM roster_items (.+)").
		WithArgs("alice").
		WillReturnError(errMySQLStorage)

	_, _, err = s.FetchRosterItems(context.Background(), "alice")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)

	// single mode
	s, mock = newRosterMock()
	mock.ExpectQuery("SELECT (.+) FROM roster_items (.+)").
		WithArgs("alice", "bob").
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow("alice", "bob", "Bob", "both", "", false, 0))

	_, err = s.FetchRosterItem(context.Background(), "alice", "bob")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)

	// none
	s, mock = newRosterMock()
	mock.ExpectQuery("SELECT (.+) FROM roster_items (.+)").
		WithArgs("alice", "bob").
		WillReturnRows(sqlmock.NewRows(cols))

	_, err = s.FetchRosterItem(context.Background(), "alice", "bob")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)

	// single error case
	s, mock = newRosterMock()
	mock.ExpectQuery("SELECT (.+) FROM roster_items (.+)").
		WithArgs("alice", "bob").
		WillReturnError(errMySQLStorage)

	_, err = s.FetchRosterItem(context.Background(), "alice", "bob")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)

	// by groups
	var cols2 = []string{"ris.user", "ris.contact", "ris.name", "ris.subscription", "ris.`group`", "ris.ask", "ris.ver"}
	s, mock = newRosterMock()
	mock.ExpectQuery("SELECT (.+) FROM roster_items ris LEFT JOIN roster_groups g ON ris.username = g.username (.+)").
		WithArgs("alice", "Work").
		WillReturnRows(sqlmock.NewRows(cols2).
			AddRow("alice", "bob", "Bob", "both", `["Work"]`, false, 0))
	mock.ExpectQuery("SELECT (.+) FROM roster_versions (.+)").
		WithArgs("alice").
		WillReturnRows(sqlmock.NewRows([]string{"ver", "deletionVer"}).AddRow(0, 0))

	_, _, err = s.FetchRosterItemsInGroups(context.Background(), "alice", []string{"Work"})

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
}

func TestMySQLStorage_InsertRosterNotification(t *testing.T) {
	rn := rostermodel.Notification{
		Contact:  "alice",
		JID:      "bob",
		Presence: &xmpp.Presence{},
	}
	presenceXML := rn.Presence.String()

	args := []driver.Value{
		rn.Contact,
		rn.JID,
		presenceXML,
		presenceXML,
	}

	s, mock := newRosterMock()
	mock.ExpectExec("INSERT INTO roster_notifications (.+) ON DUPLICATE KEY UPDATE (.+)").
		WithArgs(args...).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := s.UpsertRosterNotification(context.Background(), &rn)

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)

	s, mock = newRosterMock()
	mock.ExpectExec("INSERT INTO roster_notifications (.+) ON DUPLICATE KEY UPDATE (.+)").
		WithArgs(args...).
		WillReturnError(errMySQLStorage)

	err = s.UpsertRosterNotification(context.Background(), &rn)

	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)
}

func TestMySQLStorage_DeleteRosterNotification(t *testing.T) {
	s, mock := newRosterMock()
	mock.ExpectExec("DELETE FROM roster_notifications (.+)").
		WithArgs("user", "contact").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := s.DeleteRosterNotification(context.Background(), "user", "contact")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)

	s, mock = newRosterMock()
	mock.ExpectExec("DELETE FROM roster_notifications (.+)").
		WithArgs("user", "contact").
		WillReturnError(errMySQLStorage)

	err = s.DeleteRosterNotification(context.Background(), "user", "contact")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)
}

func TestMySQLStorage_FetchRosterNotification(t *testing.T) {
	var cols = []string{"user", "contact", "elements"}

	s, mock := newRosterMock()
	mock.ExpectQuery("SELECT (.+) FROM roster_notifications (.+)").
		WithArgs("alice").
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow("alice", "bob", "<priority>8</priority>"))

	rosterNotification, err := s.FetchRosterNotifications(context.Background(), "alice")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
	require.Len(t, rosterNotification, 1)

	// empty result
	s, mock = newRosterMock()
	mock.ExpectQuery("SELECT (.+) FROM roster_notifications (.+)").
		WithArgs("alice").
		WillReturnRows(sqlmock.NewRows(cols))

	rosterNotification, err = s.FetchRosterNotifications(context.Background(), "alice")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
	require.Len(t, rosterNotification, 0)

	// error case
	s, mock = newRosterMock()
	mock.ExpectQuery("SELECT (.+) FROM roster_notifications (.+)").
		WithArgs("alice").
		WillReturnError(errMySQLStorage)

	rosterNotification, err = s.FetchRosterNotifications(context.Background(), "alice")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)

	// bad content
	s, mock = newRosterMock()
	mock.ExpectQuery("SELECT (.+) FROM roster_notifications (.+)").
		WithArgs("alice").
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow("alice", "bob", "<priority>8"))

	rosterNotification, err = s.FetchRosterNotifications(context.Background(), "alice")

	require.Nil(t, mock.ExpectationsWereMet())
	require.NotNil(t, err)
}

func TestMySQLStorage_FetchRosterGroups(t *testing.T) {
	s, mock := newRosterMock()
	mock.ExpectQuery("SELECT `group` FROM roster_groups WHERE username = (.+) GROUP BY (.+)").
		WithArgs("alice").
		WillReturnRows(sqlmock.NewRows([]string{"group"}).
			AddRow("Home").
			AddRow("Work"))

	groups, err := s.FetchRosterGroups(context.Background(), "alice")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
	require.Len(t, groups, 2)
	require.Equal(t, "Home", groups[0])
	require.Equal(t, "Work", groups[1])

	// error case
	s, mock = newRosterMock()
	mock.ExpectQuery("SELECT `group` FROM roster_groups WHERE username = (.+) GROUP BY (.+)").
		WithArgs("alice").
		WillReturnError(errMySQLStorage)

	groups, err = s.FetchRosterGroups(context.Background(), "alice")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)
}
