package memorystorage

import (
	"context"
	"testing"

	capsmodel "github.com/dantin/cubit/model/capabilities"
	"github.com/dantin/cubit/xmpp"
	"github.com/dantin/cubit/xmpp/jid"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage_FetchPresencesMatchingJID(t *testing.T) {
	const allocID = "1"

	j1, _ := jid.NewWithString("alice@example.org/desktop", true)
	j2, _ := jid.NewWithString("bob@example.org/desktop", true)
	j3, _ := jid.NewWithString("alice@example.org/surface", true)
	j4, _ := jid.NewWithString("carl@jabber.org/desktop", true)

	p1 := xmpp.NewPresence(j1, j1.ToBareJID(), xmpp.AvailableType)
	p2 := xmpp.NewPresence(j1, j1.ToBareJID(), xmpp.AvailableType)
	p3 := xmpp.NewPresence(j1, j1.ToBareJID(), xmpp.AvailableType)
	p4 := xmpp.NewPresence(j1, j1.ToBareJID(), xmpp.AvailableType)

	s := NewPresences()

	ok, err := s.UpsertPresence(context.Background(), p1, j1, allocID)
	require.True(t, ok)
	require.Nil(t, err)

	ok, err = s.UpsertPresence(context.Background(), p2, j2, allocID)
	require.True(t, ok)
	require.Nil(t, err)

	ok, err = s.UpsertPresence(context.Background(), p3, j3, allocID)
	require.True(t, ok)
	require.Nil(t, err)

	ok, err = s.UpsertPresence(context.Background(), p4, j4, allocID)
	require.True(t, ok)
	require.Nil(t, err)

	// updating presence
	ok, err = s.UpsertPresence(context.Background(), p1, j1, allocID)
	require.False(t, ok)
	require.Nil(t, err)

	mJID, _ := jid.NewWithString("example.org", true)
	presences, _ := s.FetchPresencesMatchingJID(context.Background(), mJID)
	require.Len(t, presences, 3)

	mJID, _ = jid.NewWithString("example.org/desktop", true)
	presences, _ = s.FetchPresencesMatchingJID(context.Background(), mJID)
	require.Len(t, presences, 2)

	mJID, _ = jid.NewWithString("jabber.org", true)
	presences, _ = s.FetchPresencesMatchingJID(context.Background(), mJID)
	require.Len(t, presences, 1)

	_ = s.DeletePresence(context.Background(), j2)
	mJID, _ = jid.NewWithString("example.org/desktop", true)
	presences, _ = s.FetchPresencesMatchingJID(context.Background(), mJID)
	require.Len(t, presences, 1)

	_ = s.ClearPresences(context.Background())
	mJID, _ = jid.NewWithString("example.org", true)
	presences, _ = s.FetchPresencesMatchingJID(context.Background(), mJID)
	require.Len(t, presences, 0)
}

func TestMemoryStorage_InsertCapabilities(t *testing.T) {
	caps := capsmodel.Capabilities{Node: "n1", Ver: "1", Features: []string{"ns"}}
	s := NewPresences()

	EnableMockedError()
	err := s.UpsertCapabilities(context.Background(), &caps)
	require.Equal(t, ErrMocked, err)
	DisableMockedError()

	err = s.UpsertCapabilities(context.Background(), &caps)
	require.Nil(t, err)
}

func TestMemoryStorage_FetchCapabilities(t *testing.T) {
	caps := capsmodel.Capabilities{Node: "n1", Ver: "1", Features: []string{"ns"}}
	s := NewPresences()
	_ = s.UpsertCapabilities(context.Background(), &caps)

	EnableMockedError()
	_, err := s.FetchCapabilities(context.Background(), "n1", "1")
	require.Equal(t, ErrMocked, err)
	DisableMockedError()

	cs, _ := s.FetchCapabilities(context.Background(), "n1", "-1")
	require.Nil(t, cs)

	cs, _ = s.FetchCapabilities(context.Background(), "n1", "1")
	require.NotNil(t, cs)
}
