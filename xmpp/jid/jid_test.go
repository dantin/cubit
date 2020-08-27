package jid_test

import (
	"testing"

	"github.com/dantin/cubit/xmpp/jid"
	"github.com/stretchr/testify/require"
)

func TestBadJID(t *testing.T) {
	_, err := jid.NewWithString("username@", false)
	require.NotNil(t, err)
	longStr := ""
	for i := 0; i < 1074; i++ {
		longStr += "a"
	}
	_, err2 := jid.New(longStr, "example.org", "res", false)
	require.NotNil(t, err2)
	_, err3 := jid.New("username", longStr, "res", false)
	require.NotNil(t, err3)
	_, err4 := jid.New("username", "example.org", longStr, false)
	require.NotNil(t, err4)
}

func TestNewJID(t *testing.T) {
	j1, err := jid.New("username", "example.org", "res", false)
	require.Nil(t, err)
	require.Equal(t, "username", j1.Node())
	require.Equal(t, "example.org", j1.Domain())
	require.Equal(t, "res", j1.Resource())
	j2, err := jid.New("username", "example.org", "res", true)
	require.Nil(t, err)
	require.Equal(t, "username", j2.Node())
	require.Equal(t, "example.org", j2.Domain())
	require.Equal(t, "res", j2.Resource())
}

func TestEmptyJID(t *testing.T) {
	j, err := jid.NewWithString("", true)
	require.Nil(t, err)
	require.Equal(t, "", j.Node())
	require.Equal(t, "", j.Domain())
	require.Equal(t, "", j.Resource())
}

func TestNewJIDString(t *testing.T) {
	j, err := jid.NewWithString("username@example.org/res", false)
	require.Nil(t, err)
	require.Equal(t, "username", j.Node())
	require.Equal(t, "example.org", j.Domain())
	require.Equal(t, "res", j.Resource())
	require.Equal(t, "username@example.org", j.ToBareJID().String())
	require.Equal(t, "username@example.org/res", j.String())
}

func TestServerJID(t *testing.T) {
	j1, _ := jid.NewWithString("example.org", false)
	j2, _ := jid.NewWithString("username@example.org", false)
	j3, _ := jid.NewWithString("example.org/res", false)
	require.True(t, j1.IsServer())
	require.False(t, j2.IsServer())
	require.True(t, j3.IsServer() && j3.IsFull())
}

func TestBareJID(t *testing.T) {
	j1, _ := jid.New("username", "example.org", "res", false)
	require.True(t, j1.ToBareJID().IsBare())
	j2, _ := jid.NewWithString("example.org/res", false)
	require.False(t, j2.ToBareJID().IsBare())
}

func TestFullJID(t *testing.T) {
	j1, _ := jid.New("username", "example.org", "res", false)
	j2, _ := jid.New("", "example.org/res", "res", false)
	require.True(t, j1.IsFullWithUser())
	require.True(t, j2.IsFullWithServer())
}

func TestMatchesJID(t *testing.T) {
	j1, _ := jid.NewWithString("username@example.org/res1", false)
	j2, _ := jid.NewWithString("username@example.org", false)
	j3, _ := jid.NewWithString("example.org", false)
	j4, _ := jid.NewWithString("example.org/res1", false)
	j5, _ := jid.NewWithString("username@example2.org/res2", false)
	require.True(t, j1.MatchesWithOptions(j1, jid.MatchesNode|jid.MatchesDomain|jid.MatchesResource))
	require.True(t, j1.MatchesWithOptions(j2, jid.MatchesNode|jid.MatchesDomain))
	require.True(t, j1.MatchesWithOptions(j3, jid.MatchesDomain))
	require.True(t, j1.MatchesWithOptions(j4, jid.MatchesDomain|jid.MatchesResource))

	require.False(t, j1.MatchesWithOptions(j2, jid.MatchesNode|jid.MatchesDomain|jid.MatchesResource))
	require.False(t, j5.MatchesWithOptions(j2, jid.MatchesNode|jid.MatchesDomain))
	require.False(t, j5.MatchesWithOptions(j3, jid.MatchesDomain))
	require.False(t, j5.MatchesWithOptions(j4, jid.MatchesDomain|jid.MatchesResource))
}

func TestBadPrep(t *testing.T) {
	badNode := string([]byte{255, 255, 255})
	badDomain := string([]byte{255, 255, 255})
	badResource := string([]byte{255, 255, 255})
	j, err := jid.New(badNode, "example.org", "res", false)
	require.Nil(t, j)
	require.NotNil(t, err)
	j2, err := jid.New("username", badDomain, "res", false)
	require.Nil(t, j2)
	require.NotNil(t, err)
	j3, err := jid.New("username", "example.org", badResource, false)
	require.Nil(t, j3)
	require.NotNil(t, err)
}

func TestParseEmptyJID(t *testing.T) {
	j, err := jid.NewWithString("username@example.net/", false)
	require.Nil(t, j)
	require.NotNil(t, err)
}
