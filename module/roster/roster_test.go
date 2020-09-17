package roster

import (
	"context"
	"crypto/tls"
	"testing"
	"time"

	c2srouter "github.com/dantin/cubit/c2s/router"
	"github.com/dantin/cubit/model"
	rostermodel "github.com/dantin/cubit/model/roster"
	"github.com/dantin/cubit/module/xep0115"
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

func TestModule_Roster_MatchesIQ(t *testing.T) {
	rtr, userRep, presencesRep, rosterRep := setupTest("example.org")

	r := New(&Config{}, xep0115.New(rtr, presencesRep, "id-123"), nil, rtr, userRep, rosterRep)
	defer func() { _ = r.Shutdown() }()

	iq := xmpp.NewIQType(uuid.New().String(), xmpp.GetType)
	iq.AppendElement(xmpp.NewElementNamespace("query", rosterNamespace))

	require.True(t, r.MatchesIQ(iq))
}

func TestModule_Roster_FetchRoster(t *testing.T) {
	rtr, userRep, presencesRep, rosterRep := setupTest("example.org")

	j1, _ := jid.New("user", "example.org", "desktop", true)

	stm := stream.NewMockC2S(uuid.New().String(), j1)
	rtr.Bind(context.Background(), stm)

	r := New(&Config{}, xep0115.New(rtr, presencesRep, "cap-123"), nil, rtr, userRep, rosterRep)
	defer func() { _ = r.Shutdown() }()

	iq := xmpp.NewIQType(uuid.New().String(), xmpp.ResultType)
	iq.SetFromJID(j1)
	iq.SetToJID(j1.ToBareJID())
	q := xmpp.NewElementNamespace("query", rosterNamespace)
	q.AppendElement(xmpp.NewElementName("q2"))
	iq.AppendElement(q)

	r.ProcessIQ(context.Background(), iq)
	elem := stm.ReceiveElement()
	require.Equal(t, xmpp.ErrBadRequest.Error(), elem.Error().Elements().All()[0].Name())

	iq.SetType(xmpp.GetType)
	r.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ErrBadRequest.Error(), elem.Error().Elements().All()[0].Name())
	q.ClearElements()

	r.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, "iq", elem.Name())
	require.Equal(t, xmpp.ResultType, elem.Type())

	query := elem.Elements().ChildNamespace("query", rosterNamespace)
	require.Len(t, query.Elements(), 0)

	ri1 := &rostermodel.Item{
		Username:     "user",
		JID:          "alice@example.org",
		Name:         "Alice",
		Subscription: rostermodel.SubscriptionNone,
		Ask:          true,
		Groups:       []string{"people", "coworker"},
	}
	_, _ = rosterRep.UpsertRosterItem(context.Background(), ri1)

	ri2 := &rostermodel.Item{
		Username:     "user",
		JID:          "bob@example.org",
		Name:         "Bob",
		Subscription: rostermodel.SubscriptionNone,
		Ask:          true,
		Groups:       []string{"others"},
	}
	_, _ = rosterRep.UpsertRosterItem(context.Background(), ri2)

	r = New(&Config{Versioning: true}, xep0115.New(rtr, nil, "cap-123"), nil, rtr, userRep, rosterRep)
	defer func() { _ = r.Shutdown() }()

	r.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, "iq", elem.Name())
	require.Equal(t, xmpp.ResultType, elem.Type())

	query2 := elem.Elements().ChildNamespace("query", rosterNamespace)
	require.Len(t, query2.Elements(), 2)

	requested, _ := stm.Value(rosterRequestedCtxKey).(bool)
	require.True(t, requested)

	// test versioning
	iq = xmpp.NewIQType(uuid.New().String(), xmpp.GetType)
	iq.SetFromJID(j1)
	iq.SetToJID(j1.ToBareJID())
	q = xmpp.NewElementNamespace("query", rosterNamespace)
	q.SetAttribute("ver", "v1")
	iq.AppendElement(q)

	r.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, "iq", elem.Name())
	require.Equal(t, xmpp.ResultType, elem.Type())

	// expect set item...
	elem = stm.ReceiveElement()
	require.Equal(t, "iq", elem.Name())
	require.Equal(t, xmpp.SetType, elem.Type())
	query2 = elem.Elements().ChildNamespace("query", rosterNamespace)
	require.Equal(t, "v2", query2.Attributes().Get("ver"))
	item := query2.Elements().Child("item")
	require.Equal(t, "bob@example.org", item.Attributes().Get("jid"))

	memorystorage.EnableMockedError()
	r = New(&Config{}, xep0115.New(rtr, nil, "cap-123"), nil, rtr, userRep, rosterRep)
	defer func() { _ = r.Shutdown() }()

	r.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ErrInternalServerError.Error(), elem.Error().Elements().All()[0].Name())
	memorystorage.DisableMockedError()
}

