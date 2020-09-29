package roomsmodel

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVideoStream_Serialization(t *testing.T) {
	vs := VideoStream{}
	vs.In = "room"
	vs.Broadcast = "broadcast"
	vs.Route = "stream map in JSON"

	buf := bytes.NewBuffer(nil)
	require.Nil(t, vs.ToBytes(buf))

	vs2 := VideoStream{}
	_ = vs2.FromBytes(buf)

	require.True(t, reflect.DeepEqual(&vs, &vs2))
}
