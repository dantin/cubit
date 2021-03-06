package xmpp

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io"

	"github.com/dantin/cubit/xmpp/jid"
)

const (
	// MessageName represents "message" stanza name.
	MessageName = "message"

	// PresenceName represents "presence" stanza name.
	PresenceName = "presence"

	// IQName represents "iq" stanza name.
	IQName = "iq"
)

// Element represents a generic and mutable XML node element.
type Element struct {
	name     string
	text     string
	attrs    attributeSet
	elements elementSet
}

// NewElementName creates a mutable XML XElement instance with a given name.
func NewElementName(name string) *Element {
	return &Element{name: name}
}

// NewElementNamespace create a mutable XML XElement instance with a given name and namespace.
func NewElementNamespace(name, namespace string) *Element {
	return &Element{
		name:  name,
		attrs: attributeSet([]Attribute{{"xmlns", namespace}}),
	}
}

// NewElementFromElement creates a mutable XML XElement by copying an element.
func NewElementFromElement(elem XElement) *Element {
	e := &Element{}
	e.copyFrom(elem)
	return e
}

// NewErrorStanzaFromStanza returns a copy of an element of stanza error class.
func NewErrorStanzaFromStanza(stanza Stanza, stanzaErr *StanzaError, errorElements []XElement) Stanza {
	e := &stanzaElement{}
	e.copyFrom(stanza)
	e.SetType(ErrorType)
	e.SetFromJID(stanza.ToJID())
	e.SetToJID(stanza.FromJID())
	errEl := stanzaErr.Element()
	errEl.AppendElements(errorElements)
	e.AppendElement(errEl)
	return e
}

// NewElementFromBytes createsand returns a new Element from its bytes representation.
func NewElementFromBytes(buf *bytes.Buffer) (*Element, error) {
	e := &Element{}
	if err := e.FromBytes(buf); err != nil {
		return nil, err
	}
	return e, nil
}

// Name returns XML node name.
func (e *Element) Name() string {
	return e.name
}

// Attributes returns XML node attribute value.
func (e *Element) Attributes() AttributeSet {
	return e.attrs
}

// Elements returns all instance's child elements.
func (e *Element) Elements() ElementSet {
	return e.elements
}

// Text returns XML node text value.
// Returns an empty string if not set.
func (e *Element) Text() string {
	return e.text
}

// ID returns 'id' node attribute.
func (e *Element) ID() string {
	return e.attrs.Get("id")
}

// Namespace returns 'xmlns' node attribute.
func (e *Element) Namespace() string {
	return e.attrs.Get("xmlns")
}

// Language returns 'xml:lang' node attribute.
func (e *Element) Language() string {
	return e.attrs.Get("xml:lang")
}

// Version returns 'version' node attribute.
func (e *Element) Version() string {
	return e.attrs.Get("version")
}

// From returns 'from' node attribute.
func (e *Element) From() string {
	return e.attrs.Get("from")
}

// To returns 'to' node attribute.
func (e *Element) To() string {
	return e.attrs.Get("to")
}

// Type returns 'type' node attribute.
func (e *Element) Type() string {
	return e.attrs.Get("type")
}

// IsStanza returns true if element is an XMPP stanza.
func (e *Element) IsStanza() bool {
	switch e.Name() {
	case "iq", "presence", "message":
		return true
	}
	return false
}

// IsError returns true if element has a 'type' attribute of value 'error'.
func (e *Element) IsError() bool {
	return e.Type() == ErrorType
}

// Error returns element error sub element.
func (e *Element) Error() XElement {
	return e.elements.Child("error")
}

// String returns a string representation of the element.
func (e *Element) String() string {
	buf := bufPool.Get()
	defer bufPool.Put(buf)

	_ = e.ToXML(buf, true)
	return buf.String()
}

