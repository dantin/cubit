package c2srouter

import (
	"context"
	"testing"

	"github.com/dantin/cubit/router"
	"github.com/dantin/cubit/stream"
	"github.com/dantin/cubit/xmpp"
	"github.com/dantin/cubit/xmpp/jid"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestResources_Binding(t *testing.T) {
	j, _ := jid.NewWithString("alice@example.org/desktop", true)
	stm := stream.NewMockC2S("id-1", j)
	res := resources{}

	require.Equal(t, 0, res.len())

	res.bind(stm)

	require.Equal(t, 1, res.len())
	require.NotNil(t, res.stream("desktop"))
	require.Len(t, res.allStreams(), 1)

	res.unbind("desktop")

	require.Nil(t, res.stream("desktop"))
	require.Len(t, res.allStreams(), 0)
}

func TestResources_Route(t *testing.T) {
	j1, _ := jid.NewWithString("alice@example.org/desktop", true)
	j2, _ := jid.NewWithString("alice@example.org/surface", true)
	j3, _ := jid.NewWithString("alice@example.org/macbook", true)
	j4, _ := jid.NewWithString("alice@example.org", true)

	stm1 := stream.NewMockC2S("id-1", j1)
	stm2 := stream.NewMockC2S("id-2", j2)

	stm1.SetPresence(xmpp.NewPresence(j1.ToBareJID(), j1, xmpp.AvailableType))
	stm2.SetPresence(xmpp.NewPresence(j2.ToBareJID(), j2, xmpp.AvailableType))

	res := resources{}
	res.bind(stm1)
	res.bind(stm2)

	msgID := uuid.New().String()
	msg := xmpp.NewMessageType(msgID, xmpp.NormalType)
	msg.SetFromJID(j1)
	msg.SetToJID(j3)

	err := res.route(context.Background(), msg)

	require.Equal(t, router.ErrResourceNotFound, err)

	msg.SetToJID(j2)
	err = res.route(context.Background(), msg)

	require.Nil(t, err)

	elem := stm2.ReceiveElement()

	require.NotNil(t, elem)
	require.Equal(t, "message", elem.Name())
	require.Equal(t, msgID, elem.ID())

	msgID = uuid.New().String()
	msg = xmpp.NewMessageType(msgID, xmpp.NormalType)
	msg.SetFromJID(j1)
	msg.SetToJID(j4)

	err = res.route(context.Background(), msg)

	require.Nil(t, err)

	elem1 := stm1.ReceiveElement()
	elem2 := stm2.ReceiveElement()

	require.Equal(t, "message", elem1.Name())
	require.Equal(t, elem1.ID(), elem2.ID())
	require.Equal(t, elem1.Name(), elem2.Name())
}
