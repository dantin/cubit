package memorystorage

import (
	"context"
	"reflect"
	"testing"

	pubsubmodel "github.com/dantin/cubit/model/pubsub"
	"github.com/dantin/cubit/xmpp"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage_PubSubNode(t *testing.T) {
	s := NewPubSub()
	node := &pubsubmodel.Node{
		Host: "alice@example.org",
		Name: "status",
	}
	require.Nil(t, s.UpsertNode(context.Background(), node))

	n, err := s.FetchNode(context.Background(), "alice@example.org", "status")
	require.Nil(t, err)
	require.NotNil(t, n)

	require.True(t, reflect.DeepEqual(n, node))

	node2 := &pubsubmodel.Node{
		Host: "alice@example.org",
		Name: "status_2",
	}
	node3 := &pubsubmodel.Node{
		Host: "alice@example.org",
		Name: "status_3",
	}
	node4 := &pubsubmodel.Node{
		Host: "bob@example.org",
		Name: "status_4",
	}
	require.Nil(t, s.UpsertNode(context.Background(), node2))
	require.Nil(t, s.UpsertNode(context.Background(), node3))
	require.Nil(t, s.UpsertNode(context.Background(), node4))

	nodes, err := s.FetchNodes(context.Background(), "alice@example.org")
	require.Nil(t, err)
	require.NotNil(t, nodes)

	require.Len(t, nodes, 3)
	require.Equal(t, "status", nodes[0].Name)
	require.Equal(t, "status_2", nodes[1].Name)
	require.Equal(t, "status_3", nodes[2].Name)

	require.Nil(t, s.DeleteNode(context.Background(), "alice@example.org", "status_2"))

	nodes, err = s.FetchNodes(context.Background(), "alice@example.org")
	require.Nil(t, err)
	require.NotNil(t, nodes)

	require.Len(t, nodes, 2)
	require.Equal(t, "status", nodes[0].Name)
	require.Equal(t, "status_3", nodes[1].Name)

	// fetch hosts
	hosts, err := s.FetchHosts(context.Background())
	require.Nil(t, err)
	require.Len(t, hosts, 2)
}

func TestMemoryStorage_PubSubNodeItem(t *testing.T) {
	s := NewPubSub()
	item1 := &pubsubmodel.Item{
		ID:        "1",
		Publisher: "alice@example.org",
		Payload:   xmpp.NewElementName("a"),
	}
	item2 := &pubsubmodel.Item{
		ID:        "2",
		Publisher: "bob@example.org",
		Payload:   xmpp.NewElementName("b"),
	}
	item3 := &pubsubmodel.Item{
		ID:        "3",
		Publisher: "bob@example.org",
		Payload:   xmpp.NewElementName("c"),
	}
	require.Nil(t, s.UpsertNodeItem(context.Background(), item1, "alice@example.org", "status", 1))
	require.Nil(t, s.UpsertNodeItem(context.Background(), item2, "alice@example.org", "status", 1))

	items, err := s.FetchNodeItems(context.Background(), "alice@example.org", "status")
	require.Nil(t, err)
	require.NotNil(t, items)

	require.Len(t, items, 1)
	require.True(t, reflect.DeepEqual(&items[0], item2))

	// update item
	require.Nil(t, s.UpsertNodeItem(context.Background(), item3, "alice@example.org", "status", 2))

	items, err = s.FetchNodeItems(context.Background(), "alice@example.org", "status")
	require.Nil(t, err)
	require.NotNil(t, items)

	require.Len(t, items, 2)
	require.True(t, reflect.DeepEqual(&items[0], item2))
	require.True(t, reflect.DeepEqual(&items[1], item3))

	items, err = s.FetchNodeItemsWithIDs(context.Background(), "alice@example.org", "status", []string{"3"})
	require.Nil(t, err)
	require.NotNil(t, items)

	require.Len(t, items, 1)
	require.Equal(t, "3", items[0].ID)
}

func TestMemoryStorage_PubSubNodeAffiliation(t *testing.T) {
	s := NewPubSub()
	aff1 := &pubsubmodel.Affiliation{
		JID:         "alice@example.org",
		Affiliation: "publisher",
	}
	aff2 := &pubsubmodel.Affiliation{
		JID:         "bob@example.org",
		Affiliation: "publisher",
	}
	require.Nil(t, s.UpsertNodeAffiliation(context.Background(), aff1, "alice@example.org", "status"))
	require.Nil(t, s.UpsertNodeAffiliation(context.Background(), aff2, "alice@example.org", "status"))

	affiliations, err := s.FetchNodeAffiliations(context.Background(), "alice@example.org", "status")

	require.Nil(t, err)
	require.NotNil(t, affiliations)
	require.Len(t, affiliations, 2)

	// update affiliation
	aff2.Affiliation = "owner"
	require.Nil(t, s.UpsertNodeAffiliation(context.Background(), aff2, "alice@example.org", "status"))

	affiliations, err = s.FetchNodeAffiliations(context.Background(), "alice@example.org", "status")

	require.Nil(t, err)
	require.NotNil(t, affiliations)
	require.Len(t, affiliations, 2)

	var updated bool
	for _, aff := range affiliations {
		if aff.JID == "bob@example.org" {
			require.Equal(t, "owner", aff.Affiliation)
			updated = true
			break
		}
	}
	if !updated {
		require.Fail(t, "affiliation for 'bob@example.org' not found")
	}

	// delete affiliation
	err = s.DeleteNodeAffiliation(context.Background(), "bob@example.org", "alice@example.org", "status")
	require.Nil(t, err)

	affiliations, err = s.FetchNodeAffiliations(context.Background(), "alice@example.org", "status")
	require.Nil(t, err)
	require.NotNil(t, affiliations)
	require.Len(t, affiliations, 1)
}

func TestMemoryStorage_PubSubNodeSubscription(t *testing.T) {
	s := NewPubSub()
	node := &pubsubmodel.Node{
		Host: "alice@example.org",
		Name: "status",
	}
	_ = s.UpsertNode(context.Background(), node)

	node2 := &pubsubmodel.Node{
		Host: "bob@example.org",
		Name: "status",
	}
	_ = s.UpsertNode(context.Background(), node2)

	sub1 := &pubsubmodel.Subscription{
		SubID:        "1",
		JID:          "alice@example.org",
		Subscription: "subscribed",
	}
	sub2 := &pubsubmodel.Subscription{
		SubID:        "2",
		JID:          "bob@example.org",
		Subscription: "unsubscribed",
	}
	sub3 := &pubsubmodel.Subscription{
		SubID:        "3",
		JID:          "alice@example.org",
		Subscription: "subscribed",
	}
	require.Nil(t, s.UpsertNodeSubscription(context.Background(), sub1, "alice@example.org", "status"))
	require.Nil(t, s.UpsertNodeSubscription(context.Background(), sub2, "alice@example.org", "status"))
	require.Nil(t, s.UpsertNodeSubscription(context.Background(), sub3, "bob@example.org", "status"))

	// fetch user subscribed nodes
	nodes, err := s.FetchSubscribedNodes(context.Background(), "alice@example.org")
	require.Nil(t, err)
	require.Len(t, nodes, 2)

	subscriptions, err := s.FetchNodeSubscriptions(context.Background(), "alice@example.org", "status")
	require.Nil(t, err)
	require.NotNil(t, subscriptions)
	require.Len(t, subscriptions, 2)

	// update affiliation
	sub2.Subscription = "subscribed"
	require.Nil(t, s.UpsertNodeSubscription(context.Background(), sub2, "alice@example.org", "status"))

	subscriptions, err = s.FetchNodeSubscriptions(context.Background(), "alice@example.org", "status")
	require.Nil(t, err)
	require.NotNil(t, subscriptions)
	require.Len(t, subscriptions, 2)

	var updated bool
	for _, sub := range subscriptions {
		if sub.JID == "bob@example.org" {
			require.Equal(t, "subscribed", sub.Subscription)
			updated = true
			break
		}
	}
	if !updated {
		require.Fail(t, "subscription for 'bob@example.org' not found")
	}

	// delete subscription
	err = s.DeleteNodeSubscription(context.Background(), "bob@example.org", "alice@example.org", "status")
	require.Nil(t, err)

	subscriptions, err = s.FetchNodeSubscriptions(context.Background(), "alice@example.org", "status")
	require.Nil(t, err)
	require.NotNil(t, subscriptions)
	require.Len(t, subscriptions, 1)
}
