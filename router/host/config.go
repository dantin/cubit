package host

import (
	"crypto/tls"

	utiltls "github.com/dantin/cubit/util/tls"
)

// TLSConfig represents a TLS certification configuration.
type TLSConfig struct {
	CertFile       string `yaml:"cert_path"`
	PrivateKeyFile string `yaml:"privkey_path"`
}

// Config represents a host configuration.
type Config struct {
	Name        string
	Certificate tls.Certificate
}

// configProxy represents a named TLS certification configuration.
type configProxy struct {
	Name string    `yaml:"name"`
	TLS  TLSConfig `yaml:"tls"`
}

// UnmarshalYAML satisfies Unmarshaler interface.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	p := configProxy{}
	if err := unmarshal(&p); err != nil {
		return err
	}
	c.Name = p.Name
	cer, err := utiltls.LoadCertificate(p.TLS.PrivateKeyFile, p.TLS.CertFile, c.Name)
	if err != nil {
		return err
	}
	c.Certificate = cer
	return nil
}
