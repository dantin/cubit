package s2s

import (
	"context"
	"crypto/tls"
	"sync"

	"github.com/dantin/cubit/log"
	"github.com/dantin/cubit/router/host"
	"github.com/dantin/cubit/stream"
)

type newOutFunc = func(localDomain, remoteDomain string) *outStream

// OutProvider represents out stream provider.
type OutProvider struct {
	cfg            *Config
	hosts          *host.Hosts
	dialer         Dialer
	mu             sync.RWMutex
	outConnections map[string]stream.S2SOut
}

// NewOutProvider returns a new OutProvider.
func NewOutProvider(config *Config, hosts *host.Hosts) *OutProvider {
	return &OutProvider{
		cfg:            config,
		hosts:          hosts,
		dialer:         newDialer(),
		outConnections: make(map[string]stream.S2SOut),
	}
}

// GetOut returns an out stream.
func (p *OutProvider) GetOut(localDomain, remoteDomain string) stream.S2SOut {
	domainPair := getDomainPair(localDomain, remoteDomain)
	p.mu.RLock()
	outStm := p.outConnections[domainPair]
	p.mu.RUnlock()

	if outStm != nil {
		return outStm
	}
	p.mu.Lock()
	outStm = p.outConnections[domainPair] // 2nd check
	if outStm != nil {
		p.mu.Unlock()
		return outStm
	}
	outStm = p.newOut(localDomain, remoteDomain)
	p.outConnections[domainPair] = outStm
	p.mu.Unlock()

	log.Infof("registered s2s out stream... (domainpair: %s)", domainPair)

	return outStm
}

// Shutdown close server out stream.
func (p *OutProvider) Shutdown(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, conn := range p.outConnections {
		conn.Disconnect(ctx, nil)
	}
	p.outConnections = nil

	log.Infof("closed %d out connection(s)", len(p.outConnections))

	return nil
}

func (p *OutProvider) newOut(localDomain, remoteDomain string) *outStream {
	tlsConfig := &tls.Config{
		ServerName:   remoteDomain,
		Certificates: p.hosts.Certificates(),
	}
	cfg := &outConfig{
		keyGen:        &keyGen{secret: p.cfg.DialbackSecret},
		timeout:       p.cfg.Timeout,
		localDomain:   localDomain,
		remoteDomain:  remoteDomain,
		keepAlive:     p.cfg.KeepAlive,
		tls:           tlsConfig,
		maxStanzaSize: p.cfg.MaxStanzaSize,
	}
	return newOutStream(cfg, p.hosts, p.dialer)
}

func getDomainPair(localDomain, remoteDomain string) string {
	return localDomain + ":" + remoteDomain
}
