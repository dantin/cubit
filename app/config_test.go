package app

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig(t *testing.T) {
	var cfg1, cfg2 Config
	b, err := ioutil.ReadFile("../data/sample.yml")
	require.Nil(t, err)
	err = cfg1.FromBuffer(bytes.NewBuffer(b))
	require.Nil(t, err)
	cfg2.FromFile("../data/sample.yml")
	require.Equal(t, cfg1, cfg2)
}

func TestBadConfig(t *testing.T) {
	var cfg Config
	err := cfg.FromFile("../data/not_a_config.yml")
	require.NotNil(t, err)
}
