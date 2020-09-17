package xep0030

import (
	"context"
	"sort"
	"testing"

	rostermodel "github.com/dantin/cubit/model/roster"
	"github.com/dantin/cubit/stream"
	"github.com/dantin/cubit/xmpp"
	"github.com/dantin/cubit/xmpp/jid"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestModule_XEP0030_ServerProvider_Features(t *testing.T) {
	r, rosterRep := setupTest("example.org")

	var sp = serverProvider{router: r, rosterRep: rosterRep}

	sp.registerServerFeature("sf0")
	sp.registerServerFeature("sf1")
	sp.registerServerFeature("sf1")
	sp.registerAccountFeature("af0")
	sp.registerAccountFeature("af1")
	sp.registerAccountFeature("af1")
	require.Equal(t, sp.serverFeatures, []Feature{"sf0", "sf1"})
	require.Equal(t, sp.accountFeatures, []Feature{"af0", "af1"})

	sp.unregisterServerFeature("sf1")
	sp.unregisterAccountFeature("af0")
	require.Equal(t, sp.serverFeatures, []Feature{"sf0"})
	require.Equal(t, sp.accountFeatures, []Feature{"af1"})

	srvJID, _ := jid.New("", "example.org", "", true)
	accJID, _ := jid.New("user", "example.org", "desktop", true)
	accJID2, _ := jid.New("alice", "example.org", "desktop", true)

	features, sErr := sp.Features(context.Background(), srvJID, accJID, "node")
	require.Nil(t, features)
	require.Nil(t, sErr)

	features, sErr = sp.Features(context.Background(), srvJID, accJID, "")
	require.Equal(t, features, []Feature{"sf0"})
	require.Nil(t, sErr)

	features, sErr = sp.Features(context.Background(), accJID.ToBareJID(), accJID, "")
	require.Equal(t, features, []Feature{"af1"})
	require.Nil(t, sErr)

	features, sErr = sp.Features(context.Background(), accJID2.ToBareJID(), accJID, "")
	require.Nil(t, features)
	require.Equal(t, sErr, xmpp.ErrSubscriptionRequired)
}

func TestModule_XEP0030_ServerProvider_Identities(t *testing.T) {
	r, rosterRep := setupTest("example.org")

	var sp = serverProvider{router: r, rosterRep: rosterRep}

	srvJID, _ := jid.New("", "example.org", "", true)
	accJID, _ := jid.New("user", "example.org", "desktop", true)
	require.Nil(t, sp.Identities(context.Background(), srvJID, accJID, "node"))

	require.Equal(t, sp.Identities(context.Background(), srvJID, accJID, ""), []Identity{
		{Type: "im", Category: "server", Name: "cubit"},
	})
	require.Equal(t, sp.Identities(context.Background(), accJID.ToBareJID(), accJID, ""), []Identity{
		{Type: "registered", Category: "account"},
	})
}

func TestModule_XEP0030_ServerProvider_Items(t *testing.T) {
	r, rosterRep := setupTest("example.org")

	var sp = serverProvider{router: r, rosterRep: rosterRep}

	srvJID, _ := jid.New("", "example.org", "", true)
	accJID1, _ := jid.New("user", "example.org", "desktop", true)
	accJID2, _ := jid.New("alice", "example.org", "desktop", true)
	accJID3, _ := jid.New("alice", "example.org", "surface", true)

	stm1 := stream.NewMockC2S(uuid.New().String(), accJID1)
	stm2 := stream.NewMockC2S(uuid.New().String(), accJID2)
	stm3 := stream.NewMockC2S(uuid.New().String(), accJID3)

	r.Bind(context.Background(), stm1)
	r.Bind(context.Background(), stm2)
	r.Bind(context.Background(), stm3)

	items, sErr := sp.Items(context.Background(), srvJID, accJID1, "node")
	require.Nil(t, items)
	require.Nil(t, sErr)

	items, sErr = sp.Items(context.Background(), srvJID, accJID1, "")
	require.Equal(t, items, []Item{
		{Jid: accJID1.ToBareJID().String()},
	})
	require.Nil(t, sErr)

	items, sErr = sp.Items(context.Background(), accJID2.ToBareJID(), accJID1, "")
	require.Nil(t, items)
	require.Equal(t, sErr, xmpp.ErrSubscriptionRequired)

	_, _ = rosterRep.UpsertRosterItem(context.Background(), &rostermodel.Item{
		Username:     "user",
		JID:          "alice@example.org",
		Subscription: "both",
	})
	items, sErr = sp.Items(context.Background(), accJID2.ToBareJID(), accJID1, "")
	sort.Slice(items, func(i, j int) bool { return items[i].Jid < items[j].Jid })

	require.Equal(t, items, []Item{
		{Jid: accJID2.String()}, {Jid: accJID3.String()},
	})
	require.Nil(t, sErr)
}
