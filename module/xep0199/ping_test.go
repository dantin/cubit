package xep0199

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

func TestModule_XEP0199_Matching(t *testing.T) {
	j, _ := jid.New("user", "example.org", "desktop", true)

	x := New(&Config{}, nil, nil)
	defer func() { _ = x.Shutdown() }()

	// test MatchesIQ
	iqID := uuid.New().String()
	iq := xmpp.NewIQType(iqID, xmpp.GetType)
	iq.SetFromJID(j)

	ping := xmpp.NewElementNamespace("ping", pingNamespace)
	iq.AppendElement(ping)

	require.True(t, x.MatchesIQ(iq))
}

func TestModule_XEP0199_ReceivePing(t *testing.T) {
	r := setupTest()

	j1, _ := jid.New("alice", "example.org", "desktop", true)
	j2, _ := jid.New("bob", "example.org", "surface", true)

	stm := stream.NewMockC2S(uuid.New().String(), j1)
	r.Bind(context.Background(), stm)

	x := New(&Config{}, nil, r)
	defer func() { _ = x.Shutdown() }()

	iqID := uuid.New().String()
	iq := xmpp.NewIQType(iqID, xmpp.SetType)
	iq.SetFromJID(j1)
	iq.SetToJID(j2)

	x.ProcessIQ(context.Background(), iq)
	elem := stm.ReceiveElement()
	require.Equal(t, xmpp.ErrForbidden.Error(), elem.Error().Elements().All()[0].Name())

	iq.SetToJID(j1)
	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ErrBadRequest.Error(), elem.Error().Elements().All()[0].Name())

	ping := xmpp.NewElementNamespace("ping", pingNamespace)
	iq.AppendElement(ping)

	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ErrBadRequest.Error(), elem.Error().Elements().All()[0].Name())

	iq.SetType(xmpp.GetType)
	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, iqID, elem.ID())
}

func TestModule_XEP0199_SendPing(t *testing.T) {
	r := setupTest()

	j1, _ := jid.New("alice", "example.org", "desktop", true)
	j2, _ := jid.New("", "example.org", "", true)

	stm := stream.NewMockC2S(uuid.New().String(), j1)
	r.Bind(context.Background(), stm)

	x := New(&Config{Send: true, SendInterval: time.Second}, nil, r)
	defer func() { _ = x.Shutdown() }()

	x.SchedulePing(stm)

	// wait for ping...
	elem := stm.ReceiveElement()
	require.NotNil(t, elem)
	require.Equal(t, "iq", elem.Name())
	require.NotNil(t, elem.Elements().ChildNamespace("ping", pingNamespace))

	// send pong...
	pong := xmpp.NewIQType(elem.ID(), xmpp.ResultType)
	pong.SetFromJID(j1)
	pong.SetToJID(j2)
	x.ProcessIQ(context.Background(), pong)
	x.SchedulePing(stm)

	// wait next ping...
	elem = stm.ReceiveElement()
	require.NotNil(t, elem)
	require.Equal(t, "iq", elem.Name())
	require.NotNil(t, elem.Elements().ChildNamespace("ping", pingNamespace))

	// expect disconnection...
	err := stm.WaitDisconnection()
	require.NotNil(t, err)
}

func TestModule_XEP0199_Disconnect(t *testing.T) {
	r := setupTest()

	j1, _ := jid.New("alice", "example.org", "desktop", true)

	stm := stream.NewMockC2S(uuid.New().String(), j1)
	r.Bind(context.Background(), stm)

	x := New(&Config{Send: true, SendInterval: time.Second}, nil, r)
	defer func() { _ = x.Shutdown() }()

	x.SchedulePing(stm)

	// wait next ping...
	elem := stm.ReceiveElement()
	require.NotNil(t, elem)
	require.Equal(t, "iq", elem.Name())
	require.NotNil(t, elem.Elements().ChildNamespace("ping", pingNamespace))

	// expect disconnection...
	err := stm.WaitDisconnection()
	require.NotNil(t, err)
	require.Equal(t, "connection-timeout", err.Error())

}

func setupTest() router.Router {
	hosts, _ := host.New([]host.Config{{Name: "example.org", Certificate: tls.Certificate{}}})
	r, _ := router.New(
		hosts,
		c2srouter.New(memorystorage.NewUser(), memorystorage.NewBlockList()),
		nil,
	)
	return r

}
