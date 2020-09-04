package xmpp_test

import (
	"testing"

	"github.com/dantin/cubit/xmpp"
	"github.com/dantin/cubit/xmpp/jid"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestIQBuild(t *testing.T) {
	j, _ := jid.New("username", "example.org", "desktop", false)

	// wrong name
	elem := xmpp.NewElementName("message")
	_, err := xmpp.NewIQFromElement(elem, j, j)
	require.NotNil(t, err)

	// no ID
	elem.SetName("iq")
	_, err = xmpp.NewIQFromElement(elem, j, j)
	require.NotNil(t, err)

	// no type
	elem.SetID(uuid.New().String())
	_, err = xmpp.NewIQFromElement(elem, j, j)
	require.NotNil(t, err)

	// invalid type
	elem.SetType("invalid")
	_, err = xmpp.NewIQFromElement(elem, j, j)
	require.NotNil(t, err)

	// 'get' with no child
	elem.SetType(xmpp.GetType)
	_, err = xmpp.NewIQFromElement(elem, j, j)
	require.NotNil(t, err)

	// 'result' with more than one child
	elem.SetType(xmpp.ResultType)
	elem.AppendElements([]xmpp.XElement{xmpp.NewElementName("a"), xmpp.NewElementName("b")})
	_, err = xmpp.NewIQFromElement(elem, j, j)
	require.NotNil(t, err)

	elem.SetType(xmpp.ResultType)
	elem.ClearElements()
	elem.AppendElements([]xmpp.XElement{xmpp.NewElementName("a")})
	iq, err := xmpp.NewIQFromElement(elem, j, j)
	require.Nil(t, err)
	require.NotNil(t, iq)
}

func TestIQType(t *testing.T) {
	require.True(t, xmpp.NewIQType(uuid.New().String(), xmpp.GetType).IsGet())
	require.True(t, xmpp.NewIQType(uuid.New().String(), xmpp.SetType).IsSet())
	require.True(t, xmpp.NewIQType(uuid.New().String(), xmpp.ResultType).IsResult())
}

func TestResultIQ(t *testing.T) {
	j, _ := jid.New("", "example.org", "", true)

	id := uuid.New().String()
	iq := xmpp.NewIQType(id, xmpp.GetType)
	iq.SetFromJID(j)
	iq.SetToJID(j)
	iq.AppendElement(xmpp.NewElementNamespace("ping", "ur:xmpp:ping"))
	result := iq.ResultIQ()
	require.Equal(t, xmpp.ResultType, result.Type())
	require.Equal(t, id, result.ID())
}

func TestIQJID(t *testing.T) {
	from, _ := jid.New("username", "test.org", "desktop", false)
	to, _ := jid.New("username", "example.org", "desktop", false)
	iq := xmpp.NewIQType(uuid.New().String(), xmpp.GetType)
	iq.SetFromJID(from)
	require.Equal(t, iq.FromJID().String(), iq.From())
	iq.SetToJID(to)
	require.Equal(t, iq.ToJID().String(), iq.To())
}
