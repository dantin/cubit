package model

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/dantin/cubit/xmpp"
)

// Role represents a user role type.
type Role int

const (
	// Unknown represents a unknown role.
	Unknown Role = iota
	// Root represents a root role.
	Root
	// Admin represents a admin role.
	Admin
	// Usr represents a user role.
	Usr
)

func (r Role) String() string {
	switch r {
	case Root:
		return "root"
	case Admin:
		return "admin"
	case Usr:
		return "user"
	}
	return "unknown"
}

// ParseRoleString parses string to Role.
func ParseRoleString(s string) Role {
	switch s {
	case "root":
		return Root
	case "admin":
		return Admin
	case "user":
		return Usr
	}
	return Unknown
}

// User represents a user storage entity.
type User struct {
	Username       string
	Password       string
	Role           Role
	LastPresence   *xmpp.Presence
	LastPresenceAt time.Time
}

// FromBytes deserializes a User entity from it's gob binary representation.
func (u *User) FromBytes(buf *bytes.Buffer) error {
	dec := gob.NewDecoder(buf)
	if err := dec.Decode(&u.Username); err != nil {
		return err
	}
	if err := dec.Decode(&u.Password); err != nil {
		return err
	}
	if err := dec.Decode(&u.Role); err != nil {
		return err
	}
	var hasPresence bool
	if err := dec.Decode(&hasPresence); err != nil {
		return err
	}
	if hasPresence {
		p, err := xmpp.NewPresenceFromBytes(buf)
		if err != nil {
			return err
		}
		u.LastPresence = p
		if err := dec.Decode(&u.LastPresenceAt); err != nil {
			return err
		}
	}
	return nil
}

// ToBytes converts a User entity to it's gob binary representation.
func (u *User) ToBytes(buf *bytes.Buffer) error {
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(&u.Username); err != nil {
		return err
	}
	if err := enc.Encode(&u.Password); err != nil {
		return err
	}
	if err := enc.Encode(&u.Role); err != nil {
		return err
	}
	hasPresence := u.LastPresence != nil
	if err := enc.Encode(&hasPresence); err != nil {
		return err
	}
	if hasPresence {
		if err := u.LastPresence.ToBytes(buf); err != nil {
			return err
		}
		u.LastPresenceAt = time.Now()
		return enc.Encode(&u.LastPresenceAt)
	}
	return nil
}
