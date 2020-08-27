package memorystorage

import (
	"context"
	"testing"

	"github.com/dantin/cubit/model"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage_InsertOrUpdateBlockListItems(t *testing.T) {
	items := []model.BlockListItem{
		{Username: "demo", JID: "user@example.org"},
		{Username: "demo", JID: "alice@example.org"},
		{Username: "demo", JID: "bob@example.org"},
	}
	s := NewBlockList()
	EnableMockedError()
	require.Equal(t, ErrMocked, s.InsertBlockListItem(context.Background(), &model.BlockListItem{Username: "demo", JID: "alice@example.org"}))
	DisableMockedError()

	require.Nil(t, s.InsertBlockListItem(context.Background(), &model.BlockListItem{Username: "demo", JID: "user@example.org"}))
	require.Nil(t, s.InsertBlockListItem(context.Background(), &model.BlockListItem{Username: "demo", JID: "alice@example.org"}))
	require.Nil(t, s.InsertBlockListItem(context.Background(), &model.BlockListItem{Username: "demo", JID: "bob@example.org"}))

	EnableMockedError()
	_, err := s.FetchBlockListItems(context.Background(), "demo")
	require.Equal(t, ErrMocked, err)
	DisableMockedError()

	sItems, _ := s.FetchBlockListItems(context.Background(), "demo")
	require.Equal(t, items, sItems)
}

func TestMemoryStorage_DeleteBlockListItems(t *testing.T) {
	s := NewBlockList()
	require.Nil(t, s.InsertBlockListItem(context.Background(), &model.BlockListItem{Username: "demo", JID: "user@example.org"}))
	require.Nil(t, s.InsertBlockListItem(context.Background(), &model.BlockListItem{Username: "demo", JID: "alice@example.org"}))
	require.Nil(t, s.InsertBlockListItem(context.Background(), &model.BlockListItem{Username: "demo", JID: "bob@example.org"}))

	EnableMockedError()
	require.Equal(t, ErrMocked, s.DeleteBlockListItem(context.Background(), &model.BlockListItem{Username: "demo", JID: "alice@example.org"}))
	DisableMockedError()

	require.Nil(t, s.DeleteBlockListItem(context.Background(), &model.BlockListItem{Username: "demo", JID: "alice@example.org"}))

	sItems, _ := s.FetchBlockListItems(context.Background(), "demo")
	require.Equal(t, []model.BlockListItem{
		{Username: "demo", JID: "user@example.org"},
		{Username: "demo", JID: "bob@example.org"},
	}, sItems)
}
