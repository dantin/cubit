package s2s

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOutProvider_GetOut(t *testing.T) {
	hosts := setupTestHosts(testDomain)

	op := NewOutProvider(&Config{}, hosts)

	op.dialer.(*dialer).srvResolve = func(service, proto, name string) (cname string, addrs []*net.SRV, err error) {
		return "", []*net.SRV{{Target: "xmpp.test.org", Port: 5269}}, nil
	}
	op.dialer.(*dialer).dialContext = func(_ context.Context, _, _ string) (net.Conn, error) {
		return newFakeSocketConn(), nil
	}
	out := op.GetOut("example.org", "test.org")

	require.NotNil(t, out)

	op.mu.RLock()
	require.Len(t, op.outConnections, 1)
	op.mu.RUnlock()
}

func TestOutProvider_Shutdown(t *testing.T) {
	hosts := setupTestHosts(testDomain)

	op := NewOutProvider(&Config{}, hosts)

	op.dialer.(*dialer).srvResolve = func(service, proto, name string) (cname string, addrs []*net.SRV, err error) {
		return "", []*net.SRV{{Target: "xmpp.test.org", Port: 5269}}, nil
	}
	op.dialer.(*dialer).dialContext = func(_ context.Context, _, _ string) (net.Conn, error) {
		return newFakeSocketConn(), nil
	}
	out := op.GetOut("example.org", "test.org")
	_ = out.(*outStream).start(context.Background()) // start transport

	require.NotNil(t, out)

	op.mu.RLock()
	require.Len(t, op.outConnections, 1)
	op.mu.RUnlock()

	_ = op.Shutdown(context.Background())
	time.Sleep(time.Millisecond * 100) // wait until unregistered

	op.mu.RLock()
	require.Len(t, op.outConnections, 0)
	op.mu.RUnlock()
}
