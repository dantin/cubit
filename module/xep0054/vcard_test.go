package xep0054

import (
	"context"
	"crypto/tls"
	"testing"

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

func TestModule_XEP0054_Matching(t *testing.T) {
	j, _ := jid.New("user", "example.org", "desktop", true)

	x := New(nil, nil, nil)
	defer func() { _ = x.Shutdown() }()

	// test MatchesIQ
	iqID := uuid.New().String()
	iq := xmpp.NewIQType(iqID, xmpp.SetType)
	iq.SetFromJID(j)

	vCard := xmpp.NewElementNamespace("query", vCardNamespace)

	iq.AppendElement(xmpp.NewElementNamespace("query", "jabber:client"))
	require.False(t, x.MatchesIQ(iq))
	iq.ClearElements()
	iq.AppendElement(vCard)
	require.False(t, x.MatchesIQ(iq))
	iq.SetToJID(j.ToBareJID())
	require.False(t, x.MatchesIQ(iq))
	vCard.SetName("vCard")
	require.True(t, x.MatchesIQ(iq))
}

func TestModule_XEP0054_Set(t *testing.T) {
	r, s := setupTest("example.org")

	j, _ := jid.New("user", "example.org", "desktop", true)

	stm := stream.NewMockC2S("123", j)
	stm.SetPresence(xmpp.NewPresence(j, j, xmpp.AvailableType))

	r.Bind(context.Background(), stm)

	iqID := uuid.New().String()
	iq := xmpp.NewIQType(iqID, xmpp.SetType)
	iq.SetFromJID(j)
	iq.SetToJID(j.ToBareJID())
	iq.AppendElement(testVCard())

	x := New(nil, r, s)
	defer func() { _ = x.Shutdown() }()

	x.ProcessIQ(context.Background(), iq)
	elem := stm.ReceiveElement()
	require.NotNil(t, elem)
	require.Equal(t, xmpp.ResultType, elem.Type())
	require.Equal(t, iqID, elem.ID())

	// set empty vCard...
	iq2ID := uuid.New().String()
	iq2 := xmpp.NewIQType(iq2ID, xmpp.SetType)
	iq2.SetFromJID(j)
	iq2.SetToJID(j.ToBareJID())
	iq2.AppendElement(xmpp.NewElementNamespace("vCard", vCardNamespace))

	x.ProcessIQ(context.Background(), iq2)
	elem = stm.ReceiveElement()
	require.NotNil(t, elem)
	require.Equal(t, xmpp.ResultType, elem.Type())
	require.Equal(t, iq2ID, elem.ID())
}

func TestModule_XEP0054_SetError(t *testing.T) {
	r, s := setupTest("example.org")

	j, _ := jid.New("user", "example.org", "desktop", true)
	j2, _ := jid.New("alice", "example.org", "desktop", true)

	stm := stream.NewMockC2S("123", j)
	stm.SetPresence(xmpp.NewPresence(j, j, xmpp.AvailableType))

	r.Bind(context.Background(), stm)

	x := New(nil, r, s)
	defer func() { _ = x.Shutdown() }()

	// set other user vCard...
	iq := xmpp.NewIQType(uuid.New().String(), xmpp.SetType)
	iq.SetFromJID(j)
	iq.SetToJID(j2.ToBareJID())
	iq.AppendElement(testVCard())

	x.ProcessIQ(context.Background(), iq)
	elem := stm.ReceiveElement()
	require.Equal(t, xmpp.ErrForbidden.Error(), elem.Error().Elements().All()[0].Name())

	// storage error
	memorystorage.EnableMockedError()
	defer memorystorage.DisableMockedError()

	iq2 := xmpp.NewIQType(uuid.New().String(), xmpp.SetType)
	iq2.SetFromJID(j)
	iq2.SetToJID(j.ToBareJID())
	iq2.AppendElement(testVCard())

	x.ProcessIQ(context.Background(), iq2)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ErrInternalServerError.Error(), elem.Error().Elements().All()[0].Name())
}

