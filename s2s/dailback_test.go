package s2s

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDialBack(t *testing.T) {
	secret := "secret4dailback"
	from := "alice"
	to := "bob"
	streamID := "007"
	kg := &keyGen{secret: secret}
	require.Equal(t, "0d8435ab3b65f3a5afd69466c25c148acc2e1c6765dbde41c74495e8969d0908", kg.generate(from, to, streamID))
}
