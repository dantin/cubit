package s2srouter

import (
	"context"
	"sync"

	"github.com/dantin/cubit/router"
	"github.com/dantin/cubit/xmpp"
)

type s2sRouter struct {
	mu          sync.RWMutex
	outProvider OutProvider
	remotes     map[string]*remoteRouter
}

// New creates a router between two servers.
func New(outProvider OutProvider) router.S2SRouter {
	return &s2sRouter{
		outProvider: outProvider,
		remotes:     make(map[string]*remoteRouter),
	}
}

func (r *s2sRouter) Route(ctx context.Context, stanza xmpp.Stanza, localDomain string) error {
	remoteDomain := stanza.ToJID().Domain()

	r.mu.RLock()
	rr := r.remotes[remoteDomain]
	r.mu.RUnlock()

	if rr == nil {
		r.mu.Lock()
		rr = r.remotes[remoteDomain] // avoid double initialization
		if rr == nil {
			rr = newRemoteRouter(localDomain, remoteDomain, r.outProvider)
			r.remotes[remoteDomain] = rr
		}
		r.mu.Unlock()
	}
	rr.route(ctx, stanza)

	return nil
}
