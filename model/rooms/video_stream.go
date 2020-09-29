package roomsmodel

import (
	"bytes"
	"encoding/gob"
)

// VideoType represents a video stream type.
type VideoType int

const (
	// Unknown represents a unknown video type.
	Unknown VideoType = iota
	// Box represents a video encoding/decoding box.
	Box
	// Device represents a device video stream.
	Device
	// Camera represents a device video stream.
	Camera
)

func (tt VideoType) String() string {
	switch tt {
	case Box:
		return "box"
	case Device:
		return "device"
	case Camera:
		return "camera"
	default:
		return ""
	}
}

// ParseVideoTypeString convert a string to VideoType.
func ParseVideoTypeString(t string) VideoType {
	switch t {
	case "box":
		return Box
	case "device":
		return Device
	case "camera":
		return Camera
	default:
		return Unknown
	}
}

// VideoStream represents presence video stream info.
type VideoStream struct {
	In        string
	Broadcast string
	Stream    string

	Route  string
	Type   VideoType
	RoomID int
}

// FromBytes deserializes a VideoStream entiry from its binary representation.
func (v *VideoStream) FromBytes(buf *bytes.Buffer) error {
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&v.In); err != nil {
		return err
	}
	if err := dec.Decode(&v.Broadcast); err != nil {
		return err
	}
	return dec.Decode(&v.Route)
}

// ToBytes converts a VideoStream entiry to its binary representation.
func (v *VideoStream) ToBytes(buf *bytes.Buffer) error {
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(&v.In); err != nil {
		return err
	}
	if err := enc.Encode(&v.Broadcast); err != nil {
		return err
	}
	return enc.Encode(&v.Route)
}
