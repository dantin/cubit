package c2srouter

import (
	"context"
	"testing"

	"github.com/dantin/cubit/model"
	"github.com/dantin/cubit/router"
	memorystorage "github.com/dantin/cubit/storage/memory"
	"github.com/dantin/cubit/storage/repository"
	"github.com/dantin/cubit/stream"
	"github.com/dantin/cubit/xmpp"
	"github.com/dantin/cubit/xmpp/jid"
	"github.com/stretchr/testify/require"
)

func TestRoster_Binding(t *testing.T) {
	j1, _ := jid.NewWithString("user@example.org/desktop", true)
	j2, _ := jid.NewWithString("user@example.org/surface", true)

	stm1 := stream.NewMockC2S("1", j1)
	stm2 := stream.NewMockC2S("1", j2)

	r, _, _ := setupTest()

	r.Bind(stm1)
	r.Bind(stm2)
	stm1.SetPresence(xmpp.NewPresence(j1.ToBareJID(), j1, xmpp.AvailableType))
	stm2.SetPresence(xmpp.NewPresence(j2.ToBareJID(), j2, xmpp.AvailableType))

	require.Len(t, r.Streams("user"), 2)
	require.NotNil(t, r.Stream("user", "desktop"))
	require.NotNil(t, r.Stream("user", "surface"))

	r.Unbind("user", "desktop")
	r.Unbind("user", "surface")

	require.Len(t, r.Streams("user"), 0)

	r.(*c2sRouter).mu.RLock()
	require.Len(t, r.(*c2sRouter).tbl, 0)
	r.(*c2sRouter).mu.RUnlock()
}

func TestRoster_Routing(t *testing.T) {
	j1, _ := jid.NewWithString("alice@example.org/desktop", true)
	j2, _ := jid.NewWithString("bob@example.org/desktop", true)
	stm1 := stream.NewMockC2S("1", j1)
	stm2 := stream.NewMockC2S("2", j2)

	r, userRep, blockListRep := setupTest()

	err := r.Route(context.Background(), xmpp.NewPresence(j1, j1, xmpp.AvailableType), true)
	require.Equal(t, router.ErrNotExistingAccount, err)

	_ = userRep.UpsertUser(context.Background(), &model.User{Username: "alice"})
	_ = userRep.UpsertUser(context.Background(), &model.User{Username: "bob"})

	err = r.Route(context.Background(), xmpp.NewPresence(j1, j1, xmpp.AvailableType), true)
	require.Equal(t, router.ErrNotAuthenticated, err)

	r.Bind(stm1)
	stm1.SetPresence(xmpp.NewPresence(j1.ToBareJID(), j1, xmpp.AvailableType))

	err = r.Route(context.Background(), xmpp.NewPresence(j1, j1, xmpp.AvailableType), true)
	require.Nil(t, err)

	// block jid
	r.Bind(stm2)
	stm2.SetPresence(xmpp.NewPresence(j2.ToBareJID(), j2, xmpp.AvailableType))

	_ = blockListRep.InsertBlockListItem(context.Background(), &model.BlockListItem{
		Username: "alice",
		JID:      "example.org/desktop",
	})

	err = r.Route(context.Background(), xmpp.NewPresence(j1.ToBareJID(), j2, xmpp.AvailableType), true)
	require.Equal(t, router.ErrBlockedJID, err)
}

func setupTest() (router.C2SRouter, repository.User, repository.BlockList) {
	userRep := memorystorage.NewUser()
	blockListRep := memorystorage.NewBlockList()
	return New(userRep, blockListRep), userRep, blockListRep
}
