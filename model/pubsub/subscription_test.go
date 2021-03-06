package pubsubmodel

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSubscription_Serialization(t *testing.T) {
	s := Subscription{
		SubID:        uuid.New().String(),
		JID:          "username@example.org",
		Subscription: "subscribed",
	}
	b := bytes.NewBuffer(nil)
	require.Nil(t, s.ToBytes(b))

	var s2 Subscription
	require.Nil(t, s2.FromBytes(b))
	require.True(t, reflect.DeepEqual(s, s2))
}
