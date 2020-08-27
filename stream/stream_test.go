package stream

import (
	"context"
	"testing"

	"github.com/dantin/cubit/xmpp"
	"github.com/dantin/cubit/xmpp/jid"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/require"
)

func TestMockC2Stream(t *testing.T) {
	j1, _ := jid.NewWithString("alice@example.org/desktop", false)
	j2, _ := jid.NewWithString("bob@example.org/desktop", false)
	id := uuid.New()

	stm := NewMockC2S(id, j1)

	require.Equal(t, "alice", stm.Username())
	require.Equal(t, "example.org", stm.Domain())
	require.Equal(t, "desktop", stm.Resource())
	require.Equal(t, "alice@example.org/desktop", stm.JID().String())
	require.Equal(t, id, stm.ID())

	stm.SetJID(j2)
	require.Equal(t, "bob@example.org/desktop", stm.JID().String())

	presence := xmpp.NewPresence(j1, j2, xmpp.AvailableType)
	presence.AppendElement(xmpp.NewElementName("status"))

	stm.SetPresence(presence)

	presenceElements := stm.Presence().Elements().All()
	require.Len(t, presenceElements, 1)

	elem := xmpp.NewElementName("message")
	stm.SendElement(context.Background(), elem)
	fetch := stm.ReceiveElement()

	require.NotNil(t, fetch)
	require.Equal(t, "message", fetch.Name())

	stm.Disconnect(context.Background(), nil)
	require.True(t, stm.IsDisconnected())

	stm.SetSecured(true)
	require.True(t, stm.IsSecured())

	stm.SetAuthenticated(true)
	require.True(t, stm.IsAuthenticated())
}
