package model

import (
	"bytes"
	"testing"
	"time"

	"github.com/dantin/cubit/xmpp"
	"github.com/dantin/cubit/xmpp/jid"
	"github.com/stretchr/testify/require"
)

func TestModelUser(t *testing.T) {
	var user1 User

	j1, _ := jid.NewWithString("username@example.org", true)
	j2, _ := jid.NewWithString("username@example.org", true)

	user1.Username = "username"
	user1.Password = "passwd"
	user1.Role = "role"
	user1.LastPresence = xmpp.NewPresence(j1, j2, xmpp.AvailableType)

	buf := new(bytes.Buffer)
	require.Nil(t, user1.ToBytes(buf))
	user2 := User{}
	require.Nil(t, user2.FromBytes(buf))
	require.Equal(t, user1.Username, user2.Username)
	require.Equal(t, user1.Password, user2.Password)
	require.Equal(t, user1.Role, user2.Role)
	require.Equal(t, user1.LastPresence.String(), user2.LastPresence.String())
	require.NotEqual(t, time.Time{}, user2.LastPresenceAt)
}
