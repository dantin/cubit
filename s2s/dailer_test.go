package s2s

import (
	"context"
	"errors"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDialer_Dail(t *testing.T) {
	d := newDialer()

	// resolver error...
	mockedErr := errors.New("dialer mocked error")

	// dialer error...
	d.srvResolve = func(service, proto, name string) (cname string, addrs []*net.SRV, err error) {
		return "", []*net.SRV{{Target: "example.org", Port: 5269}}, nil

	}
	d.dialContext = func(_ context.Context, _, _ string) (net.Conn, error) {
		return nil, mockedErr

	}
	out, err := d.Dial(context.Background(), "example.org")
	require.Nil(t, out)
	require.Equal(t, mockedErr, err)
	// success
	d.dialContext = func(_ context.Context, _, _ string) (net.Conn, error) {
		return newFakeSocketConn(), nil

	}
	out, err = d.Dial(context.Background(), "example.org")
	require.NotNil(t, out)
	require.Nil(t, err)
}
