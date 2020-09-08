package s2srouter

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/dantin/cubit/stream"
	"github.com/dantin/cubit/xmpp"
	"github.com/dantin/cubit/xmpp/jid"
	"github.com/stretchr/testify/require"
)

type mockedOutS2S struct {
	sentTimes int32
}

func (s *mockedOutS2S) ID() string                            { return "out-test-1" }
func (s *mockedOutS2S) Disconnect(_ context.Context, _ error) {}
func (s *mockedOutS2S) SendElement(_ context.Context, _ xmpp.XElement) {
	atomic.AddInt32(&s.sentTimes, 1)
}

type mockedOutProvider struct {
	outStm *mockedOutS2S
}

func (p *mockedOutProvider) GetOut(_, _ string) stream.S2SOut { return p.outStm }

func TestS2SRouter_Route(t *testing.T) {
	outStm := &mockedOutS2S{}
	p := &mockedOutProvider{outStm: outStm}

	r := New(p)

	j1, _ := jid.NewWithString("alice@example.com", true)
	j2, _ := jid.NewWithString("bob@example.org/desktop", true)
	j3, _ := jid.NewWithString("alice@example.org/desktop", true)

	_ = r.Route(context.Background(), xmpp.NewPresence(j1, j2, xmpp.AvailableType), "example.com")
	_ = r.Route(context.Background(), xmpp.NewPresence(j1, j3, xmpp.AvailableType), "example.com")

	require.Equal(t, int32(2), atomic.LoadInt32(&outStm.sentTimes))
}
