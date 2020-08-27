package host

import (
	"crypto/tls"
	"sort"

	utiltls "github.com/dantin/cubit/util/tls"
)

const defaultDomain = "localhost"

// Hosts represents hosts configuation map.
type Hosts struct {
	defaultHostname string
	hosts           map[string]tls.Certificate
}

// New initializes configured hosts type and returns associated hosts.
func New(hostsConfig []Config) (*Hosts, error) {
	h := &Hosts{
		hosts: make(map[string]tls.Certificate),
	}
	if len(hostsConfig) > 0 {
		for i, host := range hostsConfig {
			if i == 0 {
				h.defaultHostname = host.Name
			}
			h.hosts[host.Name] = host.Certificate
		}
	} else {
		cer, err := utiltls.LoadCertificate("", "", defaultDomain)
		if err != nil {
			return nil, err
		}
		h.defaultHostname = defaultDomain
		h.hosts[defaultDomain] = cer
	}
	return h, nil
}

// DefaultHostName returns the default hostname.
func (h *Hosts) DefaultHostName() string {
	return h.defaultHostname
}

// IsLocalHost returns whether a domain is hosted.
func (h *Hosts) IsLocalHost(domain string) bool {
	_, ok := h.hosts[domain]
	return ok
}

// HostName returns a sorted list of hostname.
func (h *Hosts) HostName() []string {
	var ret []string
	for n := range h.hosts {
		ret = append(ret, n)
	}
	sort.Slice(ret, func(i, j int) bool { return ret[i] < ret[j] })
	return ret
}

// Certificates returns a list of TLS certificat.
func (h *Hosts) Certificates() []tls.Certificate {
	var certs []tls.Certificate
	for _, cer := range h.hosts {
		certs = append(certs, cer)
	}
	return certs
}
