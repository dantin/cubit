package mysql

import (
	"context"
	"encoding/json"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	capsmodel "github.com/dantin/cubit/model/capabilities"
	"github.com/dantin/cubit/util/pool"
	"github.com/dantin/cubit/xmpp"
	"github.com/dantin/cubit/xmpp/jid"
	"github.com/stretchr/testify/require"
)

func newPresencesMock() (*mySQLPresences, sqlmock.Sqlmock) {
	s, sqlMock := newStorageMock()
	return &mySQLPresences{
		mySQLStorage: s,
		pool:         pool.NewBufferPool(),
	}, sqlMock
}

func TestMySQLStorage_UpsertPresence(t *testing.T) {
	s, mock := newPresencesMock()
	mock.ExpectExec("INSERT INTO presences (.+) VALUES (.+) ON DUPLICATE KEY UPDATE (.+)").
		WithArgs("demo", "example.org", "desktop", `<presence from="demo@example.org/desktop" to="demo@example.org"/>`, "", "", "alloc-1234", `<presence from="demo@example.org/desktop" to="demo@example.org"/>`, "", "", "alloc-1234").
		WillReturnResult(sqlmock.NewResult(1, 1))

	j, _ := jid.NewWithString("demo@example.org/desktop", true)
	inserted, err := s.UpsertPresence(context.Background(), xmpp.NewPresence(j, j.ToBareJID(), xmpp.AvailableType), j, "alloc-1234")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
	require.True(t, inserted)
}

func TestMySQLStorage_FetchPresence(t *testing.T) {
	var columns = []string{"presence", "c.node", "c.ver", "c.features"}

	s, mock := newPresencesMock()
	mock.ExpectQuery("SELECT presence, c.node, c.ver, c.features FROM presences AS p, capabilities AS c WHERE \\(username = \\? AND domain = \\? AND resource = \\? AND p.node = c.node AND p.ver = c.ver\\)").
		WithArgs("demo", "example.org", "desktop").
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow("<presence/>", "example.org", "v1.0", `["urn:xmpp:ping"]`))

	j, _ := jid.NewWithString("demo@example.org/desktop", true)
	presenceCaps, err := s.FetchPresence(context.Background(), j)
	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)

	require.NotNil(t, presenceCaps)
	require.Equal(t, "example.org", presenceCaps.Caps.Node)
	require.Equal(t, "v1.0", presenceCaps.Caps.Ver)
	require.Len(t, presenceCaps.Caps.Features, 1)
	require.Equal(t, "urn:xmpp:ping", presenceCaps.Caps.Features[0])
}

func TestMySQLStorage_FetchPresencesMatchingJID(t *testing.T) {
	var columns = []string{"presence", "c.node", "c.ver", "c.features"}

	s, mock := newPresencesMock()
	mock.ExpectQuery("SELECT presence, c.node, c.ver, c.features FROM presences AS p, capabilities AS c WHERE \\(username = \\? AND domain = \\? AND resource = \\? AND p.node = c.node AND p.ver = c.ver\\)").
		WithArgs("demo", "example.org", "desktop").
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow("<presence/>", "example.org", "v1.0", `["urn:xmpp:ping"]`).
			AddRow("<presence/>", "example.org", "v1.0", `["urn:xmpp:ping"]`))

	j, _ := jid.NewWithString("demo@example.org/desktop", true)
	presenceCaps, err := s.FetchPresencesMatchingJID(context.Background(), j)
	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)

	require.NotNil(t, presenceCaps)
	require.Len(t, presenceCaps, 2)
	require.Equal(t, "example.org", presenceCaps[0].Caps.Node)
	require.Equal(t, "v1.0", presenceCaps[0].Caps.Ver)
	require.Len(t, presenceCaps[0].Caps.Features, 1)
	require.Equal(t, "urn:xmpp:ping", presenceCaps[0].Caps.Features[0])
}

func TestMySQLStorage_DeletePresence(t *testing.T) {
	j, _ := jid.NewWithString("demo@example.org/desktop", true)

	s, mock := newPresencesMock()
	mock.ExpectExec("DELETE FROM presences WHERE \\(username = \\? AND domain = \\? AND resource = \\?\\)").
		WithArgs(j.Node(), j.Domain(), j.Resource()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := s.DeletePresence(context.Background(), j)

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
}

func TestMySQLStorage_DeleteAllocationPresence(t *testing.T) {
	s, mock := newPresencesMock()
	mock.ExpectExec("DELETE FROM presences WHERE allocation_id = ?").
		WithArgs("1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := s.DeleteAllocationPresences(context.Background(), "1")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
}

func TestMySQLStorage_ClearPresence(t *testing.T) {
	s, mock := newPresencesMock()
	mock.ExpectExec("DELETE FROM presences").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := s.ClearPresences(context.Background())

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
}

func TestMySQLStorage_UpsertCapabilities(t *testing.T) {
	features := []string{"jabber:iq:last"}

	b, _ := json.Marshal(&features)

	s, mock := newPresencesMock()
	mock.ExpectExec("INSERT INTO capabilities (.+) VALUES (.+) ON DUPLICATE KEY UPDATE features = \\?, updated_at = NOW\\(\\)").
		WithArgs("n1", "123", b, b).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := s.UpsertCapabilities(context.Background(), &capsmodel.Capabilities{Node: "n1", Ver: "123", Features: features})

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)

	// error case
	s, mock = newPresencesMock()
	mock.ExpectExec("INSERT INTO capabilities (.+) VALUES (.+) ON DUPLICATE KEY UPDATE features = \\?, updated_at = NOW\\(\\)").
		WithArgs("n1", "123", b, b).
		WillReturnError(errMySQLStorage)

	err = s.UpsertCapabilities(context.Background(), &capsmodel.Capabilities{Node: "n1", Ver: "123", Features: features})

	require.Nil(t, mock.ExpectationsWereMet())
	require.NotNil(t, err)
	require.Equal(t, errMySQLStorage, err)
}

func TestMySQLStorage_FetchCapabilities(t *testing.T) {
	var columns = []string{"features"}
	s, mock := newPresencesMock()
	mock.ExpectQuery("SELECT features FROM capabilities WHERE \\(node = . AND ver = .\\)").
		WithArgs("n1", "123").
		WillReturnRows(sqlmock.NewRows(columns).AddRow(`["jabber:iq:last"]`))

	caps, err := s.FetchCapabilities(context.Background(), "n1", "123")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
	require.NotNil(t, caps)
	require.Len(t, caps.Features, 1)
	require.Equal(t, "jabber:iq:last", caps.Features[0])

	// error case
	s, mock = newPresencesMock()
	mock.ExpectQuery("SELECT features FROM capabilities WHERE \\(node = . AND ver = .\\)").
		WithArgs("n1", "123").
		WillReturnError(errMySQLStorage)

	caps, err = s.FetchCapabilities(context.Background(), "n1", "123")

	require.Nil(t, mock.ExpectationsWereMet())
	require.NotNil(t, err)
	require.Nil(t, caps)
	require.Equal(t, errMySQLStorage, err)
}
