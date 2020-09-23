package ultrasound

import (
	"context"
	"crypto/tls"
	"testing"

	c2srouter "github.com/dantin/cubit/c2s/router"
	"github.com/dantin/cubit/router"
	"github.com/dantin/cubit/router/host"
	memorystorage "github.com/dantin/cubit/storage/memory"
	"github.com/dantin/cubit/stream"
	"github.com/dantin/cubit/xmpp"
	"github.com/dantin/cubit/xmpp/jid"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestModule_Ultrasound_MatchesIQ(t *testing.T) {
	r := setupTest()

	srvJID, _ := jid.New("", "example.org", "", true)
	j, _ := jid.New("user", "example.org", "desktop", true)

	stm := stream.NewMockC2S(uuid.New().String(), j)
	stm.SetPresence(xmpp.NewPresence(j, j, xmpp.AvailableType))

	r.Bind(context.Background(), stm)

	cfg := Config{}
	x := New(&cfg, nil, r, nil)
	defer func() { _ = x.Shutdown() }()

	// test MatchesIQ
	iq := xmpp.NewIQType(uuid.New().String(), xmpp.GetType)
	iq.SetFromJID(j)
	iq.SetToJID(srvJID)

	query := xmpp.NewElementNamespace("query", ultrasoundNamespace)
	iq.AppendElement(query)
	require.False(t, x.MatchesIQ(iq))

	iq.ClearElements()
	query = xmpp.NewElementNamespace("profile", ultrasoundNamespace)
	iq.AppendElement(query)
	require.True(t, x.MatchesIQ(iq))

	iq.ClearElements()
	query = xmpp.NewElementNamespace("rooms", ultrasoundNamespace)
	iq.AppendElement(query)
	require.True(t, x.MatchesIQ(iq))
}

func setupTest() router.Router {
	hosts, _ := host.New([]host.Config{{Name: "example.org", Certificate: tls.Certificate{}}})
	r, _ := router.New(
		hosts,
		c2srouter.New(memorystorage.NewUser(), memorystorage.NewBlockList()),
		nil,
	)
	return r
}
