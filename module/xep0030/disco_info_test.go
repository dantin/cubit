package xep0030

import (
	"context"
	"crypto/tls"
	"testing"

	c2srouter "github.com/dantin/cubit/c2s/router"
	"github.com/dantin/cubit/module/xep0004"
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

type testDiscoInfoProvider struct {
}

func (tp *testDiscoInfoProvider) Identities(_ context.Context, toJID, fromJID *jid.JID, node string) []Identity {
	return []Identity{{Name: "test_identity"}}
}

func (tp *testDiscoInfoProvider) Items(_ context.Context, toJID, fromJID *jid.JID, node string) ([]Item, *xmpp.StanzaError) {
	return []Item{{Jid: "test.example.org"}}, nil
}

func (tp *testDiscoInfoProvider) Features(_ context.Context, toJID, fromJID *jid.JID, node string) ([]Feature, *xmpp.StanzaError) {
	return []Feature{"com.example.org.feature"}, nil
}

func (tp *testDiscoInfoProvider) Form(_ context.Context, toJID, fromJID *jid.JID, node string) (*xep0004.DataForm, *xmpp.StanzaError) {
	return nil, nil
}

func TestModule_XEP0030_Matching(t *testing.T) {
	r, rosterRep := setupTest("example.org")

	j, _ := jid.New("user", "example.org", "desktop", true)

	x := New(r, rosterRep)
	defer func() { _ = x.Shutdown() }()

	// test MatchesIQ
	iq1 := xmpp.NewIQType(uuid.New().String(), xmpp.GetType)
	iq1.SetFromJID(j)
	iq1.SetToJID(j.ToBareJID())

	require.False(t, x.MatchesIQ(iq1))

	iq1.AppendElement(xmpp.NewElementNamespace("query", discoItemsNamespace))

	iq2 := xmpp.NewIQType(uuid.New().String(), xmpp.GetType)
	iq2.SetFromJID(j)
	iq2.SetToJID(j.ToBareJID())
	iq2.AppendElement(xmpp.NewElementNamespace("query", discoItemsNamespace))

	require.True(t, x.MatchesIQ(iq1))
	require.True(t, x.MatchesIQ(iq2))

	iq1.SetType(xmpp.SetType)
	iq2.SetType(xmpp.ResultType)

	require.False(t, x.MatchesIQ(iq1))
	require.False(t, x.MatchesIQ(iq2))
}

func TestModule_XEP0030_SendFeatures(t *testing.T) {
	r, rosterRep := setupTest("example.org")

	j, _ := jid.New("user", "example.org", "desktop", true)
	srvJid, _ := jid.New("", "example.org", "", true)

	stm := stream.NewMockC2S(uuid.New().String(), j)
	stm.SetPresence(xmpp.NewPresence(j, j, xmpp.AvailableType))

	r.Bind(context.Background(), stm)

	x := New(r, rosterRep)
	defer func() { _ = x.Shutdown() }()

	x.RegisterServerFeature("s0")
	x.RegisterServerFeature("s1")
	x.RegisterServerFeature("s2")
	x.RegisterAccountFeature("a0")
	x.RegisterAccountFeature("a1")

	iq1 := xmpp.NewIQType(uuid.New().String(), xmpp.GetType)
	iq1.SetFromJID(j)
	iq1.SetToJID(srvJid)
	iq1.AppendElement(xmpp.NewElementNamespace("query", discoInfoNamespace))

	x.ProcessIQ(context.Background(), iq1)
	elem := stm.ReceiveElement()
	require.NotNil(t, elem)
	q := elem.Elements().ChildNamespace("query", discoInfoNamespace)

	require.NotNil(t, q)
	require.Len(t, q.Elements(), 6)
	require.Equal(t, "identity", q.Elements().All()[0].Name())
	require.Equal(t, "feature", q.Elements().All()[1].Name())

	x.UnregisterServerFeature("s1")
	x.UnregisterAccountFeature("a1")

	x.ProcessIQ(context.Background(), iq1)
	elem = stm.ReceiveElement()
	q = elem.Elements().ChildNamespace("query", discoInfoNamespace)

	require.NotNil(t, q)
	require.Len(t, q.Elements(), 5)

	iq1.SetToJID(j.ToBareJID())
	x.ProcessIQ(context.Background(), iq1)
	elem = stm.ReceiveElement()
	q = elem.Elements().ChildNamespace("query", discoInfoNamespace)

	require.NotNil(t, q)
	require.Len(t, q.Elements(), 4)
}

func TestModule_XEP0030_SendItems(t *testing.T) {
	r, rosterRep := setupTest("example.org")

	j, _ := jid.New("user", "example.org", "desktop", true)

	stm := stream.NewMockC2S(uuid.New().String(), j)
	stm.SetPresence(xmpp.NewPresence(j, j, xmpp.AvailableType))

	r.Bind(context.Background(), stm)

	x := New(r, rosterRep)
	defer func() { _ = x.Shutdown() }()

	iq1 := xmpp.NewIQType(uuid.New().String(), xmpp.GetType)
	iq1.SetFromJID(j)
	iq1.SetToJID(j.ToBareJID())
	iq1.AppendElement(xmpp.NewElementNamespace("query", discoItemsNamespace))

	x.ProcessIQ(context.Background(), iq1)
	elem := stm.ReceiveElement()
	require.NotNil(t, elem)
	q := elem.Elements().ChildNamespace("query", discoItemsNamespace)

	require.Len(t, q.Elements().Children("item"), 1)
}

func TestModule_XEP0030_Provider(t *testing.T) {
	r, rosterRep := setupTest("example.org")

	j, _ := jid.New("user", "example.org", "desktop", true)
	compJID, _ := jid.New("", "test.example.org", "", true)

	stm := stream.NewMockC2S(uuid.New().String(), j)
	stm.SetPresence(xmpp.NewPresence(j, j, xmpp.AvailableType))

	r.Bind(context.Background(), stm)

	x := New(r, rosterRep)
	defer func() { _ = x.Shutdown() }()

	iq1 := xmpp.NewIQType(uuid.New().String(), xmpp.GetType)
	iq1.SetFromJID(j)
	iq1.SetToJID(compJID)
	iq1.AppendElement(xmpp.NewElementNamespace("query", discoItemsNamespace))

	x.ProcessIQ(context.Background(), iq1)
	elem := stm.ReceiveElement()
	require.True(t, elem.IsError())
	require.Equal(t, xmpp.ErrItemNotFound.Error(), elem.Error().Elements().All()[0].Name())

	x.RegisterProvider(compJID.String(), &testDiscoInfoProvider{})

	x.ProcessIQ(context.Background(), iq1)
	elem = stm.ReceiveElement()
	q := elem.Elements().ChildNamespace("query", discoItemsNamespace)
	require.NotNil(t, q)

	require.Len(t, q.Elements().Children("item"), 1)

	x.UnregisterProvider(compJID.String())

	x.ProcessIQ(context.Background(), iq1)
	elem = stm.ReceiveElement()
	require.True(t, elem.IsError())
	require.Equal(t, xmpp.ErrItemNotFound.Error(), elem.Error().Elements().All()[0].Name())
}

func setupTest(domain string) (router.Router, repository.Roster) {
	hosts, _ := host.New([]host.Config{{Name: domain, Certificate: tls.Certificate{}}})
	rosterRep := memorystorage.NewRoster()
	r, _ := router.New(
		hosts,
		c2srouter.New(memorystorage.NewUser(), memorystorage.NewBlockList()),
		nil,
	)
	return r, rosterRep
}
