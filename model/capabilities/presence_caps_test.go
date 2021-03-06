package capsmodel

import (
	"bytes"
	"testing"

	"github.com/dantin/cubit/xmpp"
	"github.com/dantin/cubit/xmpp/jid"
	"github.com/stretchr/testify/require"
)

func TestPresenceCapabilities(t *testing.T) {
	j1, _ := jid.NewWithString("username@example.org", true)

	var p1, p2 PresenceCaps
	p1 = PresenceCaps{
		Presence: xmpp.NewPresence(j1, j1, xmpp.AvailableType),
	}

	buf := new(bytes.Buffer)
	require.Nil(t, p1.ToBytes(buf))
	require.Nil(t, p2.FromBytes(buf))
	require.Equal(t, p1, p2)

	var p3, p4 PresenceCaps
	p3 = PresenceCaps{
		Presence: xmpp.NewPresence(j1, j1, xmpp.AvailableType),
		Caps: &Capabilities{
			Node: "http://localhost",
			Ver:  "1.0",
		},
	}

	buf = new(bytes.Buffer)
	require.Nil(t, p3.ToBytes(buf))
	require.Nil(t, p4.FromBytes(buf))
	require.Equal(t, p3, p4)
}
