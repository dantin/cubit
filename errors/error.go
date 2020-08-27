package streamerror

import (
	"github.com/dantin/cubit/xmpp"
)

// Error represents a "stream:error" element.
type Error struct {
	reason string
}

var (
	// ErrInvalidXML represents 'invalid-xml' stream error.
	ErrInvalidXML = newStreamError("invalid-xml")

	// ErrInvalidNamespace represents 'invalid-namespace' stream error.
	ErrInvalidNamespace = newStreamError("invalid-namespace")

	// ErrHostUnknown represents 'host-unknown' stream error.
	ErrHostUnknown = newStreamError("host-unknown")

	// ErrInvalidFrom represents 'invalid-from' stream error.
	ErrInvalidFrom = newStreamError("invalid-from")

	// ErrPolicyViolation represents 'connection-timeout' stream error.
	ErrPolicyViolation = newStreamError("policy-violation")

	// ErrRemoteConnectionFailed represents 'remote-connection-failed' stream error.
	ErrRemoteConnectionFailed = newStreamError("remote-connection-failed")

	// ErrConnectionTimeout represents 'connection-timeout' stream error.
	ErrConnectionTimeout = newStreamError("connection-timeout")

	// ErrUnsupportedStanzaType represents 'unsupported-stanza-type' stream error.
	ErrUnsupportedStanzaType = newStreamError("unsupported-stanza-type")

	// ErrUnsupportedVersion represents 'unsupported-version' stream error.
	ErrUnsupportedVersion = newStreamError("unsupported-version")

	// ErrNotAuthorized represents 'not-authorized' stream error.
	ErrNotAuthorized = newStreamError("not-authorized")

	// ErrResourceConstraint represents 'resource-constraint' stream error.
	ErrResourceConstraint = newStreamError("resource-constraint")

	// ErrSystemShutdown represents 'system-shutdown' stream error.
	ErrSystemShutdown = newStreamError("system-shutdown")

	// ErrUndefinedCondition represents 'system-shutdown' stream error.
	ErrUndefinedCondition = newStreamError("undefined-condition")

	// ErrInternalServerError represents 'internal-server-error' stream error.
	ErrInternalServerError = newStreamError("internal-server-error")
)

func newStreamError(reason string) *Error {
	return &Error{reason: reason}
}

// Element returns stream error XML node.
func (se *Error) Element() xmpp.XElement {
	ret := xmpp.NewElementName("stream:error")
	reason := xmpp.NewElementNamespace(se.reason, "urn:ietf:params:xml:ns:xmpp-stream")
	ret.AppendElement(reason)
	return ret
}

// Error satisfies error interface.
func (se *Error) Error() string {
	return se.reason
}
