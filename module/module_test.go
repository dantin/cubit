package module

import (
	"context"
	"crypto/tls"
	"io/ioutil"
	"testing"
	"time"

	c2srouter "github.com/dantin/cubit/c2s/router"
	"github.com/dantin/cubit/router"
	"github.com/dantin/cubit/router/host"
	"github.com/dantin/cubit/storage"
	"github.com/dantin/cubit/stream"
	"github.com/dantin/cubit/xmpp"
	"github.com/dantin/cubit/xmpp/jid"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

type fakeModule struct {
	shutdownCh chan bool
}

func (m *fakeModule) Shutdown() error {
	if m.shutdownCh != nil {
		close(m.shutdownCh)
	}
	return nil
}

func TestModules_New(t *testing.T) {
	mods := setupModules(t)
	defer func() { _ = mods.Shutdown(context.Background()) }()

	require.Len(t, mods.all, 10)
}

func TestModules_ProcessIQ(t *testing.T) {
	mods := setupModules(t)
	defer func() { _ = mods.Shutdown(context.Background()) }()

	j0, _ := jid.NewWithString("user@example.org/desktop", true)
	j1, _ := jid.NewWithString("user@example.org/surface", true)

	stm := stream.NewMockC2S(uuid.New().String(), j0)
	stm.SetPresence(xmpp.NewPresence(j0.ToBareJID(), j0, xmpp.AvailableType))

	mods.router.Bind(context.Background(), stm)

	iqID := uuid.New().String()
	iq := xmpp.NewIQType(iqID, xmpp.GetType)
	iq.SetFromJID(j0)
	iq.SetToJID(j1)
	mods.ProcessIQ(context.Background(), iq)

	elem := stm.ReceiveElement()
	require.NotNil(t, elem)
	require.Equal(t, iqID, elem.ID())
	require.Equal(t, xmpp.IQName, elem.Name())
	require.Equal(t, xmpp.ErrorType, elem.Type())
}

func TestModules_Shutdown(t *testing.T) {
	mods := setupModules(t)

	var mod fakeModule
	mod.shutdownCh = make(chan bool)

	mods.all = append(mods.all, &mod)
	_ = mods.Shutdown(context.Background())

	select {
	case <-mod.shutdownCh:
		break
	case <-time.After(time.Millisecond * 250):
		require.Fail(t, "modules shutdown timeout")
	}
}

func setupModules(t *testing.T) *Modules {
	var config Config
	b, err := ioutil.ReadFile("../data/modules_cfg.yml")
	require.Nil(t, err)
	err = yaml.Unmarshal(b, &config)
	require.Nil(t, err)

	hosts, _ := host.New([]host.Config{{Name: "example.org", Certificate: tls.Certificate{}}})

	rep, _ := storage.New(&storage.Config{Type: storage.Memory})
	r, _ := router.New(
		hosts,
		c2srouter.New(rep.User(), rep.BlockList()),
		nil,
	)
	return New(&config, r, rep, "id-123")
}
