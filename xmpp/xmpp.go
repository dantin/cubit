package xmpp

import (
	"bytes"
	"fmt"
	"io"

	"github.com/dantin/cubit/util/pool"
	"github.com/dantin/cubit/xmpp/jid"
)

var bufPool = pool.NewBufferPool()

// ErrorType represents an 'error' stanza type.
const ErrorType = "error"

// XElement represents a generic XML node element.
type XElement interface {
	fmt.Stringer

	Name() string
	Attributes() AttributeSet
	Elements() ElementSet

	Text() string

	ID() string
	Namespace() string
	Language() string
	Version() string
	From() string
	To() string
	Type() string

	IsStanza() bool

	IsError() bool
	Error() XElement

	ToXML(w io.Writer, includeClosing bool) error
	ToBytes(buf *bytes.Buffer) error
}

// Stanza represents an XMPP stanza element.
type Stanza interface {
	XElement
	FromJID() *jid.JID
	ToJID() *jid.JID
}
