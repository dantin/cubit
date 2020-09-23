package model

import (
	"bytes"
	"encoding/gob"
)

// Room represents presence room info.
type Room struct {
	Username string
	Camera   string
	Device   string
}

// FromBytes deserializes a Room entiry from its binary representation.
func (r *Room) FromBytes(buf *bytes.Buffer) error {
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&r.Username); err != nil {
		return err
	}
	if err := dec.Decode(&r.Camera); err != nil {
		return err
	}
	return dec.Decode(&r.Device)
}

// ToBytes converts a Room entiry to its binary representation.
func (r *Room) ToBytes(buf *bytes.Buffer) error {
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(&r.Username); err != nil {
		return err
	}
	if err := enc.Encode(&r.Camera); err != nil {
		return err
	}
	return enc.Encode(&r.Device)
}
