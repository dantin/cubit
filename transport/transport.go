package transport

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"time"

	"github.com/dantin/cubit/transport/compress"
)

// Type represents a stream transport type (socket).
type Type int

const (
	// Socket represents a socket transport type.
	Socket Type = iota + 1
)

func (tt Type) String() string {
	switch tt {
	case Socket:
		return "socket"
	}
	return ""
}

// ChannelBindingMechanism represents a scram channel binding mechanism.
type ChannelBindingMechanism int

const (
	// TLSUnique represents 'tls-unique' channel binding mechanism.
	TLSUnique ChannelBindingMechanism = iota
)

// Transport represents a stream transport mechanism.
type Transport interface {
	io.ReadWriteCloser

	// Type returns transport type value.
	Type() Type

	// WriteString writes a raw string to the transport.
	WriteString(s string) (n int, err error)

	// Flush writes any buffered data to the underlying io.Writer.
	Flush() error

	// SetWriteDeadline sets the deadline for future write calls.
	SetWriteDeadline(d time.Time) error

	// StartTLS secures the transport using SSL/TLS.
	StartTLS(cfg *tls.Config, asClient bool)

	// EnableCompression activates a compression mechanism on the transport.
	EnableCompression(level compress.Level)

	// ChannelBindingBytes returns current transport channel binding bytes.
	ChannelBindingBytes(mechanism ChannelBindingMechanism) []byte

	// PeerCertificates returns the certificate chain presented by remote peer.
	PeerCertificates() []*x509.Certificate
}

type tlsStateQueryable interface {
	ConnectionState() tls.ConnectionState
}
