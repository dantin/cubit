package jid

import (
	"bytes"
	"encoding/gob"
	"errors"
	"net"
	"strings"
	"unicode/utf8"

	"github.com/dantin/cubit/util/pool"
	"golang.org/x/net/idna"
	"golang.org/x/text/secure/precis"
)

var bufPool = pool.NewBufferPool()

// MatchingOptions represents a matching jid mask.
type MatchingOptions int8

const (
	// MatchesNode indicates that left and right operand has same node value.
	MatchesNode = MatchingOptions(1)

	// MatchesDomain indicates that left and right operand has same domain value.
	MatchesDomain = MatchingOptions(2)

	// MatchesResource indicates that left and right operand has same resource value.
	MatchesResource = MatchingOptions(4)

	// MatchesBare indicates that left and right operand has same node and domain value.
	MatchesBare = MatchesNode | MatchesDomain

	// MatchesFull indicates that left and right operand has same node, domain and resource value.
	MatchesFull = MatchesNode | MatchesDomain | MatchesResource
)

// JID represents an XMPP address (JID).
// A JID is made up of a node (generally a username), a domain, and a resource.
// The node and resource are optional; domain is required.
type JID struct {
	node     string
	domain   string
	resource string
}

// New constructs a JID given a user, domain, and resource.
// This construction allows the caller to specify if stringprep should be applied or not.
func New(node, domain, resource string, skipStringPrep bool) (*JID, error) {
	if skipStringPrep {
		return &JID{
			node:     node,
			domain:   domain,
			resource: resource,
		}, nil
	}
	var j JID
	if err := j.stringPrep(node, domain, resource); err != nil {
		return nil, err
	}
	return &j, nil
}

// NewWithString constructs a JID from it's string representation.
// This construction allows the caller to specify if stringprep should be applied or not.
func NewWithString(str string, skipStringPrep bool) (*JID, error) {
	if len(str) == 0 {
		return &JID{}, nil
	}
	var node, domain, resource string

	atIndex := strings.Index(str, "@")
	slashIndex := strings.Index(str, "/")

	// node
	if atIndex > 0 {
		node = str[0:atIndex]
	}

	// domain
	if atIndex+1 == len(str) {
		return nil, errors.New("JID with empty domain not valid")
	}
	if atIndex < 0 {
		if slashIndex > 0 {
			domain = str[0:slashIndex]
		} else {
			domain = str
		}
	} else {
		if slashIndex > 0 {
			domain = str[atIndex+1 : slashIndex]
		} else {
			domain = str[atIndex+1:]
		}
	}

	// resource
	if slashIndex > 0 {
		if slashIndex+1 < len(str) {
			resource = str[slashIndex+1:]
		} else {
			return nil, errors.New("JID resource must not be empty")
		}
	}
	return New(node, domain, resource, skipStringPrep)
}

// NewFromBytes constructs a JID from it's gob binary representation.
func NewFromBytes(buf *bytes.Buffer) (*JID, error) {
	var j JID
	if err := j.FromBytes(buf); err != nil {
		return nil, err
	}
	return &j, nil
}

// Node returns the node, or empty string if this JID does not contain node information.
func (j *JID) Node() string {
	return j.node
}

// Domain returns the domain.
func (j *JID) Domain() string {
	return j.domain
}

// Resource returns the resource, or empty string if this JID does not contain resource information.
func (j *JID) Resource() string {
	return j.resource
}

// ToBareJID returns the JID equivalent of the bare JID, which is the JID with resource information removed.
func (j *JID) ToBareJID() *JID {
	if len(j.node) == 0 {
		return &JID{node: "", domain: j.domain, resource: ""}
	}
	return &JID{node: j.node, domain: j.domain, resource: ""}

}

// IsServer returns true if instance is a server JID.
func (j *JID) IsServer() bool {
	return len(j.node) == 0
}

// IsBare returns true if instance is a bare JID.
func (j *JID) IsBare() bool {
	return len(j.node) > 0 && len(j.resource) == 0
}

// IsFull returns true if instance is a full JID.
func (j *JID) IsFull() bool {
	return len(j.resource) > 0
}

// IsFullWithServer returns true if instance is a full server JID.
func (j *JID) IsFullWithServer() bool {
	return len(j.node) == 0 && len(j.resource) > 0
}

// IsFullWithUser returns true if instance is a full client JID.
func (j *JID) IsFullWithUser() bool {
	return len(j.node) > 0 && len(j.resource) > 0
}

// Matches tells whether or not j2 matches j.
func (j *JID) Matches(j2 *JID) bool {
	if j.IsFullWithUser() {
		return j.MatchesWithOptions(j2, MatchesNode|MatchesDomain|MatchesResource)
	} else if j.IsFullWithServer() {
		return j.MatchesWithOptions(j2, MatchesDomain|MatchesResource)
	} else if j.IsBare() {
		return j.MatchesWithOptions(j2, MatchesNode|MatchesDomain)
	}
	return j.MatchesWithOptions(j2, MatchesDomain)
}

