package xep0163

import (
	"context"
	"reflect"
	"testing"

	pubsubmodel "github.com/dantin/cubit/model/pubsub"
	rostermodel "github.com/dantin/cubit/model/roster"
	memorystorage "github.com/dantin/cubit/storage/memory"
	"github.com/dantin/cubit/xmpp"
	"github.com/dantin/cubit/xmpp/jid"
	"github.com/stretchr/testify/require"
)

func TestModule_XEP0163_DiscoInfoProvider_Identities(t *testing.T) {
	j1, _ := jid.New("alice", "example.org", "desktop", true)
	j2, _ := jid.New("bob", "example.org", "surface", true)

	dp := &discoInfoProvider{}

	ids := dp.Identities(context.Background(), j1, j2, "")
	require.Len(t, ids, 2)

	require.Equal(t, "collection", ids[0].Type)
	require.Equal(t, "pubsub", ids[0].Category)
	require.Equal(t, "pep", ids[1].Type)
	require.Equal(t, "pubsub", ids[1].Category)

	ids = dp.Identities(context.Background(), j1, j2, "node")
	require.Len(t, ids, 2)

	require.Equal(t, "leaf", ids[0].Type)
	require.Equal(t, "pubsub", ids[0].Category)
	require.Equal(t, "pep", ids[1].Type)
	require.Equal(t, "pubsub", ids[1].Category)
}

func TestModule_XEP0163_DiscoInfoProvider_Items(t *testing.T) {
	j1, _ := jid.New("alice", "example.org", "desktop", true)
	j2, _ := jid.New("bob", "example.org", "surface", true)

	pubSubRep := memorystorage.NewPubSub()

	_ = pubSubRep.UpsertNode(context.Background(), &pubsubmodel.Node{
		Host:    "alice@example.org",
		Name:    "current_status",
		Options: defaultNodeOptions,
	})
	rosterRep := memorystorage.NewRoster()
	dp := &discoInfoProvider{
		rosterRep: rosterRep,
		pubSubRep: pubSubRep,
	}

	items, err := dp.Items(context.Background(), j1, j2, "")
	require.Nil(t, items)
	require.NotNil(t, err)
	require.Equal(t, xmpp.ErrSubscriptionRequired, err)

	_, _ = rosterRep.UpsertRosterItem(context.Background(), &rostermodel.Item{
		Username:     "bob",
		JID:          "alice@example.org",
		Subscription: rostermodel.SubscriptionTo,
	})

	items, err = dp.Items(context.Background(), j1, j2, "")
	require.Nil(t, err)
	require.Len(t, items, 1)

	require.Equal(t, "alice@example.org", items[0].Jid)
	require.Equal(t, "current_status", items[0].Node)
}

func TestModule_XEP0163_DiscoInfoProvider_Features(t *testing.T) {
	j1, _ := jid.New("alice", "example.org", "desktop", true)
	j2, _ := jid.New("bob", "example.org", "desktop", true)

	dp := &discoInfoProvider{}

	features, _ := dp.Features(context.Background(), j1, j2, "")
	require.True(t, reflect.DeepEqual(features, pepFeatures))

	features, _ = dp.Features(context.Background(), j1, j2, "node")
	require.True(t, reflect.DeepEqual(features, pepFeatures))
}

func TestModule_XEP0163_DiscoInfoProvider_Form(t *testing.T) {
	j1, _ := jid.New("alice", "example.org", "desktop", true)
	j2, _ := jid.New("bob", "example.org", "desktop", true)

	dp := &discoInfoProvider{}

	features, _ := dp.Features(context.Background(), j1, j2, "")
	require.True(t, reflect.DeepEqual(features, pepFeatures))

	form, _ := dp.Form(context.Background(), j1, j2, "")
	require.Nil(t, form)

	form, _ = dp.Form(context.Background(), j1, j2, "node")
	require.Nil(t, form)
}