func TestModule_Roster_Update(t *testing.T) {
	rtr, userRep, presencesRep, rosterRep := setupTest("example.org")

	j1, _ := jid.New("user", "example.org", "desktop", true)
	j2, _ := jid.New("user", "example.org", "surface", true)

	stm1 := stream.NewMockC2S(uuid.New().String(), j1)
	stm1.SetAuthenticated(true)
	stm2 := stream.NewMockC2S(uuid.New().String(), j2)
	stm2.SetAuthenticated(true)
	stm2.SetValue(rosterRequestedCtxKey, true)

	r := New(&Config{}, xep0115.New(rtr, presencesRep, "cap-123"), nil, rtr, userRep, rosterRep)
	defer func() { _ = r.Shutdown() }()

	rtr.Bind(context.Background(), stm1)
	rtr.Bind(context.Background(), stm2)

	iqID := uuid.New().String()
	iq := xmpp.NewIQType(iqID, xmpp.SetType)
	iq.SetFromJID(j1)
	iq.SetToJID(j1.ToBareJID())
	q := xmpp.NewElementNamespace("query", rosterNamespace)
	item := xmpp.NewElementName("item")
	item.SetAttribute("jid", "alice@example.org")
	item.SetAttribute("subscription", rostermodel.SubscriptionNone)
	item.SetAttribute("name", "Alice")
	q.AppendElement(item)
	q.AppendElement(item)
	iq.AppendElement(q)

	r.ProcessIQ(context.Background(), iq)
	elem := stm1.ReceiveElement()
	require.Equal(t, xmpp.ErrBadRequest.Error(), elem.Error().Elements().All()[0].Name())

	q.ClearElements()
	q.AppendElement(item)

	r.ProcessIQ(context.Background(), iq)
	elem = stm1.ReceiveElement()
	require.Equal(t, "iq", elem.Name())
	require.Equal(t, xmpp.ResultType, elem.Type())
	require.Equal(t, iqID, elem.ID())

	// expecting roster push...
	elem = stm2.ReceiveElement()
	require.Equal(t, xmpp.SetType, elem.Type())

	// update name
	item.SetAttribute("name", "My Alice")
	q.ClearElements()
	q.AppendElement(item)

	r.ProcessIQ(context.Background(), iq)
	elem = stm1.ReceiveElement()
	require.Equal(t, "iq", elem.Name())
	require.Equal(t, xmpp.ResultType, elem.Type())
	require.Equal(t, iqID, elem.ID())

	ri, err := rosterRep.FetchRosterItem(context.Background(), "user", "alice@example.org")
	require.Nil(t, err)
	require.NotNil(t, ri)
	require.Equal(t, "user", ri.Username)
	require.Equal(t, "alice@example.org", ri.JID)
	require.Equal(t, "My Alice", ri.Name)
}

func TestModule_Roster_RemoveItem(t *testing.T) {
	rtr, userRep, presencesRep, rosterRep := setupTest("example.org")

	// insert contact's roster item
	_, _ = rosterRep.UpsertRosterItem(context.Background(), &rostermodel.Item{
		Username:     "alice",
		JID:          "bob@example.org",
		Name:         "My Bob",
		Subscription: rostermodel.SubscriptionBoth,
	})
	_, _ = rosterRep.UpsertRosterItem(context.Background(), &rostermodel.Item{
		Username:     "bob",
		JID:          "alice@example.org",
		Name:         "My Bob",
		Subscription: rostermodel.SubscriptionBoth,
	})
	j, _ := jid.New("alice", "example.org", "desktop", true)

	stm := stream.NewMockC2S(uuid.New().String(), j)
	stm.SetPresence(xmpp.NewPresence(j, j, xmpp.AvailableType))

	rtr.Bind(context.Background(), stm)

	r := New(&Config{}, xep0115.New(rtr, presencesRep, "cap-123"), nil, rtr, userRep, rosterRep)
	defer func() { _ = r.Shutdown() }()

	// remove item
	iqID := uuid.New().String()
	iq := xmpp.NewIQType(iqID, xmpp.SetType)
	iq.SetFromJID(j)
	iq.SetToJID(j.ToBareJID())

	q := xmpp.NewElementNamespace("query", rosterNamespace)
	item := xmpp.NewElementName("item")
	item.SetAttribute("jid", "bob@example.org")
	item.SetAttribute("subscription", rostermodel.SubscriptionRemove)
	q.AppendElement(item)
	iq.AppendElement(q)

	r.ProcessIQ(context.Background(), iq)
	elem := stm.ReceiveElement()
	require.Equal(t, iqID, elem.ID())

	ri, err := rosterRep.FetchRosterItem(context.Background(), "alice", "bob@example.org")
	require.Nil(t, err)
	require.Nil(t, ri)
}

