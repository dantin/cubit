package memorystorage

import (
	"context"
	"testing"

	"github.com/dantin/cubit/xmpp"
	"github.com/dantin/cubit/xmpp/jid"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage_InsertOfflineMessage(t *testing.T) {
	j, _ := jid.NewWithString("alice@example.org/desktop", false)
	message := xmpp.NewElementName("message")
	message.SetID(uuid.New().String())
	message.AppendElement(xmpp.NewElementName("body"))
	m, _ := xmpp.NewMessageFromElement(message, j, j)

	s := NewOffline()
	EnableMockedError()
	require.Equal(t, ErrMocked, s.InsertOfflineMessage(context.Background(), m, "alice"))
	DisableMockedError()

	require.Nil(t, s.InsertOfflineMessage(context.Background(), m, "alice"))
}

func TestMemoryStorage_CountOfflineMessages(t *testing.T) {
	j, _ := jid.NewWithString("alice@example.org/desktop", false)
	message := xmpp.NewElementName("message")
	message.SetID(uuid.New().String())
	message.AppendElement(xmpp.NewElementName("body"))
	m, _ := xmpp.NewMessageFromElement(message, j, j)

	s := NewOffline()

	_ = s.InsertOfflineMessage(context.Background(), m, "alice")

	EnableMockedError()
	_, err := s.CountOfflineMessages(context.Background(), "alice")
	require.Equal(t, ErrMocked, err)
	DisableMockedError()

	cnt, _ := s.CountOfflineMessages(context.Background(), "alice")
	require.Equal(t, 1, cnt)
}

func TestMemoryStorage_FetchOfflineMessages(t *testing.T) {
	j, _ := jid.NewWithString("alice@example.org/desktop", false)
	message := xmpp.NewElementName("message")
	message.SetID(uuid.New().String())
	message.AppendElement(xmpp.NewElementName("body"))
	m, _ := xmpp.NewMessageFromElement(message, j, j)

	s := NewOffline()
	_ = s.InsertOfflineMessage(context.Background(), m, "alice")

	EnableMockedError()
	_, err := s.FetchOfflineMessages(context.Background(), "alice")
	require.Equal(t, ErrMocked, err)
	DisableMockedError()

	elems, err := s.FetchOfflineMessages(context.Background(), "alice")
	require.Len(t, elems, 1)
}

func TestMemoryStorage_DeleteOfflineMessages(t *testing.T) {
	j, _ := jid.NewWithString("alice@example.org/desktop", false)
	message := xmpp.NewElementName("message")
	message.SetID(uuid.New().String())
	message.AppendElement(xmpp.NewElementName("body"))
	m, _ := xmpp.NewMessageFromElement(message, j, j)

	s := NewOffline()
	_ = s.InsertOfflineMessage(context.Background(), m, "alice")

	EnableMockedError()
	require.Equal(t, ErrMocked, s.DeleteOfflineMessages(context.Background(), "alice"))
	DisableMockedError()

	require.Nil(t, s.DeleteOfflineMessages(context.Background(), "alice"))

	elems, _ := s.FetchOfflineMessages(context.Background(), "alice")
	require.Len(t, elems, 0)
}