func TestModule_XEP0054_Get(t *testing.T) {
	r, s := setupTest("example.org")

	j, _ := jid.New("user", "example.org", "desktop", true)
	j2, _ := jid.New("alice", "example.org", "desktop", true)

	stm := stream.NewMockC2S(uuid.New().String(), j)
	stm.SetPresence(xmpp.NewPresence(j, j, xmpp.AvailableType))

	r.Bind(context.Background(), stm)

	iqSet := xmpp.NewIQType(uuid.New().String(), xmpp.SetType)
	iqSet.SetFromJID(j)
	iqSet.SetToJID(j.ToBareJID())
	iqSet.AppendElement(testVCard())

	x := New(nil, r, s)
	defer func() { _ = x.Shutdown() }()

	x.ProcessIQ(context.Background(), iqSet)
	_ = stm.ReceiveElement() // wait until set...

	iqGetID := uuid.New().String()
	iqGet := xmpp.NewIQType(iqGetID, xmpp.GetType)
	iqGet.SetFromJID(j)
	iqGet.SetToJID(j.ToBareJID())
	iqGet.AppendElement(xmpp.NewElementNamespace("vCard", vCardNamespace))

	x.ProcessIQ(context.Background(), iqGet)
	elem := stm.ReceiveElement()
	require.NotNil(t, elem)
	vCard := elem.Elements().ChildNamespace("vCard", vCardNamespace)
	fn := vCard.Elements().Child("FN")
	require.Equal(t, "Forrest Gump", fn.Text())

	// non existing vCard...
	iqGet2ID := uuid.New().String()
	iqGet2 := xmpp.NewIQType(iqGet2ID, xmpp.GetType)
	iqGet2.SetFromJID(j)
	iqGet2.SetToJID(j2.ToBareJID())
	iqGet2.AppendElement(xmpp.NewElementNamespace("vCard", vCardNamespace))

	x.ProcessIQ(context.Background(), iqGet2)
	elem = stm.ReceiveElement()
	require.NotNil(t, elem)
	vCard = elem.Elements().ChildNamespace("vCard", vCardNamespace)
	require.Equal(t, 0, vCard.Elements().Count())
}

func TestModule_XEP0054_GetError(t *testing.T) {
	r, s := setupTest("example.org")

	j, _ := jid.New("user", "example.org", "desktop", true)

	stm := stream.NewMockC2S("id-123", j)
	stm.SetPresence(xmpp.NewPresence(j, j, xmpp.AvailableType))

	r.Bind(context.Background(), stm)

	iqSet := xmpp.NewIQType(uuid.New().String(), xmpp.SetType)
	iqSet.SetFromJID(j)
	iqSet.SetToJID(j.ToBareJID())
	iqSet.AppendElement(testVCard())

	x := New(nil, r, s)
	defer func() { _ = x.Shutdown() }()

	x.ProcessIQ(context.Background(), iqSet)
	_ = stm.ReceiveElement() // wait until set...

	iqGetID := uuid.New().String()
	iqGet := xmpp.NewIQType(iqGetID, xmpp.GetType)
	iqGet.SetFromJID(j)
	iqGet.SetToJID(j.ToBareJID())
	vCard := xmpp.NewElementNamespace("vCard", vCardNamespace)
	vCard.AppendElement(xmpp.NewElementName("FN"))
	iqGet.AppendElement(vCard)

	x.ProcessIQ(context.Background(), iqGet)
	elem := stm.ReceiveElement()
	require.Equal(t, xmpp.ErrBadRequest.Error(), elem.Error().Elements().All()[0].Name())

	iqGet2ID := uuid.New().String()
	iqGet2 := xmpp.NewIQType(iqGet2ID, xmpp.GetType)
	iqGet2.SetFromJID(j)
	iqGet2.SetToJID(j.ToBareJID())
	iqGet2.AppendElement(xmpp.NewElementNamespace("vCard", vCardNamespace))

	memorystorage.EnableMockedError()
	defer memorystorage.DisableMockedError()

	x.ProcessIQ(context.Background(), iqGet2)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ErrInternalServerError.Error(), elem.Error().Elements().All()[0].Name())
}

func testVCard() xmpp.XElement {
	vCard := xmpp.NewElementNamespace("vCard", vCardNamespace)
	fn := xmpp.NewElementName("FN")
	fn.SetText("Forrest Gump")
	org := xmpp.NewElementName("ORG")
	org.SetText("Bubba Gump Shrimp Co.")
	vCard.AppendElement(fn)
	vCard.AppendElement(org)
	return vCard
}

func setupTest(domain string) (router.Router, *memorystorage.VCard) {
	hosts, _ := host.New([]host.Config{{Name: domain, Certificate: tls.Certificate{}}})
	s := memorystorage.NewVCard()
	r, _ := router.New(
		hosts,
		c2srouter.New(memorystorage.NewUser(), memorystorage.NewBlockList()),
		nil,
	)
	return r, s
}