func TestModule_Roster_OnlineJIDs(t *testing.T) {
	rtr, userRep, presencesRep, rosterRep := setupTest("example.org")

	j1, _ := jid.New("alice", "example.org", "desktop", true)
	j2, _ := jid.New("bob", "example.org", "desktop", true)
	j3, _ := jid.New("carl", "example.org", "desktop", true)
	j4, _ := jid.New("alice", "example.org", "surface", true)
	j5, _ := jid.New("boss", "jabber.org", "ipad", true)

	stm1 := stream.NewMockC2S(uuid.New().String(), j1)
	stm1.SetAuthenticated(true)
	stm2 := stream.NewMockC2S(uuid.New().String(), j2)
	stm2.SetAuthenticated(true)

	stm1.SetPresence(xmpp.NewPresence(j1, j1, xmpp.AvailableType))
	stm2.SetPresence(xmpp.NewPresence(j2, j2, xmpp.AvailableType))

	rtr.Bind(context.Background(), stm1)
	rtr.Bind(context.Background(), stm2)

	// user entity
	_ = userRep.UpsertUser(context.Background(), &model.User{
		Username:     "alice",
		LastPresence: xmpp.NewPresence(j1, j1.ToBareJID(), xmpp.UnavailableType),
	})

	// roster items
	_, _ = rosterRep.UpsertRosterItem(context.Background(), &rostermodel.Item{
		Username:     "bob",
		JID:          "alice@example.org",
		Subscription: rostermodel.SubscriptionBoth,
	})
	_, _ = rosterRep.UpsertRosterItem(context.Background(), &rostermodel.Item{
		Username:     "alice",
		JID:          "bob@example.org",
		Subscription: rostermodel.SubscriptionBoth,
	})

	// pending notification
	_ = rosterRep.UpsertRosterNotification(context.Background(), &rostermodel.Notification{
		Contact:  "alice",
		JID:      j3.ToBareJID().String(),
		Presence: xmpp.NewPresence(j3.ToBareJID(), j1.ToBareJID(), xmpp.SubscribeType),
	})

	ph := xep0115.New(rtr, presencesRep, "cap-123")
	r := New(&Config{}, ph, nil, rtr, userRep, rosterRep)
	defer func() { _ = r.Shutdown() }()

	// online presence...
	r.ProcessPresence(context.Background(), xmpp.NewPresence(j1, j1.ToBareJID(), xmpp.AvailableType))

	time.Sleep(time.Millisecond * 150) // wait until processed...

	// receive pending approval notification...
	elem := stm1.ReceiveElement()
	require.Equal(t, "presence", elem.Name())
	require.Equal(t, j3.ToBareJID().String(), elem.From())
	require.Equal(t, xmpp.SubscribeType, elem.Type())

	// expect user's available presence
	elem = stm2.ReceiveElement()
	require.Equal(t, "presence", elem.Name())
	require.Equal(t, j1.String(), elem.From())
	require.Equal(t, xmpp.AvailableType, elem.Type())

	// check if last presence was updated
	usr, err := userRep.FetchUser(context.Background(), "alice")
	require.Nil(t, err)
	require.NotNil(t, usr)
	require.NotNil(t, usr.LastPresence)
	require.Equal(t, xmpp.AvailableType, usr.LastPresence.Type())

	// send remaining online presences...
	r.ProcessPresence(context.Background(), xmpp.NewPresence(j2, j2.ToBareJID(), xmpp.AvailableType))
	r.ProcessPresence(context.Background(), xmpp.NewPresence(j3, j3.ToBareJID(), xmpp.AvailableType))
	r.ProcessPresence(context.Background(), xmpp.NewPresence(j4, j1.ToBareJID(), xmpp.AvailableType))
	r.ProcessPresence(context.Background(), xmpp.NewPresence(j5, j1.ToBareJID(), xmpp.AvailableType))

	time.Sleep(time.Millisecond * 150) // wait until processed...

	ln1, _ := ph.PresencesMatchingJID(context.Background(), j1)
	require.Len(t, ln1, 1)

	j6, _ := jid.NewWithString("example.org", true)
	ln6, _ := ph.PresencesMatchingJID(context.Background(), j6)
	require.Len(t, ln6, 4)

	j7, _ := jid.NewWithString("jabber.org", true)
	ln7, _ := ph.PresencesMatchingJID(context.Background(), j7)
	require.Len(t, ln7, 1)

	j8, _ := jid.NewWithString("example.org/desktop", true)
	ln8, _ := ph.PresencesMatchingJID(context.Background(), j8)
	require.Len(t, ln8, 3)

	j9, _ := jid.NewWithString("alice@example.org", true)
	ln9, _ := ph.PresencesMatchingJID(context.Background(), j9)
	require.Len(t, ln9, 2)

	// send unavailable presences...
	r.ProcessPresence(context.Background(), xmpp.NewPresence(j1, j1.ToBareJID(), xmpp.UnavailableType))
	r.ProcessPresence(context.Background(), xmpp.NewPresence(j2, j2.ToBareJID(), xmpp.UnavailableType))
	r.ProcessPresence(context.Background(), xmpp.NewPresence(j3, j3.ToBareJID(), xmpp.UnavailableType))
	r.ProcessPresence(context.Background(), xmpp.NewPresence(j4, j4.ToBareJID(), xmpp.UnavailableType))
	r.ProcessPresence(context.Background(), xmpp.NewPresence(j5, j1.ToBareJID(), xmpp.UnavailableType))

	time.Sleep(time.Millisecond * 150) // wait until processed...

	ln1, _ = ph.PresencesMatchingJID(context.Background(), j1)
	ln6, _ = ph.PresencesMatchingJID(context.Background(), j6)
	ln7, _ = ph.PresencesMatchingJID(context.Background(), j7)
	ln8, _ = ph.PresencesMatchingJID(context.Background(), j8)
	ln9, _ = ph.PresencesMatchingJID(context.Background(), j9)
	require.Len(t, ln1, 0)
	require.Len(t, ln6, 0)
	require.Len(t, ln7, 0)
	require.Len(t, ln8, 0)
	require.Len(t, ln9, 0)
}

