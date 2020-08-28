package utiltls

import (
	"crypto/tls"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadCertificate(t *testing.T) {
	t.Run("Valid", func(t *testing.T) {
		//defer os.RemoveAll(".cert/")

		tlsCfg, err := LoadCertificate("../../data/cert/test.server.key", "../../data/cert/test.server.crt", "localhost")

		require.Nil(t, err)
		require.NotNil(t, tlsCfg)
	})
	t.Run("SelfSigned", func(t *testing.T) {
		//defer os.RemoveAll(".cert/")

		tlsCfg, err := LoadCertificate("", "", "localhost")

		require.Nil(t, err)
		require.NotNil(t, tlsCfg)
	})

	t.Run("Failed", func(t *testing.T) {
		tlsCfg, err := LoadCertificate("", "", "example.org")

		require.Equal(t, tls.Certificate{}, tlsCfg)
		require.NotNil(t, err)
		require.Equal(t, "must specify a private key and a server certificate for the domain 'example.org'", err.Error())
	})
}
