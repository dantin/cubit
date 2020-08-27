package xmpp_test

import (
	"testing"

	"github.com/dantin/cubit/xmpp"
	"github.com/dantin/cubit/xmpp/jid"
	"github.com/stretchr/testify/require"
)

func TestMessageBuild(t *testing.T) {
	j, _ := jid.New("username", "example.org", "desktop", false)

	// wrong name
	elem := xmpp.NewElementName("iq")
	_, err := xmpp.NewMessageFromElement(elem, j, j)
	require.NotNil(t, err)

	// invalid type
	elem.SetName("message")
	elem.SetType("invalid")
	_, err = xmpp.NewMessageFromElement(elem, j, j)
	require.NotNil(t, err)

	// valid message
	elem.SetType(xmpp.ChatType)
	elem.AppendElement(xmpp.NewElementName("body"))
	message, err := xmpp.NewMessageFromElement(elem, j, j)
	require.Nil(t, err)
	require.NotNil(t, message)
	require.True(t, message.IsMessageWithBody())

	msg2 := xmpp.NewMessageType("id_123", xmpp.GroupChatType)
	require.Equal(t, "id_123", msg2.ID())
	require.Equal(t, xmpp.GroupChatType, msg2.Type())
}

func TestMessageType(t *testing.T) {
	message, _ := xmpp.NewMessageFromElement(xmpp.NewElementName("message"), &jid.JID{}, &jid.JID{})
	require.True(t, message.IsNormal())

	message.SetType(xmpp.NormalType)
	require.True(t, message.IsNormal())

	message.SetType(xmpp.HeadlineType)
	require.True(t, message.IsHeadline())

	message.SetType(xmpp.ChatType)
	require.True(t, message.IsChat())

	message.SetType(xmpp.GroupChatType)
	require.True(t, message.IsGroupChat())
}

func TestMessageJID(t *testing.T) {
	from, _ := jid.New("username", "test.org", "desktop", false)
	to, _ := jid.New("username", "example.org", "desktop", false)
	message, _ := xmpp.NewMessageFromElement(xmpp.NewElementName("message"), &jid.JID{}, &jid.JID{})
	message.SetFromJID(from)
	require.Equal(t, message.FromJID().String(), message.From())
	message.SetToJID(to)
	require.Equal(t, message.ToJID().String(), message.To())
}