func TestModule_Roster_Probe(t *testing.T) {
	rtr, userRep, presencesRep, rosterRep := setupTest("example.org")

	j1, _ := jid.New("alice", "example.org", "desktop", true)
	j2, _ := jid.New("bob", "example.org", "desktop", true)

	stm := stream.NewMockC2S(uuid.New().String(), j1)
	stm.SetAuthenticated(true)

	stm.SetPresence(xmpp.NewPresence(j1, j1, xmpp.AvailableType))

	rtr.Bind(context.Background(), stm)

	r := New(&Config{}, xep0115.New(rtr, presencesRep, "cap-123"), nil, rtr, userRep, rosterRep)
	defer func() { _ = r.Shutdown() }()

	_ = userRep.UpsertUser(context.Background(), &model.User{
		Username:     "bob",
		LastPresence: xmpp.NewPresence(j2.ToBareJID(), j2.ToBareJID(), xmpp.UnavailableType),
	})

	_, _ = rosterRep.UpsertRosterItem(context.Background(), &rostermodel.Item{
		Username:     "bob",
		JID:          "alice@example.org",
		Subscription: rostermodel.SubscriptionFrom,
	})
	r.ProcessPresence(context.Background(), xmpp.NewPresence(j1, j2, xmpp.ProbeType))
	elem := stm.ReceiveElement()
	require.Equal(t, xmpp.UnavailableType, elem.Type())

	// test available presence...
	p2 := xmpp.NewPresence(j2, j2.ToBareJID(), xmpp.AvailableType)
	_ = userRep.UpsertUser(context.Background(), &model.User{
		Username:     "bob",
		LastPresence: p2,
	})
	r.ProcessPresence(context.Background(), xmpp.NewPresence(j1, j2, xmpp.ProbeType))
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.AvailableType, elem.Type())
	require.Equal(t, "bob@example.org/desktop", elem.From())
}

