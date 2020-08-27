package transport

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTypeStrings(t *testing.T) {
	require.Equal(t, "socket", Socket.String())
	require.Equal(t, "", Type(9).String())
}
