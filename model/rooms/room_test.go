package roomsmodel

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRoom_Serialization(t *testing.T) {
	room := Room{}
	room.ID = 1
	room.Name = "room"
	room.Username = "alice"
	room.Streams = []VideoStream{
		VideoStream{In: "camera_in", Broadcast: "camera_broadcast", Route: "camera_route_JSON"},
		VideoStream{In: "device_in", Broadcast: "device_broadcast", Route: "device_route_JSON"},
	}

	buf := bytes.NewBuffer(nil)
	require.Nil(t, room.ToBytes(buf))

	room2 := Room{}
	_ = room2.FromBytes(buf)

	require.True(t, reflect.DeepEqual(&room, &room2))
}