// MatchesWithOptions tells whether two jids are equivalent based on matching options.
func (j *JID) MatchesWithOptions(j2 *JID, options MatchingOptions) bool {
	if (options&MatchesNode) > 0 && j.node != j2.node {
		return false
	}
	if (options&MatchesDomain) > 0 && j.domain != j2.domain {
		return false
	}
	if (options&MatchesResource) > 0 && j.resource != j2.resource {
		return false
	}
	return true
}

// String returns a string representation of the JID.
func (j *JID) String() string {
	buf := bufPool.Get()
	defer bufPool.Put(buf)
	if len(j.node) > 0 {
		buf.WriteString(j.node)
		buf.WriteString("@")
	}
	buf.WriteString(j.domain)
	if len(j.resource) > 0 {
		buf.WriteString("/")
		buf.WriteString(j.resource)
	}
	return buf.String()
}

// FromBytes deserializes a JID entity from it's gob binary representation.
func (j *JID) FromBytes(buf *bytes.Buffer) error {
	dec := gob.NewDecoder(buf)
	var node, domain, resource string
	if err := dec.Decode(&node); err != nil {
		return err
	}
	if err := dec.Decode(&domain); err != nil {
		return err
	}
	if err := dec.Decode(&resource); err != nil {
		return err
	}
	return j.stringPrep(node, domain, resource)
}

// ToBytes converts a JID entity to it's gob binary representation.
func (j *JID) ToBytes(buf *bytes.Buffer) error {
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(&j.node); err != nil {
		return err
	}
	if err := enc.Encode(&j.domain); err != nil {
		return err
	}
	if err := enc.Encode(&j.resource); err != nil {
		return err
	}
	return nil
}

func (j *JID) stringPrep(node, domain, resource string) error {
	// Ensure that parts are valid UTF-8 (and short circuit the rest of the
	// process if they're not). We'll check the domain after performing
	// the IDNA ToUnicode operation.
	if !utf8.ValidString(node) || !utf8.ValidString(resource) {
		return errors.New("JID contains invalid UTF8")
	}

	// RFC 7622 §3.2.1.  Preparation
	//
	//    An entity that prepares a string for inclusion in an XMPP domain
	//    slot MUST ensure that the string consists only of Unicode code points
	//    that are allowed in NR-LDH labels or U-labels as defined in
	//    [RFC5890].  This implies that the string MUST NOT include A-labels as
	//    defined in [RFC5890]; each A-label MUST be converted to a U-label
	//    during preparation of a string for inclusion in a domain slot.
	var err error
	domain, err = idna.ToUnicode(domain)
	if err != nil {
		return err
	}
	if !utf8.ValidString(domain) {
		return errors.New("domain contains invalid UTF8")
	}

	// RFC 7622 §3.2.2.  Enforcement
	//
	//   An entity that performs enforcement in XMPP domain slots MUST
	//   prepare a string as described in Section 3.2.1 and MUST also apply
	//   the normalization, case-mapping, and width-mapping rules defined in
	//   [RFC5892].
	//
	var nodeLen int
	data := make([]byte, 0, len(node)+len(domain)+len(resource))

	if node != "" {
		data, err = precis.UsernameCaseMapped.Append(data, []byte(node))
		if err != nil {
			return err
		}
		nodeLen = len(data)
	}
	data = append(data, []byte(domain)...)

	if resource != "" {
		data, err = precis.OpaqueString.Append(data, []byte(resource))
		if err != nil {
			return err
		}
	}
	if err := commonChecks(data[:nodeLen], domain, data[nodeLen+len(domain):]); err != nil {
		return err
	}
	j.node = string(data[:nodeLen])
	j.domain = string(data[nodeLen : nodeLen+len(domain)])
	j.resource = string(data[nodeLen+len(domain):])
	return nil
}

func commonChecks(node []byte, domain string, resource []byte) error {
	l := len(node)
	if l > 1023 {
		return errors.New("node must be smaller than 1024 bytes")
	}

	// RFC 7622 §3.3.1 provides a small table of characters which are still not
	// allowed in node's even though the IdentifierClass base class and the
	// UsernameCaseMapped profile don't forbid them; disallow them here.
	if bytes.ContainsAny(node, `"&'/:<>@`) {
		return errors.New("node contains forbidden character")
	}

	l = len(resource)
	if l > 1023 {
		return errors.New("resource must be smaller than 1024 bytes")
	}

	l = len(domain)
	if l < 1 || l > 1023 {
		return errors.New("domain must be between 1 and 1023 bytes")
	}

	return checkIP6String(domain)
}

func checkIP6String(domain string) error {
	if l := len(domain); l > 2 && strings.HasPrefix(domain, "[") &&
		strings.HasSuffix(domain, "]") {
		if ip := net.ParseIP(domain[1 : l-1]); ip == nil || ip.To4() != nil {
			return errors.New("domain is not a valid IPv6 address")
		}
	}
	return nil
}
