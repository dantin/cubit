package compress

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTypeStrings(t *testing.T) {
	require.Equal(t, "default", DefaultCompression.String())
	require.Equal(t, "best", BestCompression.String())
	require.Equal(t, "speed", SpeedCompression.String())
	require.Equal(t, "", Level(99).String())
}
