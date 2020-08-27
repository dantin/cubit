package rostermodel

import (
	"bytes"
	"testing"

	"github.com/dantin/cubit/xmpp"
	"github.com/dantin/cubit/xmpp/jid"
	"github.com/stretchr/testify/require"
)

func TestModelRosterNotification(t *testing.T) {
	var rn1, rn2 Notification

	j1, _ := jid.NewWithString("alex@example.org", true)
	j2, _ := jid.NewWithString("david@example.org", true)

	rn1 = Notification{
		Contact:  "david",
		JID:      "alex@example.org",
		Presence: xmpp.NewPresence(j1, j2, xmpp.AvailableType),
	}
	buf := new(bytes.Buffer)
	require.Nil(t, rn1.ToBytes(buf))
	require.Nil(t, rn2.FromBytes(buf))
	require.Equal(t, "alex@example.org", rn2.JID)
	require.Equal(t, "david", rn2.Contact)
	require.NotNil(t, rn1.Presence)
	require.NotNil(t, rn2.Presence)
	require.Equal(t, rn1.Presence.String(), rn2.Presence.String())
}
