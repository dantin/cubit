package memorystorage

import (
	"context"
	"testing"

	"github.com/dantin/cubit/xmpp"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage_InsertVCard(t *testing.T) {
	vCard := xmpp.NewElementName("vCard")
	fn := xmpp.NewElementName("FN")
	fn.SetText("TEXT")
	vCard.AppendElement(fn)

	s := NewVCard()

	EnableMockedError()
	require.Equal(t, ErrMocked, s.UpsertVCard(context.Background(), vCard, "user"))
	DisableMockedError()

	require.Nil(t, s.UpsertVCard(context.Background(), vCard, "user"))
}

func TestMemoryStorage_FetchVCard(t *testing.T) {
	vCard := xmpp.NewElementName("vCard")
	fn := xmpp.NewElementName("FN")
	fn.SetText("TEXT")
	vCard.AppendElement(fn)

	s := NewVCard()
	_ = s.UpsertVCard(context.Background(), vCard, "user")

	EnableMockedError()
	_, err := s.FetchVCard(context.Background(), "user")
	require.Equal(t, ErrMocked, err)
	DisableMockedError()

	elem, _ := s.FetchVCard(context.Background(), "user")
	require.NotNil(t, elem)
}
