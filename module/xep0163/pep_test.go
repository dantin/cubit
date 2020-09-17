package xep0163

import (
	"context"
	"crypto/tls"
	"testing"

	c2srouter "github.com/dantin/cubit/c2s/router"
	capsmodel "github.com/dantin/cubit/model/capabilities"
	pubsubmodel "github.com/dantin/cubit/model/pubsub"
	rostermodel "github.com/dantin/cubit/model/roster"
	"github.com/dantin/cubit/module/xep0004"
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

func TestModule_XEP0163_Matching(t *testing.T) {
	r, _, rosterRep, pubSubRep := setupTest("example.org")

	j, _ := jid.New("user", "example.org", "desktop", true)

	stm := stream.NewMockC2S(uuid.New().String(), j)
	stm.SetPresence(xmpp.NewPresence(j, j, xmpp.AvailableType))

	r.Bind(context.Background(), stm)

	p := New(nil, nil, r, rosterRep, pubSubRep)
	defer func() { _ = p.Shutdown() }()

	// test MatchesIQ
	iq := xmpp.NewIQType(uuid.New().String(), xmpp.GetType)
	iq.SetFromJID(j)
	iq.SetToJID(j.ToBareJID())
	iq.AppendElement(xmpp.NewElementNamespace("pubsub", pubSubOwnerNamespace))

	require.True(t, p.MatchesIQ(iq))
}

func TestModule_XEP0163_CreateNode(t *testing.T) {
	r, _, rosterRep, pubSubRep := setupTest("example.org")

	j, _ := jid.New("user", "example.org", "desktop", true)

	stm := stream.NewMockC2S(uuid.New().String(), j)
	stm.SetPresence(xmpp.NewPresence(j, j, xmpp.AvailableType))

	r.Bind(context.Background(), stm)

	p := New(nil, nil, r, rosterRep, pubSubRep)
	defer func() { _ = p.Shutdown() }()

	iqID := uuid.New().String()
	iq := xmpp.NewIQType(iqID, xmpp.SetType)
	iq.SetFromJID(j)
	iq.SetToJID(j.ToBareJID())

	pubSub := xmpp.NewElementNamespace("pubsub", pubSubNamespace)
	create := xmpp.NewElementName("create")
	create.SetAttribute("node", "current_status")
	pubSub.AppendElement(create)
	iq.AppendElement(pubSub)

	p.ProcessIQ(context.Background(), iq)
	elem := stm.ReceiveElement()
	require.NotNil(t, elem)
	require.Equal(t, iqID, elem.ID())
	require.Equal(t, xmpp.ResultType, elem.Type())

	// read node
	n, _ := pubSubRep.FetchNode(context.Background(), "user@example.org", "current_status")
	require.NotNil(t, n)
	require.Equal(t, n.Options, defaultNodeOptions)
}

func TestModule_XEP0163_GetNodeConfiguration(t *testing.T) {
	r, _, rosterRep, pubSubRep := setupTest("example.org")

	j, _ := jid.New("user", "example.org", "desktop", true)

	stm := stream.NewMockC2S(uuid.New().String(), j)
	stm.SetPresence(xmpp.NewPresence(j, j, xmpp.AvailableType))

	r.Bind(context.Background(), stm)

	_ = pubSubRep.UpsertNode(context.Background(), &pubsubmodel.Node{
		Host:    "user@example.org",
		Name:    "current_status",
		Options: defaultNodeOptions,
	})

	_ = pubSubRep.UpsertNodeAffiliation(context.Background(), &pubsubmodel.Affiliation{
		JID:         "user@example.org",
		Affiliation: pubsubmodel.Owner,
	}, "user@example.org", "current_status")

	p := New(nil, nil, r, rosterRep, pubSubRep)
	defer func() { _ = p.Shutdown() }()

	iqID := uuid.New().String()
	iq := xmpp.NewIQType(iqID, xmpp.GetType)
	iq.SetFromJID(j)
	iq.SetToJID(j.ToBareJID())

	pubSub := xmpp.NewElementNamespace("pubsub", pubSubOwnerNamespace)
	configureElem := xmpp.NewElementName("configure")
	configureElem.SetAttribute("node", "current_status")
	pubSub.AppendElement(configureElem)
	iq.AppendElement(pubSub)

	p.ProcessIQ(context.Background(), iq)
	elem := stm.ReceiveElement()
	require.NotNil(t, elem)
	require.Equal(t, iqID, elem.ID())
	require.Equal(t, xmpp.ResultType, elem.Type())

	// get form element
	pubSubRes := elem.Elements().ChildNamespace("pubsub", pubSubOwnerNamespace)
	require.NotNil(t, pubSubRes)
	configElem := pubSubRes.Elements().Child("configure")
	require.NotNil(t, configElem)
	formEl := configElem.Elements().ChildNamespace("x", xep0004.FormNamespace)
	require.NotNil(t, formEl)

	configForm, err := xep0004.NewFormFromElement(formEl)
	require.Nil(t, err)
	require.Equal(t, xep0004.Form, configForm.Type)
}

func TestModule_XEP0163_SetNodeConfiguration(t *testing.T) {
	r, _, rosterRep, pubSubRep := setupTest("example.org")

	j1, _ := jid.New("alice", "example.org", "desktop", true)
	j2, _ := jid.New("bob", "example.org", "desktop", true)

	stm1 := stream.NewMockC2S(uuid.New().String(), j1)
	stm2 := stream.NewMockC2S(uuid.New().String(), j2)
	stm1.SetPresence(xmpp.NewPresence(j1, j1, xmpp.AvailableType))
	stm2.SetPresence(xmpp.NewPresence(j2, j2, xmpp.AvailableType))

	r.Bind(context.Background(), stm1)
	r.Bind(context.Background(), stm2)

	nodeOpts := defaultNodeOptions
	nodeOpts.NotifyConfig = true

	// create node and affiliations
	_ = pubSubRep.UpsertNode(context.Background(), &pubsubmodel.Node{
		Host:    "alice@example.org",
		Name:    "current_status",
		Options: nodeOpts,
	})

	_ = pubSubRep.UpsertNodeAffiliation(context.Background(), &pubsubmodel.Affiliation{
		JID:         "alice@example.org",
		Affiliation: pubsubmodel.Owner,
	}, "alice@example.org", "current_status")

	_ = pubSubRep.UpsertNodeSubscription(context.Background(), &pubsubmodel.Subscription{
		JID:          "alice@example.org",
		Subscription: pubsubmodel.Subscribed,
	}, "alice@example.org", "current_status")

	_ = pubSubRep.UpsertNodeSubscription(context.Background(), &pubsubmodel.Subscription{
		JID:          "bob@example.org",
		Subscription: pubsubmodel.Subscribed,
	}, "alice@example.org", "current_status")

	_, _ = rosterRep.UpsertRosterItem(context.Background(), &rostermodel.Item{
		Username:     "alice",
		JID:          "bob@example.org",
		Subscription: "both",
	})

	// process pubsub command
	p := New(nil, nil, r, rosterRep, pubSubRep)
	defer func() { _ = p.Shutdown() }()

	iqID := uuid.New().String()
	iq := xmpp.NewIQType(iqID, xmpp.SetType)
	iq.SetFromJID(j1)
	iq.SetToJID(j1.ToBareJID())

	pubSub := xmpp.NewElementNamespace("pubsub", pubSubOwnerNamespace)
	configureElem := xmpp.NewElementName("configure")
	configureElem.SetAttribute("node", "current_status")

	// attach config update
	nodeOpts.Title = "a fancy new title"

	configForm := nodeOpts.ResultForm()
	configForm.Type = xep0004.Submit
	configureElem.AppendElement(configForm.Element())

	pubSub.AppendElement(configureElem)
	iq.AppendElement(pubSub)

	p.ProcessIQ(context.Background(), iq)

	elem := stm1.ReceiveElement()
	require.NotNil(t, elem)
	require.Equal(t, "message", elem.Name()) // notification
	require.NotNil(t, elem.Elements().ChildNamespace("event", pubSubEventNamespace))

	elem2 := stm2.ReceiveElement()
	require.NotNil(t, elem2)
	require.Equal(t, "message", elem.Name()) // notification
	eventElem := elem2.Elements().ChildNamespace("event", pubSubEventNamespace)
	require.NotNil(t, eventElem)

	configElemResp := eventElem.Elements().Child("configuration")
	require.NotNil(t, configElemResp)
	require.Equal(t, "current_status", configElemResp.Attributes().Get("node"))

	// result IQ
	elem = stm1.ReceiveElement()
	require.NotNil(t, elem)
	require.Equal(t, iqID, elem.ID())
	require.Equal(t, xmpp.ResultType, elem.Type())

	// check if configuration was applied
	n, _ := pubSubRep.FetchNode(context.Background(), "alice@example.org", "current_status")
	require.NotNil(t, n)
	require.Equal(t, nodeOpts.Title, n.Options.Title)
}

func TestModule_XEP0163_DeleteNode(t *testing.T) {
	r, _, rosterRep, pubSubRep := setupTest("example.org")

	j1, _ := jid.New("alice", "example.org", "desktop", true)
	j2, _ := jid.New("bob", "example.org", "desktop", true)

	stm1 := stream.NewMockC2S(uuid.New().String(), j1)
	stm2 := stream.NewMockC2S(uuid.New().String(), j2)
	stm1.SetPresence(xmpp.NewPresence(j1, j1, xmpp.AvailableType))
	stm2.SetPresence(xmpp.NewPresence(j2, j2, xmpp.AvailableType))

	r.Bind(context.Background(), stm1)
	r.Bind(context.Background(), stm2)

	nodeOpts := defaultNodeOptions
	nodeOpts.NotifyDelete = true

	// create node and affiliations
	_ = pubSubRep.UpsertNode(context.Background(), &pubsubmodel.Node{
		Host:    "alice@example.org",
		Name:    "current_status",
		Options: nodeOpts,
	})

	_ = pubSubRep.UpsertNodeAffiliation(context.Background(), &pubsubmodel.Affiliation{
		JID:         "alice@example.org",
		Affiliation: pubsubmodel.Owner,
	}, "alice@example.org", "current_status")

	_ = pubSubRep.UpsertNodeSubscription(context.Background(), &pubsubmodel.Subscription{
		JID:          "alice@example.org",
		Subscription: pubsubmodel.Subscribed,
	}, "alice@example.org", "current_status")

	_ = pubSubRep.UpsertNodeSubscription(context.Background(), &pubsubmodel.Subscription{
		JID:          "bob@example.org",
		Subscription: pubsubmodel.Subscribed,
	}, "alice@example.org", "current_status")

	_, _ = rosterRep.UpsertRosterItem(context.Background(), &rostermodel.Item{
		Username:     "alice",
		JID:          "bob@example.org",
		Subscription: "both",
	})

	// process pubsub command
	p := New(nil, nil, r, rosterRep, pubSubRep)
	defer func() { _ = p.Shutdown() }()

	iqID := uuid.New().String()
	iq := xmpp.NewIQType(iqID, xmpp.SetType)
	iq.SetFromJID(j1)
	iq.SetToJID(j1.ToBareJID())

	pubSub := xmpp.NewElementNamespace("pubsub", pubSubOwnerNamespace)
	deleteElem := xmpp.NewElementName("delete")
	deleteElem.SetAttribute("node", "current_status")
	pubSub.AppendElement(deleteElem)
	iq.AppendElement(pubSub)

	p.ProcessIQ(context.Background(), iq)
	elem := stm1.ReceiveElement()
	require.NotNil(t, elem)
	require.Equal(t, "message", elem.Name()) // notification
	require.NotNil(t, elem.Elements().ChildNamespace("event", pubSubEventNamespace))

	elem2 := stm2.ReceiveElement()
	require.NotNil(t, elem2)
	require.Equal(t, "message", elem.Name()) // notification
	eventElem := elem2.Elements().ChildNamespace("event", pubSubEventNamespace)
	require.NotNil(t, eventElem)

	deleteElemResp := eventElem.Elements().Child("delete")
	require.NotNil(t, deleteElemResp)
	require.Equal(t, "current_status", deleteElemResp.Attributes().Get("node"))

	// result IQ
	elem = stm1.ReceiveElement()
	require.NotNil(t, elem)
	require.Equal(t, iqID, elem.ID())
	require.Equal(t, xmpp.ResultType, elem.Type())

	// read node
	n, _ := pubSubRep.FetchNode(context.Background(), "alice@example.org", "current_status")
	require.Nil(t, n)
}

func TestModule_XEP0163_UpdateAffiliations(t *testing.T) {
	r, _, rosterRep, pubSubRep := setupTest("example.org")

	j1, _ := jid.New("user", "example.org", "desktop", true)

	stm1 := stream.NewMockC2S(uuid.New().String(), j1)
	stm1.SetPresence(xmpp.NewPresence(j1, j1, xmpp.AvailableType))

	r.Bind(context.Background(), stm1)

	// create node
	_ = pubSubRep.UpsertNode(context.Background(), &pubsubmodel.Node{
		Host:    "user@example.org",
		Name:    "current_status",
		Options: defaultNodeOptions,
	})

	_ = pubSubRep.UpsertNodeAffiliation(context.Background(), &pubsubmodel.Affiliation{
		JID:         "user@example.org",
		Affiliation: pubsubmodel.Owner,
	}, "user@example.org", "current_status")

	// process pubsub command
	p := New(nil, nil, r, rosterRep, pubSubRep)
	defer func() { _ = p.Shutdown() }()

	// create new affiliation
	iqID := uuid.New().String()
	iq := xmpp.NewIQType(iqID, xmpp.SetType)
	iq.SetFromJID(j1)
	iq.SetToJID(j1.ToBareJID())

	pubSub := xmpp.NewElementNamespace("pubsub", pubSubOwnerNamespace)
	affElem := xmpp.NewElementName("affiliations")
	affElem.SetAttribute("node", "current_status")

	affiliation := xmpp.NewElementName("affiliation")
	affiliation.SetAttribute("jid", "alice@example.org")
	affiliation.SetAttribute("affiliation", pubsubmodel.Owner)
	affElem.AppendElement(affiliation)
	pubSub.AppendElement(affElem)
	iq.AppendElement(pubSub)

	p.ProcessIQ(context.Background(), iq)
	elem := stm1.ReceiveElement()
	require.NotNil(t, elem)
	require.Equal(t, "iq", elem.Name())
	require.Equal(t, xmpp.ResultType, elem.Type())

	aff, _ := pubSubRep.FetchNodeAffiliation(context.Background(), "user@example.org", "current_status", "alice@example.org")
	require.NotNil(t, aff)
	require.Equal(t, "alice@example.org", aff.JID)
	require.Equal(t, pubsubmodel.Owner, aff.Affiliation)

	// remove affiliation
	affiliation.SetAttribute("affiliation", pubsubmodel.None)

	p.ProcessIQ(context.Background(), iq)
	elem = stm1.ReceiveElement()
	require.NotNil(t, elem)
	require.Equal(t, "iq", elem.Name())
	require.Equal(t, xmpp.ResultType, elem.Type())

	aff, _ = pubSubRep.FetchNodeAffiliation(context.Background(), "user@example.org", "current_status", "alice@example.org")
	require.Nil(t, aff)
}

func TestModule_XEP0163_RetrieveAffiliations(t *testing.T) {
	r, _, rosterRep, pubSubRep := setupTest("example.org")

	j1, _ := jid.New("user", "example.org", "desktop", true)

	stm1 := stream.NewMockC2S(uuid.New().String(), j1)
	stm1.SetPresence(xmpp.NewPresence(j1, j1, xmpp.AvailableType))

	r.Bind(context.Background(), stm1)

	// create node and affiliations
	_ = pubSubRep.UpsertNode(context.Background(), &pubsubmodel.Node{
		Host:    "user@example.org",
		Name:    "current_status",
		Options: defaultNodeOptions,
	})

	_ = pubSubRep.UpsertNodeAffiliation(context.Background(), &pubsubmodel.Affiliation{
		JID:         "user@example.org",
		Affiliation: pubsubmodel.Owner,
	}, "user@example.org", "current_status")

	_ = pubSubRep.UpsertNodeAffiliation(context.Background(), &pubsubmodel.Affiliation{
		JID:         "alice@example.org",
		Affiliation: pubsubmodel.Owner,
	}, "user@example.org", "current_status")

	// process pubsub command
	p := New(nil, nil, r, rosterRep, pubSubRep)
	defer func() { _ = p.Shutdown() }()

	iqID := uuid.New().String()
	iq := xmpp.NewIQType(iqID, xmpp.GetType)
	iq.SetFromJID(j1)
	iq.SetToJID(j1.ToBareJID())

	pubSub := xmpp.NewElementNamespace("pubsub", pubSubOwnerNamespace)
	affElem := xmpp.NewElementName("affiliations")
	affElem.SetAttribute("node", "current_status")
	pubSub.AppendElement(affElem)
	iq.AppendElement(pubSub)

	p.ProcessIQ(context.Background(), iq)
	elem := stm1.ReceiveElement()
	require.NotNil(t, elem)
	require.Equal(t, "iq", elem.Name())

	pubSubElem := elem.Elements().ChildNamespace("pubsub", pubSubOwnerNamespace)
	require.NotNil(t, pubSubElem)

	affiliationsElem := pubSubElem.Elements().Child("affiliations")
	require.NotNil(t, affiliationsElem)

	affiliations := affiliationsElem.Elements().Children("affiliation")
	require.Len(t, affiliations, 2)

	require.Equal(t, "user@example.org", affiliations[0].Attributes().Get("jid"))
	require.Equal(t, pubsubmodel.Owner, affiliations[0].Attributes().Get("affiliation"))
	require.Equal(t, "alice@example.org", affiliations[1].Attributes().Get("jid"))
	require.Equal(t, pubsubmodel.Owner, affiliations[1].Attributes().Get("affiliation"))
}

func TestModule_XEP0163_UpdateSubscriptions(t *testing.T) {
	r, _, rosterRep, pubSubRep := setupTest("example.org")

	j1, _ := jid.New("user", "example.org", "desktop", true)

	stm1 := stream.NewMockC2S(uuid.New().String(), j1)
	stm1.SetPresence(xmpp.NewPresence(j1, j1, xmpp.AvailableType))

	r.Bind(context.Background(), stm1)

	// create node
	_ = pubSubRep.UpsertNode(context.Background(), &pubsubmodel.Node{
		Host:    "user@example.org",
		Name:    "current_status",
		Options: defaultNodeOptions,
	})
	_ = pubSubRep.UpsertNodeAffiliation(context.Background(), &pubsubmodel.Affiliation{
		JID:         "user@example.org",
		Affiliation: pubsubmodel.Owner,
	}, "user@example.org", "current_status")

	// process pubsub command
	p := New(nil, nil, r, rosterRep, pubSubRep)
	defer func() { _ = p.Shutdown() }()

	// create new subscription
	iqID := uuid.New().String()
	iq := xmpp.NewIQType(iqID, xmpp.SetType)
	iq.SetFromJID(j1)
	iq.SetToJID(j1.ToBareJID())

	pubSub := xmpp.NewElementNamespace("pubsub", pubSubOwnerNamespace)
	subElem := xmpp.NewElementName("subscriptions")
	subElem.SetAttribute("node", "current_status")

	sub := xmpp.NewElementName("subscription")
	sub.SetAttribute("jid", "alice@example.org")
	sub.SetAttribute("subscription", pubsubmodel.Subscribed)
	subElem.AppendElement(sub)
	pubSub.AppendElement(subElem)
	iq.AppendElement(pubSub)

	p.ProcessIQ(context.Background(), iq)
	elem := stm1.ReceiveElement()
	require.NotNil(t, elem)
	require.Equal(t, "iq", elem.Name())
	require.Equal(t, xmpp.ResultType, elem.Type())

	subs, _ := pubSubRep.FetchNodeSubscriptions(context.Background(), "user@example.org", "current_status")
	require.NotNil(t, subs)
	require.Len(t, subs, 1)
	require.Equal(t, "alice@example.org", subs[0].JID)
	require.Equal(t, pubsubmodel.Subscribed, subs[0].Subscription)

	// remove subscription
	sub.SetAttribute("subscription", pubsubmodel.None)

	p.ProcessIQ(context.Background(), iq)
	elem = stm1.ReceiveElement()
	require.NotNil(t, elem)
	require.Equal(t, "iq", elem.Name())
	require.Equal(t, xmpp.ResultType, elem.Type())

	subs, _ = pubSubRep.FetchNodeSubscriptions(context.Background(), "user@example.org", "current_status")
	require.Nil(t, subs)
}

func TestModule_XEP0163_RetrieveSubscriptions(t *testing.T) {
	r, _, rosterRep, pubSubRep := setupTest("example.org")

	j1, _ := jid.New("user", "example.org", "desktop", true)

	stm1 := stream.NewMockC2S(uuid.New().String(), j1)
	stm1.SetPresence(xmpp.NewPresence(j1, j1, xmpp.AvailableType))

	r.Bind(context.Background(), stm1)

	// create node and affiliations
	_ = pubSubRep.UpsertNode(context.Background(), &pubsubmodel.Node{
		Host:    "user@example.org",
		Name:    "current_status",
		Options: defaultNodeOptions,
	})

	_ = pubSubRep.UpsertNodeAffiliation(context.Background(), &pubsubmodel.Affiliation{
		JID:         "user@example.org",
		Affiliation: pubsubmodel.Owner,
	}, "user@example.org", "current_status")

	_ = pubSubRep.UpsertNodeSubscription(context.Background(), &pubsubmodel.Subscription{
		SubID:        uuid.New().String(),
		JID:          "alice@example.org",
		Subscription: pubsubmodel.Subscribed,
	}, "user@example.org", "current_status")

	// process pubsub command
	p := New(nil, nil, r, rosterRep, pubSubRep)
	defer func() { _ = p.Shutdown() }()

	iqID := uuid.New().String()
	iq := xmpp.NewIQType(iqID, xmpp.GetType)
	iq.SetFromJID(j1)
	iq.SetToJID(j1.ToBareJID())

	pubSub := xmpp.NewElementNamespace("pubsub", pubSubOwnerNamespace)
	affElem := xmpp.NewElementName("subscriptions")
	affElem.SetAttribute("node", "current_status")
	pubSub.AppendElement(affElem)
	iq.AppendElement(pubSub)

	p.ProcessIQ(context.Background(), iq)
	elem := stm1.ReceiveElement()
	require.NotNil(t, elem)
	require.Equal(t, "iq", elem.Name())

	pubSubElem := elem.Elements().ChildNamespace("pubsub", pubSubOwnerNamespace)
	require.NotNil(t, pubSubElem)

	subscriptionsElem := pubSubElem.Elements().Child("subscriptions")
	require.NotNil(t, subscriptionsElem)

	subscriptions := subscriptionsElem.Elements().Children("subscription")
	require.Len(t, subscriptions, 1)

	require.Equal(t, "alice@example.org", subscriptions[0].Attributes().Get("jid"))
	require.Equal(t, pubsubmodel.Subscribed, subscriptions[0].Attributes().Get("subscription"))
}

func TestModule_XEP0163_Subscribe(t *testing.T) {
	r, _, rosterRep, pubSubRep := setupTest("example.org")

	j1, _ := jid.New("user", "example.org", "desktop", true)
	j2, _ := jid.New("alice", "example.org", "desktop", true)

	stm1 := stream.NewMockC2S(uuid.New().String(), j1)
	stm2 := stream.NewMockC2S(uuid.New().String(), j2)
	stm1.SetPresence(xmpp.NewPresence(j1, j1, xmpp.AvailableType))
	stm2.SetPresence(xmpp.NewPresence(j2, j2, xmpp.AvailableType))

	r.Bind(context.Background(), stm1)
	r.Bind(context.Background(), stm2)

	// create node and affiliations
	nodeOpts := defaultNodeOptions
	nodeOpts.NotifySub = true

	_ = pubSubRep.UpsertNode(context.Background(), &pubsubmodel.Node{
		Host:    "user@example.org",
		Name:    "current_status",
		Options: nodeOpts,
	})

	_ = pubSubRep.UpsertNodeAffiliation(context.Background(), &pubsubmodel.Affiliation{
		JID:         "user@example.org",
		Affiliation: pubsubmodel.Owner,
	}, "user@example.org", "current_status")

	_, _ = rosterRep.UpsertRosterItem(context.Background(), &rostermodel.Item{
		Username:     "user",
		JID:          "alice@example.org",
		Subscription: "both",
	})

	// process pubsub command
	p := New(nil, nil, r, rosterRep, pubSubRep)
	defer func() { _ = p.Shutdown() }()

	iqID := uuid.New().String()
	iq := xmpp.NewIQType(iqID, xmpp.SetType)
	iq.SetFromJID(j2)
	iq.SetToJID(j1.ToBareJID())

	pubSub := xmpp.NewElementNamespace("pubsub", pubSubNamespace)
	subElem := xmpp.NewElementName("subscribe")
	subElem.SetAttribute("node", "current_status")
	subElem.SetAttribute("jid", "alice@example.org")
	pubSub.AppendElement(subElem)
	iq.AppendElement(pubSub)

	p.ProcessIQ(context.Background(), iq)
	elem := stm2.ReceiveElement()

	// command reply
	require.NotNil(t, elem)
	require.Equal(t, "iq", elem.Name())
	require.Equal(t, xmpp.ResultType, elem.Type())

	pubSubElem := elem.Elements().ChildNamespace("pubsub", pubSubNamespace)
	require.NotNil(t, pubSubElem)
	subscriptionElem := pubSubElem.Elements().Child("subscription")
	require.NotNil(t, subscriptionElem)
	require.Equal(t, "alice@example.org", subscriptionElem.Attributes().Get("jid"))
	require.Equal(t, "subscribed", subscriptionElem.Attributes().Get("subscription"))
	require.Equal(t, "current_status", subscriptionElem.Attributes().Get("node"))

	// subscription notification
	elem = stm1.ReceiveElement()
	require.NotNil(t, elem)
	require.Equal(t, "message", elem.Name())

	eventElem := elem.Elements().ChildNamespace("event", pubSubEventNamespace)
	require.NotNil(t, eventElem)

	subscriptionElem = eventElem.Elements().Child("subscription")
	require.NotNil(t, subscriptionElem)
	require.Equal(t, "alice@example.org", subscriptionElem.Attributes().Get("jid"))
	require.Equal(t, "subscribed", subscriptionElem.Attributes().Get("subscription"))
	require.Equal(t, "current_status", subscriptionElem.Attributes().Get("node"))

	// check storage subscription
	subs, _ := pubSubRep.FetchNodeSubscriptions(context.Background(), "user@example.org", "current_status")
	require.Len(t, subs, 1)
	require.Equal(t, "alice@example.org", subs[0].JID)
	require.Equal(t, pubsubmodel.Subscribed, subs[0].Subscription)
}

func TestModule_XEP0163_Unsubscribe(t *testing.T) {
	r, _, rosterRep, pubSubRep := setupTest("example.org")

	j1, _ := jid.New("user", "example.org", "desktop", true)
	j2, _ := jid.New("alice", "example.org", "desktop", true)

	stm2 := stream.NewMockC2S(uuid.New().String(), j2)
	stm2.SetPresence(xmpp.NewPresence(j2, j2, xmpp.AvailableType))

	r.Bind(context.Background(), stm2)

	// create node and affiliations
	_ = pubSubRep.UpsertNode(context.Background(), &pubsubmodel.Node{
		Host:    "user@example.org",
		Name:    "current_status",
		Options: defaultNodeOptions,
	})

	_ = pubSubRep.UpsertNodeAffiliation(context.Background(), &pubsubmodel.Affiliation{
		JID:         "user@example.org",
		Affiliation: pubsubmodel.Owner,
	}, "user@example.org", "current_status")

	_, _ = rosterRep.UpsertRosterItem(context.Background(), &rostermodel.Item{
		Username:     "user",
		JID:          "alice@example.org",
		Subscription: "both",
	})

	_ = pubSubRep.UpsertNodeSubscription(context.Background(), &pubsubmodel.Subscription{
		SubID:        uuid.New().String(),
		JID:          "alice@example.org",
		Subscription: pubsubmodel.Subscribed,
	}, "user@example.org", "current_status")

	// process pubsub command
	p := New(nil, nil, r, rosterRep, pubSubRep)
	defer func() { _ = p.Shutdown() }()

	iqID := uuid.New().String()
	iq := xmpp.NewIQType(iqID, xmpp.SetType)
	iq.SetFromJID(j2)
	iq.SetToJID(j1.ToBareJID())

	pubSub := xmpp.NewElementNamespace("pubsub", pubSubNamespace)
	subElem := xmpp.NewElementName("unsubscribe")
	subElem.SetAttribute("node", "current_status")
	subElem.SetAttribute("jid", "alice@example.org")
	pubSub.AppendElement(subElem)
	iq.AppendElement(pubSub)

	p.ProcessIQ(context.Background(), iq)
	elem := stm2.ReceiveElement()

	// command reply
	require.NotNil(t, elem)
	require.Equal(t, "iq", elem.Name())
	require.Equal(t, xmpp.ResultType, elem.Type())

	// check storage subscription
	subs, _ := pubSubRep.FetchNodeSubscriptions(context.Background(), "user@example.org", "current_status")
	require.Len(t, subs, 0)
}

func TestModule_XEP0163_RetrieveItems(t *testing.T) {
	r, _, rosterRep, pubSubRep := setupTest("example.org")

	j1, _ := jid.New("user", "example.org", "desktop", true)
	j2, _ := jid.New("alice", "example.org", "desktop", true)

	stm1 := stream.NewMockC2S(uuid.New().String(), j1)
	stm2 := stream.NewMockC2S(uuid.New().String(), j2)
	stm1.SetPresence(xmpp.NewPresence(j1, j1, xmpp.AvailableType))
	stm2.SetPresence(xmpp.NewPresence(j2, j2, xmpp.AvailableType))
	r.Bind(context.Background(), stm1)
	r.Bind(context.Background(), stm2)

	// create node and affiliations
	_ = pubSubRep.UpsertNode(context.Background(), &pubsubmodel.Node{
		Host:    "user@example.org",
		Name:    "current_status",
		Options: defaultNodeOptions,
	})
	_ = pubSubRep.UpsertNodeAffiliation(context.Background(), &pubsubmodel.Affiliation{
		JID:         "user@example.org",
		Affiliation: pubsubmodel.Owner,
	}, "user@example.org", "current_status")
	_, _ = rosterRep.UpsertRosterItem(context.Background(), &rostermodel.Item{
		Username:     "user",
		JID:          "alice@example.org",
		Subscription: "both",
	})

	// create items
	_ = pubSubRep.UpsertNodeItem(context.Background(), &pubsubmodel.Item{
		ID:        "1",
		Publisher: "alice@example.org",
		Payload:   xmpp.NewElementName("m1"),
	}, "user@example.org", "current_status", 2)

	_ = pubSubRep.UpsertNodeItem(context.Background(), &pubsubmodel.Item{
		ID:        "2",
		Publisher: "alice@example.org",
		Payload:   xmpp.NewElementName("m2"),
	}, "user@example.org", "current_status", 2)

	p := New(nil, nil, r, rosterRep, pubSubRep)
	defer func() { _ = p.Shutdown() }()

	// retrieve all items
	iqID := uuid.New().String()
	iq := xmpp.NewIQType(iqID, xmpp.GetType)
	iq.SetFromJID(j2)
	iq.SetToJID(j1.ToBareJID())

	pubSub := xmpp.NewElementNamespace("pubsub", pubSubNamespace)
	itemsCmdElem := xmpp.NewElementName("items")
	itemsCmdElem.SetAttribute("node", "current_status")
	pubSub.AppendElement(itemsCmdElem)
	iq.AppendElement(pubSub)

	p.ProcessIQ(context.Background(), iq)
	elem := stm2.ReceiveElement()
	require.Equal(t, "iq", elem.Name())
	require.Equal(t, xmpp.ResultType, elem.Type())

	pubSubElem := elem.Elements().ChildNamespace("pubsub", pubSubNamespace)
	require.NotNil(t, pubSubElem)
	itemsElem := pubSubElem.Elements().Child("items")
	require.NotNil(t, itemsElem)
	items := itemsElem.Elements().Children("item")
	require.Len(t, items, 2)

	require.Equal(t, "1", items[0].Attributes().Get("id"))
	require.Equal(t, "2", items[1].Attributes().Get("id"))

	// retrieve item i2
	i2Elem := xmpp.NewElementName("item")
	i2Elem.SetAttribute("id", "2")
	itemsCmdElem.AppendElement(i2Elem)

	p.ProcessIQ(context.Background(), iq)
	elem = stm2.ReceiveElement()
	require.Equal(t, "iq", elem.Name())
	require.Equal(t, xmpp.ResultType, elem.Type())

	pubSubElem = elem.Elements().ChildNamespace("pubsub", pubSubNamespace)
	require.NotNil(t, pubSubElem)
	itemsElem = pubSubElem.Elements().Child("items")
	require.NotNil(t, itemsElem)
	items = itemsElem.Elements().Children("item")
	require.Len(t, items, 1)

	require.Equal(t, "2", items[0].Attributes().Get("id"))
}

func TestModule_XEP0163_SubscribeToAll(t *testing.T) {
	r, _, rosterRep, pubSubRep := setupTest("example.org")

	j1, _ := jid.New("user", "example.org", "desktop", true)

	// create node and affiliations
	_ = pubSubRep.UpsertNode(context.Background(), &pubsubmodel.Node{
		Host:    "alice@example.org",
		Name:    "current_status_1",
		Options: defaultNodeOptions,
	})
	_ = pubSubRep.UpsertNode(context.Background(), &pubsubmodel.Node{
		Host:    "alice@example.org",
		Name:    "current_status_2",
		Options: defaultNodeOptions,
	})
	_ = pubSubRep.UpsertNodeItem(context.Background(), &pubsubmodel.Item{
		ID:        "2",
		Publisher: "alice@example.org",
		Payload:   xmpp.NewElementName("m2"),
	}, "alice@example.org", "current_status_2", 2)

	_, _ = rosterRep.UpsertRosterItem(context.Background(), &rostermodel.Item{
		Username:     "alice",
		JID:          "user@example.org",
		Subscription: "both",
	})
	p := New(nil, nil, r, rosterRep, pubSubRep)
	defer func() { _ = p.Shutdown() }()

	err := p.subscribeToAll(context.Background(), "alice@example.org", j1)
	require.Nil(t, err)

	nodes, _ := pubSubRep.FetchSubscribedNodes(context.Background(), j1.ToBareJID().String())
	require.NotNil(t, nodes)
	require.Len(t, nodes, 2)

	err = p.unsubscribeFromAll(context.Background(), "alice@example.org", j1)
	require.Nil(t, err)

	nodes, _ = pubSubRep.FetchSubscribedNodes(context.Background(), j1.ToBareJID().String())
	require.Nil(t, nodes)
}

func TestModule_XEP0163_FilteredNotifications(t *testing.T) {
	r, presencesRep, rosterRep, pubSubRep := setupTest("example.org")

	j1, _ := jid.New("user", "example.org", "desktop", true)
	j2, _ := jid.New("alice", "example.org", "desktop", true)
	stm1 := stream.NewMockC2S(uuid.New().String(), j1)
	stm2 := stream.NewMockC2S(uuid.New().String(), j2)
	stm1.SetPresence(xmpp.NewPresence(j1, j1, xmpp.AvailableType))
	stm2.SetPresence(xmpp.NewPresence(j2, j2, xmpp.AvailableType))
	r.Bind(context.Background(), stm1)
	r.Bind(context.Background(), stm2)

	// create node, affiliations and subscriptions
	_ = pubSubRep.UpsertNode(context.Background(), &pubsubmodel.Node{
		Host:    "user@example.org",
		Name:    "current_status",
		Options: defaultNodeOptions,
	})

	_ = pubSubRep.UpsertNodeAffiliation(context.Background(), &pubsubmodel.Affiliation{
		JID:         "user@example.org",
		Affiliation: pubsubmodel.Owner,
	}, "user@example.org", "current_status")

	_, _ = rosterRep.UpsertRosterItem(context.Background(), &rostermodel.Item{
		Username:     "user",
		JID:          "alice@example.org",
		Subscription: "both",
	})

	_ = pubSubRep.UpsertNodeSubscription(context.Background(), &pubsubmodel.Subscription{
		SubID:        uuid.New().String(),
		JID:          "alice@example.org",
		Subscription: pubsubmodel.Subscribed,
	}, "user@example.org", "current_status")

	// set capabilities
	_ = presencesRep.UpsertCapabilities(context.Background(), &capsmodel.Capabilities{
		Node:     "http://code.google.com/p/exodus",
		Ver:      "v0.1",
		Features: []string{"current_status+notify"},
	})
	caps := xep0115.New(r, presencesRep, "id-123")

	// register presence
	pr2 := xmpp.NewPresence(j2, j2, xmpp.AvailableType)
	c := xmpp.NewElementNamespace("c", "http://jabber.org/protocol/caps")
	c.SetAttribute("hash", "sha-1")
	c.SetAttribute("node", "http://code.google.com/p/exodus")
	c.SetAttribute("ver", "v0.1")
	pr2.AppendElement(c)

	_, _ = caps.RegisterPresence(context.Background(), pr2)

	// process pubsub command
	p := New(nil, caps, r, rosterRep, pubSubRep)
	defer func() { _ = p.Shutdown() }()

	iqID := uuid.New().String()
	iq := xmpp.NewIQType(iqID, xmpp.SetType)
	iq.SetFromJID(j1)
	iq.SetToJID(j1.ToBareJID())

	pubSubEl := xmpp.NewElementNamespace("pubsub", pubSubNamespace)
	publishEl := xmpp.NewElementName("publish")
	publishEl.SetAttribute("node", "current_status")
	itemEl := xmpp.NewElementName("item")
	itemEl.SetAttribute("id", "123")
	entryEl := xmpp.NewElementNamespace("entry", "http://www.w3.org/2005/Atom")
	itemEl.AppendElement(entryEl)
	publishEl.AppendElement(itemEl)
	pubSubEl.AppendElement(publishEl)

	iq.AppendElement(pubSubEl)

	p.ProcessIQ(context.Background(), iq)
	elem := stm2.ReceiveElement()
	require.Equal(t, "message", elem.Name())
	require.Equal(t, xmpp.HeadlineType, elem.Type())

	eventEl := elem.Elements().ChildNamespace("event", pubSubEventNamespace)
	require.NotNil(t, eventEl)

	itemsEl := eventEl.Elements().Child("items")
	require.NotNil(t, itemsEl)

	require.Equal(t, "123", itemsEl.Elements().Child("item").Attributes().Get("id"))
}

func setupTest(domain string) (router.Router, repository.Presences, repository.Roster, repository.PubSub) {
	hosts, _ := host.New([]host.Config{{Name: domain, Certificate: tls.Certificate{}}})

	presencesRep := memorystorage.NewPresences()
	rosterRep := memorystorage.NewRoster()
	pubSubRep := memorystorage.NewPubSub()
	r, _ := router.New(
		hosts,
		c2srouter.New(memorystorage.NewUser(), memorystorage.NewBlockList()),
		nil,
	)
	return r, presencesRep, rosterRep, pubSubRep
}