// ToXML serializes element to a row XML representation.
// includeClosing determines if closing tag should be attached.
func (e *Element) ToXML(w io.Writer, includeClosing bool) error {
	if _, err := io.WriteString(w, "<"); err != nil {
		return err
	}
	if _, err := io.WriteString(w, e.name); err != nil {
		return err
	}

	// serialize attributes
	for _, attr := range e.attrs {
		if len(attr.Value) == 0 {
			continue
		}
		if _, err := io.WriteString(w, ` `); err != nil {
			return err
		}
		if _, err := io.WriteString(w, attr.Label); err != nil {
			return err
		}
		if _, err := io.WriteString(w, `="`); err != nil {
			return err
		}
		if _, err := io.WriteString(w, attr.Value); err != nil {
			return err
		}
		if _, err := io.WriteString(w, `"`); err != nil {
			return err
		}
	}

	// serialize elements
	if e.elements.Count() > 0 || len(e.text) > 0 {
		if _, err := io.WriteString(w, ">"); err != nil {
			return err
		}
		if len(e.text) > 0 {
			if err := escapeText(w, []byte(e.text), false); err != nil {
				return err
			}
		}
		for _, elem := range e.elements {
			if err := elem.ToXML(w, true); err != nil {
				return err
			}
		}

		if includeClosing {
			if _, err := io.WriteString(w, "</"); err != nil {
				return err
			}
			if _, err := io.WriteString(w, e.name); err != nil {
				return err
			}
			if _, err := io.WriteString(w, ">"); err != nil {
				return err
			}
		}
	} else {
		if includeClosing {
			if _, err := io.WriteString(w, "/>"); err != nil {
				return err
			}
		} else {
			if _, err := io.WriteString(w, ">"); err != nil {
				return err
			}

		}

	}
	return nil
}

// FromBytes deserializes an element node from it's gob binary representation.
func (e *Element) FromBytes(buf *bytes.Buffer) error {
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&e.name); err != nil {
		return err
	}
	if err := dec.Decode(&e.text); err != nil {
		return err
	}
	if err := e.attrs.FromBytes(buf); err != nil {
		return err
	}
	return e.elements.FromBytes(buf)
}

// ToBytes serializes an element node to it's gob binary representation.
func (e *Element) ToBytes(buf *bytes.Buffer) error {
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(&e.name); err != nil {
		return err
	}
	if err := enc.Encode(&e.text); err != nil {
		return err
	}
	if err := e.attrs.ToBytes(buf); err != nil {
		return err
	}
	return e.elements.ToBytes(buf)
}

func (e *Element) copyFrom(el XElement) {
	e.name = el.Name()
	e.text = el.Text()
	e.attrs.copyFrom(el.Attributes().(attributeSet))
	e.elements.copyFrom(el.Elements().(elementSet))
}

type stanzaElement struct {
	Element
	fromJID *jid.JID
	toJID   *jid.JID
}

// NewStanzaFromElement returns a new stanza instance derived from an XMPP element.
func NewStanzaFromElement(elem XElement) (Stanza, error) {
	fromJID, err := jid.NewWithString(elem.From(), false)
	if err != nil {
		return nil, err
	}
	toJID, err := jid.NewWithString(elem.To(), false)
	if err != nil {
		return nil, err
	}
	switch elem.Name() {
	case IQName:
		return NewIQFromElement(elem, fromJID, toJID)
	case PresenceName:
		return NewPresenceFromElement(elem, fromJID, toJID)
	case MessageName:
		return NewMessageFromElement(elem, fromJID, toJID)
	}
	return nil, fmt.Errorf("unrecognized stanza name: %s", elem.Name())
}

// ToJID returns iq 'from' JID value.
func (s *stanzaElement) ToJID() *jid.JID {
	return s.toJID
}

// SetToJID sets the IQ 'to' JID value.
func (s *stanzaElement) SetToJID(j *jid.JID) {
	s.toJID = j
	s.SetTo(j.String())
}

// FromJID returns presence 'from' JID value.
func (s *stanzaElement) FromJID() *jid.JID {
	return s.fromJID
}

// SetFromJID sets the IQ 'from' JID value.
func (s *stanzaElement) SetFromJID(j *jid.JID) {
	s.fromJID = j
	s.SetFrom(j.String())
}

// FromBytes deserializes a stanza element from it's gob binary representation.
func (s *stanzaElement) FromBytes(buf *bytes.Buffer) error {
	if err := s.Element.FromBytes(buf); err != nil {
		return err
	}

	// set from and to JIDs.
	fromJID, err := jid.NewWithString(s.From(), false)
	if err != nil {
		return err
	}
	toJID, err := jid.NewWithString(s.To(), false)
	if err != nil {
		return err
	}
	s.SetFromJID(fromJID)
	s.SetToJID(toJID)
	return nil
}
