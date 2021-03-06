package capsmodel

import (
	"bytes"
	"encoding/gob"
)

// Capabilities represents presence capabilities info.
type Capabilities struct {
	Node     string
	Ver      string
	Features []string
}

// HasFeature returns whether or not Capabilities instance a concrete feature.
func (c *Capabilities) HasFeature(feature string) bool {
	for _, f := range c.Features {
		if f == feature {
			return true
		}
	}
	return false
}

// FromBytes deserializes a Capabilities entiry from its binary representation.
func (c *Capabilities) FromBytes(buf *bytes.Buffer) error {
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&c.Node); err != nil {
		return err
	}
	if err := dec.Decode(&c.Ver); err != nil {
		return err
	}
	return dec.Decode(&c.Features)
}

// ToBytes converts a Capabilities entiry to its binary representation.
func (c *Capabilities) ToBytes(buf *bytes.Buffer) error {
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(&c.Node); err != nil {
		return err
	}
	if err := enc.Encode(&c.Ver); err != nil {
		return err
	}
	return enc.Encode(&c.Features)
}
