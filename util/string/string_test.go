package utilstring

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSplitKeyAndValue(t *testing.T) {
	key, value := SplitKeyAndValue("key=value", '=')
	require.Equal(t, "key", key)
	require.Equal(t, "value", value)

	key, value = SplitKeyAndValue("nosep", '=')
	require.Equal(t, "", key)
	require.Equal(t, "", value)
}
