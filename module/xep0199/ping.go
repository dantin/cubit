package xep0199

import (
	"context"
	"fmt"
	"sync"
	"time"

	streamerror "github.com/dantin/cubit/errors"
	"github.com/dantin/cubit/log"
	"github.com/dantin/cubit/module/xep0030"
	"github.com/dantin/cubit/router"
	"github.com/dantin/cubit/stream"
	"github.com/dantin/cubit/util/runqueue"
	"github.com/dantin/cubit/xmpp"
	"github.com/dantin/cubit/xmpp/jid"
	"github.com/google/uuid"
)

const pingNamespace = "urn:xmpp:ping"

const pingWriteTimeout = time.Second

// Config represents XMPP Ping module (XEP-0199) configuration.
type Config struct {
	Send         bool
	SendInterval time.Duration
}

type configProxy struct {
	Send         bool `yaml:"send"`
	SendInterval int  `yaml:"send_interval"`
}

// UnmarshalYAML satisfies Unmarshaler interface.
func (c *Config) UnmarshalYAML(unmarshal func(interface{}) error) error {
	p := configProxy{}
	if err := unmarshal(&p); err != nil {
		return err
	}
	c.Send = p.Send
	c.SendInterval = time.Second * time.Duration(p.SendInterval)
	if c.Send && c.SendInterval < time.Second {
		return fmt.Errorf("xep0199.Config: send interval must be 1 or higher")
	}
	return nil
}

type ping struct {
	identifier string
	timer      *time.Timer
	stm        stream.C2S
}

// Ping represents a ping server stream module.
type Ping struct {
	cfg           *Config
	router        router.Router
	pings         map[string]*ping
	activePingsMu sync.RWMutex
	activePings   map[string]*ping
	runQueue      *runqueue.RunQueue
}

// New returns an ping IQ handler module.
func New(config *Config, disco *xep0030.DiscoInfo, router router.Router) *Ping {
	p := &Ping{
		cfg:         config,
		router:      router,
		pings:       make(map[string]*ping),
		activePings: make(map[string]*ping),
		runQueue:    runqueue.New("xep0199"),
	}
	if disco != nil {
		disco.RegisterServerFeature(pingNamespace)
		disco.RegisterAccountFeature(pingNamespace)
	}
	return p
}

// MatchesIQ returns whether or not an IQ should be processed by the ping module.
func (x *Ping) MatchesIQ(iq *xmpp.IQ) bool {
	return x.isPongIQ(iq) || iq.Elements().ChildNamespace("ping", pingNamespace) != nil
}

// ProcessIQ processes a ping IQ taking according actions over the associated stream.
func (x *Ping) ProcessIQ(ctx context.Context, iq *xmpp.IQ) {
	x.runQueue.Run(func() {
		stm := x.router.LocalStream(iq.FromJID().Node(), iq.FromJID().Resource())
		if stm == nil {
			return
		}
		x.processIQ(ctx, iq, stm)
	})
}

// SchedulePing schedules a new ping in a 'send interval' period, cancelling previous scheduled ping.
func (x *Ping) SchedulePing(stm stream.C2S) {
	x.runQueue.Run(func() { x.schedulePing(stm) })
}

// CancelPing cancels a previous scheduled ping.
func (x *Ping) CancelPing(stm stream.C2S) {
	x.runQueue.Run(func() { x.cancelPing(stm) })
}

// Shutdown shuts down ping module.
func (x *Ping) Shutdown() error {
	c := make(chan struct{})
	x.runQueue.Stop(func() {
		for _, pi := range x.pings {
			pi.timer.Stop()
		}
		close(c)
	})
	<-c
	return nil
}

func (x *Ping) processIQ(ctx context.Context, iq *xmpp.IQ, stm stream.C2S) {
	if x.isPongIQ(iq) {
		x.handlePongIQ(iq, stm)
		return
	}
	toJid := iq.ToJID()
	if !toJid.IsServer() && toJid.Node() != stm.Username() {
		stm.SendElement(ctx, iq.ForbiddenError())
		return
	}
	p := iq.Elements().ChildNamespace("ping", pingNamespace)
	if p == nil || p.Elements().Count() > 0 {
		stm.SendElement(ctx, iq.BadRequestError())
		return
	}
	log.Infof("received ping... id: %s", iq.ID())
	if iq.IsGet() {
		log.Infof("sent pong... id: %s", iq.ID())
		stm.SendElement(ctx, iq.ResultIQ())
	} else {
		stm.SendElement(ctx, iq.BadRequestError())
	}
}

func (x *Ping) schedulePing(stm stream.C2S) {
	if !x.cfg.Send || !stm.JID().IsFull() {
		return
	}
	userJID := stm.JID().String()

	if pi := x.pings[userJID]; pi != nil {
		if _, ok := x.activePings[pi.identifier]; ok {
			// waiting for pong
			return
		}
		// cancel previous ping
		pi.timer.Stop()
	}
	x.schedulePingTimer(stm)
}

func (x *Ping) cancelPing(stm stream.C2S) {
	if !x.cfg.Send || !stm.JID().IsFull() {
		return
	}
	userJID := stm.JID().String()

	if pi := x.pings[userJID]; pi != nil {
		pi.timer.Stop()

		delete(x.pings, userJID)
		delete(x.activePings, pi.identifier)
	}
}

func (x *Ping) schedulePingTimer(stm stream.C2S) {
	pi := &ping{
		identifier: uuid.New().String(),
		stm:        stm,
	}
	pi.timer = time.AfterFunc(x.cfg.SendInterval, func() {
		x.runQueue.Run(func() {
			x.sendPing(pi)
		})
	})
	x.pings[stm.JID().String()] = pi
}

func (x *Ping) handlePongIQ(iq *xmpp.IQ, stm stream.C2S) {
	pongID := iq.ID()
	if pi := x.activePings[pongID]; pi != nil && pi.stm == stm {
		log.Infof("received pong... id: %s", pongID)

		pi.timer.Stop()
		x.schedulePingTimer(stm)
	}
}

func (x *Ping) sendPing(pi *ping) {
	//ctx, _ := context.WithTimeout(context.Background(), pingWriteTimeout)
	ctx := context.TODO()

	srvJID, _ := jid.New("", pi.stm.JID().Domain(), "", true)

	iq := xmpp.NewIQType(pi.identifier, xmpp.GetType)
	iq.SetFromJID(srvJID)
	iq.SetToJID(pi.stm.JID())
	iq.AppendElement(xmpp.NewElementNamespace("ping", pingNamespace))

	pi.stm.SendElement(ctx, iq)

	log.Infof("sent ping... id: %s", pi.identifier)

	pi.timer = time.AfterFunc(x.cfg.SendInterval/3, func() {
		x.runQueue.Run(func() {
			x.disconnectStream(pi)
		})
	})
	x.activePingsMu.Lock()
	x.activePings[pi.identifier] = pi
	x.activePingsMu.Unlock()
}

func (x *Ping) disconnectStream(pi *ping) {
	//ctx, _ := context.WithTimeout(context.Background(), pingWriteTimeout)
	ctx := context.TODO()
	pi.stm.Disconnect(ctx, streamerror.ErrConnectionTimeout)
}

func (x *Ping) isPongIQ(iq *xmpp.IQ) bool {
	x.activePingsMu.RLock()
	_, ok := x.activePings[iq.ID()]
	x.activePingsMu.RUnlock()

	return ok && (iq.IsResult() || iq.Type() == xmpp.ErrorType)
}
