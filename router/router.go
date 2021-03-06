package router

import (
	"context"

	"github.com/dantin/cubit/router/host"
	"github.com/dantin/cubit/stream"
	"github.com/dantin/cubit/xmpp"
	"github.com/dantin/cubit/xmpp/jid"
)

// Router represents a router between two entities.
type Router interface {
	// Hosts returns router hosts container.
	Hosts() *host.Hosts

	// Route routes a stanza applying server rules for handling XML stanzas.
	// (https://xmpp.org/rfcs/rfc3921.html#rules)
	Route(ctx context.Context, stanza xmpp.Stanza) error

	// MustRoute forces stanza routing by ignoring user's blocking list.
	MustRoute(ctx context.Context, stanza xmpp.Stanza) error

	// Bind sets a c2s stream as bound.
	Bind(ctx context.Context, stm stream.C2S)

	// Unbind unbinds a previously bound c2s stream.
	Unbind(ctx context.Context, j *jid.JID)

	// LocalStream returns the steam associated to a given username and resource.
	LocalStream(username, resource string) stream.C2S

	// LocalStreams returns all steams associated to a given username.
	LocalStreams(username string) []stream.C2S
}

// C2SRouter represents a router between client and server.
type C2SRouter interface {
	// Route routes a stanza applying server rules for handling XML stanzas.
	// (https://xmpp.org/rfcs/rfc3921.html#rules)
	Route(ctx context.Context, stanza xmpp.Stanza, validateStanza bool) error

	// Bind sets a c2s stream as bound.
	Bind(stm stream.C2S)

	// Unbind unbinds a previously bound c2s stream.
	Unbind(username, resource string)

	// Stream returns the stream associated to a given username and resource.
	Stream(username, resource string) stream.C2S

	// Streams returns all steams associated to a given username.
	Streams(username string) []stream.C2S
}

// S2SRouter represents a router between server and server.
type S2SRouter interface {
	// Route routes a stanza applying server rules for handling XML stanzas.
	// (https://xmpp.org/rfcs/rfc3921.html#rules)
	Route(ctx context.Context, stanza xmpp.Stanza, localDomain string) error
}

type router struct {
	hosts *host.Hosts
	c2s   C2SRouter
	s2s   S2SRouter
}

// New creates a new router.
func New(hosts *host.Hosts, c2sRouter C2SRouter, s2sRouter S2SRouter) (Router, error) {
	r := &router{
		hosts: hosts,
		c2s:   c2sRouter,
		s2s:   s2sRouter,
	}
	return r, nil
}

func (r *router) Hosts() *host.Hosts {
	return r.hosts
}

func (r *router) Route(ctx context.Context, stanza xmpp.Stanza) error {
	return r.route(ctx, stanza, true)
}

func (r *router) MustRoute(ctx context.Context, stanza xmpp.Stanza) error {
	return r.route(ctx, stanza, false)
}

func (r *router) Bind(ctx context.Context, stm stream.C2S) {
	r.c2s.Bind(stm)
}

func (r *router) Unbind(ctx context.Context, j *jid.JID) {
	r.c2s.Unbind(j.Node(), j.Resource())
}

func (r *router) LocalStream(username, resource string) stream.C2S {
	return r.c2s.Stream(username, resource)
}

func (r *router) LocalStreams(username string) []stream.C2S {
	return r.c2s.Streams(username)
}

func (r *router) route(ctx context.Context, stanza xmpp.Stanza, validateStanza bool) error {
	toJID := stanza.ToJID()
	if !r.hosts.IsLocalHost(toJID.Domain()) {
		if r.s2s == nil {
			return ErrFailedRemoteConnect
		}
		return r.s2s.Route(ctx, stanza, r.hosts.DefaultHostName())
	}
	return r.c2s.Route(ctx, stanza, validateStanza)
}
