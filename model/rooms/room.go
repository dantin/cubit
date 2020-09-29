package roomsmodel

import (
	"bytes"
	"encoding/gob"
)

// RoomType represents a room type.
type RoomType int

const (
	// Normal represents a normal room.
	Normal RoomType = iota + 1
	// QC represents a QC room.
	QC
)

func (tt RoomType) String() string {
	switch tt {
	case Normal:
		return "normal"
	case QC:
		return "qc"
	}
	return ""
}

// Room represents presence room info.
type Room struct {
	ID       int
	Name     string
	Username string
	Streams  []VideoStream
}

// FromBytes deserializes a Room entiry from its binary representation.
func (r *Room) FromBytes(buf *bytes.Buffer) error {
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&r.ID); err != nil {
		return err
	}
	if err := dec.Decode(&r.Name); err != nil {
		return err
	}
	if err := dec.Decode(&r.Username); err != nil {
		return err
	}
	return dec.Decode(&r.Streams)
}

// ToBytes converts a Room entiry to its binary representation.
func (r *Room) ToBytes(buf *bytes.Buffer) error {
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(&r.ID); err != nil {
		return err
	}
	if err := enc.Encode(&r.Name); err != nil {
		return err
	}
	if err := enc.Encode(&r.Username); err != nil {
		return err
	}
	return enc.Encode(&r.Streams)
}
