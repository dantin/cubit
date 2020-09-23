package model

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoom_Serialization(t *testing.T) {
	room := Room{}
	room.Username = "user"
	room.Camera = "camera_url"
	room.Device = "device_url"

	buf := bytes.NewBuffer(nil)
	require.Nil(t, room.ToBytes(buf))

	room2 := Room{}
	_ = room2.FromBytes(buf)

	require.True(t, reflect.DeepEqual(&room, &room2))
}
