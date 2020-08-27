package xmpp_test

import (
	"testing"

	"github.com/dantin/cubit/xmpp"
	"github.com/stretchr/testify/require"
)

func TestDelay(t *testing.T) {
	e := xmpp.NewElementName("element")
	e.Delay("example.org", "test text")
	delay := e.Elements().Child("delay")
	require.NotNil(t, delay)
	require.Equal(t, "example.org", delay.Attributes().Get("from"))
	require.Equal(t, "test text", delay.Text())
}
