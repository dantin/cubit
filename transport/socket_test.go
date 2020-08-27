package transport

import (
	"bytes"
	"crypto/tls"
	"net"
	"testing"
	"time"

	"github.com/dantin/cubit/transport/compress"
	"github.com/dantin/cubit/xmpp"
	"github.com/stretchr/testify/require"
)

type fakeAddr int

var (
	localAddr  = fakeAddr(1)
	remoteAddr = fakeAddr(2)
)

func (a fakeAddr) Network() string { return "net" }
func (a fakeAddr) String() string  { return "str" }

type fakeSocketConn struct {
	r      *bytes.Buffer
	w      *bytes.Buffer
	closed bool
}

func newFakeSocketConn() *fakeSocketConn {
	return &fakeSocketConn{
		r: new(bytes.Buffer),
		w: new(bytes.Buffer),
	}
}

func (c *fakeSocketConn) Read(b []byte) (n int, err error)   { return c.r.Read(b) }
func (c *fakeSocketConn) Write(b []byte) (n int, err error)  { return c.w.Write(b) }
func (c *fakeSocketConn) Close() error                       { c.closed = true; return nil }
func (c *fakeSocketConn) LocalAddr() net.Addr                { return localAddr }
func (c *fakeSocketConn) RemoteAddr() net.Addr               { return remoteAddr }
func (c *fakeSocketConn) SetDeadline(t time.Time) error      { return nil }
func (c *fakeSocketConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *fakeSocketConn) SetWriteDeadline(t time.Time) error { return nil }

func TestSocket(t *testing.T) {
	buff := make([]byte, 4096)
	conn := newFakeSocketConn()
	st := NewSocketTransport(conn)
	st2 := st.(*socketTransport)

	el1 := xmpp.NewElementNamespace("elem", "test:ns")
	el1.ToXML(st, true)
	_ = st.Flush()
	require.Equal(t, 0, bytes.Compare([]byte(el1.String()), conn.w.Bytes()))

	el2 := xmpp.NewElementNamespace("elem2", "test2:ns")
	el2.ToXML(conn.r, true)
	n, err := st.Read(buff)
	require.Nil(t, err)
	require.Equal(t, el2.String(), string(buff[:n]))

	st.EnableCompression(compress.BestCompression)
	require.True(t, st2.compressed)

	st.(*socketTransport).conn = &net.TCPConn{}
	st.StartTLS(&tls.Config{}, false)
	_, ok := st2.conn.(*tls.Conn)
	require.True(t, ok)
	st.(*socketTransport).conn = conn

	require.Nil(t, st2.ChannelBindingBytes(ChannelBindingMechanism(99)))
	require.Nil(t, st2.ChannelBindingBytes(TLSUnique))

	st.Close()
	require.True(t, conn.closed)
}