func TestModule_Roster_Subscription(t *testing.T) {
	rtr, userRep, presencesRep, rosterRep := setupTest("example.org")

	j1, _ := jid.New("alice", "example.org", "desktop", true)
	j2, _ := jid.New("bob", "example.org", "desktop", true)

	r := New(&Config{}, xep0115.New(rtr, presencesRep, "cap-123"), nil, rtr, userRep, rosterRep)
	defer func() { _ = r.Shutdown() }()

	r.ProcessPresence(context.Background(), xmpp.NewPresence(j1.ToBareJID(), j2.ToBareJID(), xmpp.SubscribeType))
	time.Sleep(time.Millisecond * 150) // wait until processed...

	rns, err := rosterRep.FetchRosterNotifications(context.Background(), "bob")
	require.Nil(t, err)
	require.Len(t, rns, 1)

	// resend request...
	r.ProcessPresence(context.Background(), xmpp.NewPresence(j1.ToBareJID(), j2.ToBareJID(), xmpp.SubscribeType))

	// contact request cancellation
	r.ProcessPresence(context.Background(), xmpp.NewPresence(j2.ToBareJID(), j1.ToBareJID(), xmpp.UnsubscribedType))
	time.Sleep(time.Millisecond * 150) // wait until processed...

	rns, err = rosterRep.FetchRosterNotifications(context.Background(), "bob")
	require.Nil(t, err)
	require.Len(t, rns, 0)

	ri, err := rosterRep.FetchRosterItem(context.Background(), "alice", "bob@example.org")
	require.Nil(t, err)
	require.Equal(t, rostermodel.SubscriptionNone, ri.Subscription)

	// contact accepts request...
	r.ProcessPresence(context.Background(), xmpp.NewPresence(j1.ToBareJID(), j2.ToBareJID(), xmpp.SubscribeType))
	r.ProcessPresence(context.Background(), xmpp.NewPresence(j2.ToBareJID(), j1.ToBareJID(), xmpp.SubscribedType))
	time.Sleep(time.Millisecond * 150) // wait until processed...

	ri, err = rosterRep.FetchRosterItem(context.Background(), "alice", "bob@example.org")
	require.Nil(t, err)
	require.Equal(t, rostermodel.SubscriptionTo, ri.Subscription)

	// contact subscribes to user's presence...
	r.ProcessPresence(context.Background(), xmpp.NewPresence(j2.ToBareJID(), j1.ToBareJID(), xmpp.SubscribeType))
	r.ProcessPresence(context.Background(), xmpp.NewPresence(j1.ToBareJID(), j2.ToBareJID(), xmpp.SubscribedType))
	time.Sleep(time.Millisecond * 150) // wait until processed...

	ri, err = rosterRep.FetchRosterItem(context.Background(), "bob", "alice@example.org")
	require.Nil(t, err)
	require.Equal(t, rostermodel.SubscriptionBoth, ri.Subscription)

	// user unsubscribes from contact's presence...
	r.ProcessPresence(context.Background(), xmpp.NewPresence(j1.ToBareJID(), j2.ToBareJID(), xmpp.UnsubscribeType))
	time.Sleep(time.Millisecond * 150) // wait until processed...

	ri, err = rosterRep.FetchRosterItem(context.Background(), "alice", "bob@example.org")
	require.Nil(t, err)
	require.Equal(t, rostermodel.SubscriptionFrom, ri.Subscription)

	// user cancels contact subscription
	r.ProcessPresence(context.Background(), xmpp.NewPresence(j1.ToBareJID(), j2.ToBareJID(), xmpp.UnsubscribedType))
	time.Sleep(time.Millisecond * 150) // wait until processed...

	ri, err = rosterRep.FetchRosterItem(context.Background(), "alice", "bob@example.org")
	require.Nil(t, err)
	require.Equal(t, rostermodel.SubscriptionNone, ri.Subscription)

	ri, err = rosterRep.FetchRosterItem(context.Background(), "bob", "alice@example.org")
	require.Nil(t, err)
	require.Equal(t, rostermodel.SubscriptionNone, ri.Subscription)
}

func setupTest(domain string) (router.Router, repository.User, repository.Presences, repository.Roster) {
	hosts, _ := host.New([]host.Config{{Name: domain, Certificate: tls.Certificate{}}})

	userRep := memorystorage.NewUser()
	presencesRep := memorystorage.NewPresences()
	rosterRep := memorystorage.NewRoster()
	r, _ := router.New(
		hosts,
		c2srouter.New(userRep, memorystorage.NewBlockList()),
		nil,
	)
	return r, userRep, presencesRep, rosterRep
}
