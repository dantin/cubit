package stream

import (
	"context"

	"github.com/dantin/cubit/xmpp"
	"github.com/dantin/cubit/xmpp/jid"
)

// InStream represents a generic incoming stream.
type InStream interface {
	ID() string
	Disconnect(ctx context.Context, err error)
}

// InOutStream represents a generic bi-direction stream.
type InOutStream interface {
	InStream
	SendElement(ctx context.Context, elem xmpp.XElement)
}

// C2S represents a client-to-server bi-direction XMPP stream.
type C2S interface {
	InOutStream

	Context() context.Context

	SetValue(key, value interface{})
	Value(key interface{}) interface{}

	Username() string
	Domain() string
	Resource() string

	JID() *jid.JID

	IsSecured() bool
	IsAuthenticated() bool

	Presence() *xmpp.Presence
}

// S2SIn represents an incoming server-to-server one-direction XMPP stream.
type S2SIn interface {
	InStream
}

// S2SOut represents an outgoing server-to-server one-dirction XMPP stream.
type S2SOut interface {
	InOutStream
}
