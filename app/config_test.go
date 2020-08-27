package app

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	var cfg1, cfg2 Config
	b, err := ioutil.ReadFile("../misc/testdata/config_basic.yml")
	require.Nil(t, err)
	err = cfg1.FromBuffer(bytes.NewBuffer(b))
	require.Nil(t, err)
	cfg2.FromFile("../misc/testdata/config_basic.yml")
	require.Equal(t, cfg1, cfg2)
}

func TestBadConfig(t *testing.T) {
	var cfg Config
	err := cfg.FromFile("../misc/testdata/not_a_config.yml")
	require.NotNil(t, err)
}
