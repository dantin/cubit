package xep0049

import (
	"context"
	"crypto/tls"
	"testing"

	c2srouter "github.com/dantin/cubit/c2s/router"
	"github.com/dantin/cubit/router"
	"github.com/dantin/cubit/router/host"
	memorystorage "github.com/dantin/cubit/storage/memory"
	"github.com/dantin/cubit/storage/repository"
	"github.com/dantin/cubit/stream"
	"github.com/dantin/cubit/xmpp"
	"github.com/dantin/cubit/xmpp/jid"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestModule_XEP0049_Matching(t *testing.T) {
	j1, _ := jid.New("user", "example.org", "desktop", true)
	j2, _ := jid.New("alice", "example.org", "desktop", true)

	stm := stream.NewMockC2S("s-123", j1)
	defer stm.Disconnect(context.Background(), nil)

	x := New(nil, nil)
	defer func() { _ = x.Shutdown() }()

	iq := xmpp.NewIQType(uuid.New().String(), xmpp.GetType)
	iq.SetFromJID(j1)
	iq.SetToJID(j2.ToBareJID())
	require.False(t, x.MatchesIQ(iq))

	iq.AppendElement(xmpp.NewElementNamespace("query", privateNamespace))
	require.True(t, x.MatchesIQ(iq))
}

func TestModule_XEP0049_InvalidIQ(t *testing.T) {
	r, s := setupTest("example.org")

	j1, _ := jid.New("user", "example.org", "desktop", true)
	j2, _ := jid.New("alice", "example.org", "desktop", true)

	stm := stream.NewMockC2S("s-123", j1)
	stm.SetPresence(xmpp.NewPresence(j1, j1, xmpp.AvailableType))

	r.Bind(context.Background(), stm)

	x := New(r, s)
	defer func() { _ = x.Shutdown() }()

	iq := xmpp.NewIQType(uuid.New().String(), xmpp.GetType)
	iq.SetFromJID(j1)
	iq.SetToJID(j2.ToBareJID())
	q := xmpp.NewElementNamespace("query", privateNamespace)
	iq.AppendElement(q)

	x.ProcessIQ(context.Background(), iq)
	elem := stm.ReceiveElement()
	require.Equal(t, xmpp.ErrForbidden.Error(), elem.Error().Elements().All()[0].Name())

	iq.SetType(xmpp.ResultType)
	iq.SetToJID(j1.ToBareJID())
	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ErrBadRequest.Error(), elem.Error().Elements().All()[0].Name())

	iq.SetType(xmpp.GetType)
	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ErrNotAcceptable.Error(), elem.Error().Elements().All()[0].Name())

	exodus := xmpp.NewElementNamespace("exodus", "exodus:ns")
	exodus.AppendElement(xmpp.NewElementName("exodus2"))
	q.AppendElement(exodus)
	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ErrNotAcceptable.Error(), elem.Error().Elements().All()[0].Name())

	exodus.ClearElements()
	exodus.SetNamespace("jabber:client")
	iq.SetType(xmpp.SetType)
	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ErrNotAcceptable.Error(), elem.Error().Elements().All()[0].Name())

	exodus.SetNamespace("")
	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ErrBadRequest.Error(), elem.Error().Elements().All()[0].Name())
}

func TestModule_XEP0049_SetAndGetPrivate(t *testing.T) {
	r, s := setupTest("example.org")

	j, _ := jid.New("user", "example.org", "desktop", true)

	stm := stream.NewMockC2S("s-123", j)
	stm.SetPresence(xmpp.NewPresence(j, j, xmpp.AvailableType))

	r.Bind(context.Background(), stm)

	x := New(r, s)
	defer func() { _ = x.Shutdown() }()

	iqID := uuid.New().String()
	iq := xmpp.NewIQType(iqID, xmpp.SetType)
	iq.SetFromJID(j)
	iq.SetToJID(j.ToBareJID())
	q := xmpp.NewElementNamespace("query", privateNamespace)
	iq.AppendElement(q)

	exodus1 := xmpp.NewElementNamespace("exodus1", "exodus:ns")
	exodus2 := xmpp.NewElementNamespace("exodus2", "exodus:ns")
	q.AppendElement(exodus1)
	q.AppendElement(exodus2)

	// set error
	memorystorage.EnableMockedError()
	x.ProcessIQ(context.Background(), iq)
	elem := stm.ReceiveElement()
	require.Equal(t, xmpp.ErrInternalServerError.Error(), elem.Error().Elements().All()[0].Name())
	memorystorage.DisableMockedError()

	// set success
	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ResultType, elem.Type())
	require.Equal(t, iqID, elem.ID())

	// get error
	q.RemoveElements("exodus2")
	iq.SetType(xmpp.GetType)

	memorystorage.EnableMockedError()
	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ErrInternalServerError.Error(), elem.Error().Elements().All()[0].Name())
	memorystorage.DisableMockedError()

	// get success
	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ResultType, elem.Type())
	require.Equal(t, iqID, elem.ID())

	q2 := elem.Elements().ChildNamespace("query", privateNamespace)
	require.Len(t, q2.Elements(), 2)
	require.Equal(t, "exodus:ns", q2.Elements().All()[0].Namespace())

	// get non existing
	exodus1.SetNamespace("exodus:ns:2")
	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ResultType, elem.Type())
	require.Equal(t, iqID, elem.ID())
	q3 := elem.Elements().ChildNamespace("query", privateNamespace)
	require.Len(t, q3.Elements(), 1)
	require.Equal(t, "exodus:ns:2", q3.Elements().All()[0].Namespace())
}

func setupTest(domain string) (router.Router, repository.Private) {
	hosts, _ := host.New([]host.Config{{Name: domain, Certificate: tls.Certificate{}}})
	s := memorystorage.NewPrivate()
	r, _ := router.New(
		hosts,
		c2srouter.New(memorystorage.NewUser(), memorystorage.NewBlockList()),
		nil,
	)
	return r, s
}
