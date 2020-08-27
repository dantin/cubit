package memorystorage

import (
	"context"
	"testing"

	"github.com/dantin/cubit/xmpp"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage_InsertPrivateXML(t *testing.T) {
	private := xmpp.NewElementNamespace("demo", "demo:ns")

	s := NewPrivate()
	EnableMockedError()
	err := s.UpsertPrivateXML(context.Background(), []xmpp.XElement{private}, "demo:ns", "alice")
	require.Equal(t, ErrMocked, err)
	DisableMockedError()

	err = s.UpsertPrivateXML(context.Background(), []xmpp.XElement{private}, "demo:ns", "alice")
	require.Nil(t, err)
}

func TestMemoryStorage_FetchPrivateXML(t *testing.T) {
	private := xmpp.NewElementNamespace("demo", "demo:ns")

	s := NewPrivate()
	_ = s.UpsertPrivateXML(context.Background(), []xmpp.XElement{private}, "demo:ns", "alice")

	EnableMockedError()
	_, err := s.FetchPrivateXML(context.Background(), "demo:ns", "alice")
	require.Equal(t, ErrMocked, err)
	DisableMockedError()

	elems, _ := s.FetchPrivateXML(context.Background(), "demo:ns", "alice")
	require.Len(t, elems, 1)
}
