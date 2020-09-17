package offline

import (
	"context"
	"crypto/tls"
	"testing"
	"time"

	c2srouter "github.com/dantin/cubit/c2s/router"
	"github.com/dantin/cubit/router"
	"github.com/dantin/cubit/router/host"
	memorystorage "github.com/dantin/cubit/storage/memory"
	"github.com/dantin/cubit/stream"
	"github.com/dantin/cubit/xmpp"
	"github.com/dantin/cubit/xmpp/jid"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestModule_Offline_ArchiveMessage(t *testing.T) {
	r, s := setupTest("example.org")

	j1, _ := jid.New("alice", "example.org", "desktop", true)
	j2, _ := jid.New("bob", "example.org", "surface", true)

	stm := stream.NewMockC2S(uuid.New().String(), j1)
	stm.SetPresence(xmpp.NewPresence(j1, j1, xmpp.AvailableType))

	r.Bind(context.Background(), stm)

	x := New(&Config{QueueSize: 1}, nil, r, s)
	defer func() { _ = x.Shutdown() }()

	msgID := uuid.New().String()
	msg := xmpp.NewMessageType(msgID, "normal")
	msg.SetFromJID(j1)
	msg.SetToJID(j2)
	x.ArchiveMessage(context.Background(), msg)

	// wait for insertion...
	time.Sleep(time.Millisecond * 250)

	msgs, err := s.FetchOfflineMessages(context.Background(), "bob")
	require.Nil(t, err)
	require.Len(t, msgs, 1)

	msg2 := xmpp.NewMessageType(msgID, "normal")
	msg2.SetFromJID(j1)
	msg2.SetToJID(j2)

	x.ArchiveMessage(context.Background(), msg)

	elem := stm.ReceiveElement()
	require.NotNil(t, elem)
	require.Equal(t, xmpp.ErrServiceUnavailable.Error(), elem.Error().Elements().All()[0].Name())

	// deliver offline messages...
	stm2 := stream.NewMockC2S("abcd", j2)
	stm2.SetPresence(xmpp.NewPresence(j2, j2, xmpp.AvailableType))

	r.Bind(context.Background(), stm2)

	x2 := New(&Config{QueueSize: 1}, nil, r, s)
	defer func() { _ = x.Shutdown() }()

	x2.DeliverOfflineMessages(context.Background(), stm2)

	elem = stm2.ReceiveElement()
	require.NotNil(t, elem)
	require.Equal(t, msgID, elem.ID())
}

func setupTest(domain string) (router.Router, *memorystorage.Offline) {
	hosts, _ := host.New([]host.Config{{Name: domain, Certificate: tls.Certificate{}}})

	s := memorystorage.NewOffline()
	r, _ := router.New(
		hosts,
		c2srouter.New(memorystorage.NewUser(), memorystorage.NewBlockList()),
		nil,
	)
	return r, s
}
