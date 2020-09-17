package c2s

import (
	"context"
	"testing"
	"time"

	"github.com/dantin/cubit/component"
	"github.com/dantin/cubit/model"
	"github.com/dantin/cubit/module"
	"github.com/dantin/cubit/router"
	"github.com/dantin/cubit/storage"
	"github.com/dantin/cubit/storage/repository"
	"github.com/dantin/cubit/stream"
	"github.com/dantin/cubit/transport"
	"github.com/dantin/cubit/transport/compress"
	"github.com/dantin/cubit/xmpp"
	"github.com/dantin/cubit/xmpp/jid"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestC2SInStream_ConnectTimeout(t *testing.T) {
	r, userRep, blockListRep := setupTest("localhost")

	stm, _ := tUtilStreamInit(r, userRep, blockListRep)
	time.Sleep(time.Millisecond * 1500)
	require.Equal(t, disconnected, stm.getState())
}

func TestC2SInStream_Disconnect(t *testing.T) {
	r, userRep, blockListRep := setupTest("localhost")

	stm, conn := tUtilStreamInit(r, userRep, blockListRep)
	stm.Disconnect(context.Background(), nil)
	require.True(t, conn.waitClose())

	require.Equal(t, disconnected, stm.getState())
}

func TestC2SInStream_Features(t *testing.T) {
	r, userRep, blockListRep := setupTest("localhost")

	// unsecured features
	stm, conn := tUtilStreamInit(r, userRep, blockListRep)
	tUtilStreamOpen(conn)

	elem := conn.outboundRead()
	require.Equal(t, "stream:stream", elem.Name())

	elem = conn.outboundRead()
	require.Equal(t, "stream:features", elem.Name())
	require.NotNil(t, elem.Elements().ChildNamespace("starttls", tlsNamespace))

	require.Equal(t, connected, stm.getState())

	// secured features
	stm2, conn2 := tUtilStreamInit(r, userRep, blockListRep)
	stm2.setSecured(true)

	tUtilStreamOpen(conn2)

	elem = conn2.outboundRead()
	require.Equal(t, "stream:stream", elem.Name())

	elem = conn2.outboundRead()
	require.Equal(t, "stream:features", elem.Name())
	require.NotNil(t, elem.Elements().ChildNamespace("mechanisms", saslNamespace))
}

func TestC2SInStream_TLS(t *testing.T) {
	r, userRep, blockListRep := setupTest("localhost")

	_ = userRep.UpsertUser(context.Background(), &model.User{Username: "user", Password: "password"})

	stm, conn := tUtilStreamInit(r, userRep, blockListRep)
	tUtilStreamOpen(conn)

	_ = conn.outboundRead() // read stream opening...
	_ = conn.outboundRead() // read stream features...

	_, _ = conn.inboundWrite([]byte(`<starttls xmlns="urn:ietf:params:xml:ns:xmpp-tls"/>`))

	elem := conn.outboundRead()

	require.Equal(t, "proceed", elem.Name())
	require.Equal(t, "urn:ietf:params:xml:ns:xmpp-tls", elem.Namespace())

	require.True(t, stm.IsSecured())
}

func TestC2SInStream_FailAuthenticate(t *testing.T) {
	r, userRep, blockListRep := setupTest("localhost")

	_ = userRep.UpsertUser(context.Background(), &model.User{Username: "user", Password: "password"})

	_, conn := tUtilStreamInit(r, userRep, blockListRep)
	tUtilStreamOpen(conn)
	_ = conn.outboundRead() // read stream opening...
	_ = conn.outboundRead() // read stream features...

	// wrong mechanism
	_, _ = conn.inboundWrite([]byte(`<auth xmlns="urn:ietf:params:xml:ns:xmpp-sasl" mechanism="FOO"/>`))

	elem := conn.outboundRead()
	require.Equal(t, "failure", elem.Name())

	_, _ = conn.inboundWrite([]byte(`<auth xmlns="urn:ietf:params:xml:ns:xmpp-sasl" mechanism="PLAIN">AHVzZXIAYQ==</auth>`))

	elem = conn.outboundRead()
	require.Equal(t, "failure", elem.Name())

	// non-SASL
	_, _ = conn.inboundWrite([]byte(`<iq type='set' id='auth2'><query xmlns='jabber:iq:auth'>
<username>bob</username>
<password>foo</password>
</query>
</iq>`))

	elem = conn.outboundRead()
	require.Equal(t, "iq", elem.Name())
	require.Equal(t, xmpp.ErrorType, elem.Type())
	require.NotNil(t, elem.Elements().Child("error"))
}

func TestC2SInStream_Compression(t *testing.T) {
	r, userRep, blockListRep := setupTest("localhost")

	_ = userRep.UpsertUser(context.Background(), &model.User{Username: "user", Password: "password"})

	stm, conn := tUtilStreamInit(r, userRep, blockListRep)
	tUtilStreamOpen(conn)
	_ = conn.outboundRead() // read stream opening...
	_ = conn.outboundRead() // read stream features...

	tUtilStreamAuthenticate(conn, t)

	tUtilStreamOpen(conn)
	_ = conn.outboundRead() // read stream opening...
	_ = conn.outboundRead() // read stream features...

	// no method...
	_, _ = conn.inboundWrite([]byte(`<compress xmlns="http://jabber.org/protocol/compress"/>`))
	elem := conn.outboundRead()
	require.Equal(t, "failure", elem.Name())
	require.NotNil(t, elem.Elements().Child("setup-failed"))

	// invalid method...
	_, _ = conn.inboundWrite([]byte(`<compress xmlns="http://jabber.org/protocol/compress">
<method>7z</method>
</compress>`))
	elem = conn.outboundRead()
	require.Equal(t, "failure", elem.Name())
	require.NotNil(t, elem.Elements().Child("unsupported-method"))

	// valid method...
	_, _ = conn.inboundWrite([]byte(`<compress xmlns="http://jabber.org/protocol/compress">
<method>zlib</method>
</compress>`))

	elem = conn.outboundRead()
	require.Equal(t, "compressed", elem.Name())
	require.Equal(t, "http://jabber.org/protocol/compress", elem.Namespace())

	time.Sleep(time.Millisecond * 100) // wait until processed...

	require.True(t, stm.isCompressed())
}

func TestC2SInStream_StartSession(t *testing.T) {
	r, userRep, blockListRep := setupTest("localhost")

	_ = userRep.UpsertUser(context.Background(), &model.User{Username: "user", Password: "password"})

	stm, conn := tUtilStreamInit(r, userRep, blockListRep)
	tUtilStreamOpen(conn)
	_ = conn.outboundRead() // read stream opening...
	_ = conn.outboundRead() // read stream features...

	tUtilStreamAuthenticate(conn, t)

	tUtilStreamOpen(conn)
	_ = conn.outboundRead() // read stream opening...
	_ = conn.outboundRead() // read stream features...

	tUtilStreamBind(conn, t)
	tUtilStreamStartSession(conn, t)

	require.Equal(t, bound, stm.getState())
}

func TestStream_SendIQ(t *testing.T) {
	r, userRep, blockListRep := setupTest("localhost")

	_ = userRep.UpsertUser(context.Background(), &model.User{Username: "user", Password: "password"})

	stm, conn := tUtilStreamInit(r, userRep, blockListRep)
	tUtilStreamOpen(conn)
	_ = conn.outboundRead() // read stream opening...
	_ = conn.outboundRead() // read stream features...

	tUtilStreamAuthenticate(conn, t)

	tUtilStreamOpen(conn)
	_ = conn.outboundRead() // read stream opening...
	_ = conn.outboundRead() // read stream features...

	tUtilStreamBind(conn, t)
	tUtilStreamStartSession(conn, t)

	require.Equal(t, bound, stm.getState())

	// request roster...
	iqID := uuid.New().String()
	iq := xmpp.NewIQType(iqID, xmpp.GetType)
	iq.AppendElement(xmpp.NewElementNamespace("query", "jabber:iq:roster"))

	_, _ = conn.inboundWrite([]byte(iq.String()))

	elem := conn.outboundRead()
	require.Equal(t, "iq", elem.Name())
	require.Equal(t, iqID, elem.ID())
	require.NotNil(t, elem.Elements().ChildNamespace("query", "jabber:iq:roster"))

	requested, _ := stm.Value("roster:requested").(bool)
	require.True(t, requested)
}

func TestStream_SendPresence(t *testing.T) {
	r, userRep, blockListRep := setupTest("localhost")

	_ = userRep.UpsertUser(context.Background(), &model.User{Username: "user", Password: "password"})

	stm, conn := tUtilStreamInit(r, userRep, blockListRep)
	tUtilStreamOpen(conn)
	_ = conn.outboundRead() // read stream opening...
	_ = conn.outboundRead() // read stream features...

	tUtilStreamAuthenticate(conn, t)

	tUtilStreamOpen(conn)
	_ = conn.outboundRead() // read stream opening...
	_ = conn.outboundRead() // read stream features...

	tUtilStreamBind(conn, t)
	tUtilStreamStartSession(conn, t)

	require.Equal(t, bound, stm.getState())

	_, _ = conn.inboundWrite([]byte(`
<presence>
<show>away</show>
<status>away!</status>
<priority>5</priority>
<x xmlns="vcard-temp:x:update">
<photo>photo-string</photo>
</x>
</presence>
`))
	time.Sleep(time.Millisecond * 100) // wait until processed...

	p := stm.Presence()
	require.NotNil(t, p)
	require.Equal(t, int8(5), p.Priority())
	x := xmpp.NewElementName("x")
	x.AppendElements(stm.Presence().Elements().All())
	require.NotNil(t, x.Elements().Child("show"))
	require.NotNil(t, x.Elements().Child("status"))
	require.NotNil(t, x.Elements().Child("priority"))
	require.NotNil(t, x.Elements().Child("x"))
}

func TestStream_SendMessage(t *testing.T) {
	r, userRep, blockListRep := setupTest("localhost")

	_ = userRep.UpsertUser(context.Background(), &model.User{Username: "user", Password: "password"})

	stm, conn := tUtilStreamInit(r, userRep, blockListRep)
	tUtilStreamOpen(conn)
	_ = conn.outboundRead() // read stream opening...
	_ = conn.outboundRead() // read stream features...

	tUtilStreamAuthenticate(conn, t)

	tUtilStreamOpen(conn)
	_ = conn.outboundRead() // read stream opening...
	_ = conn.outboundRead() // read stream features...

	tUtilStreamBind(conn, t)

	require.Equal(t, bound, stm.getState())

	// define a second stream...
	jFrom, _ := jid.New("user", "localhost", "desktop", true)
	jTo, _ := jid.New("ortuman", "localhost", "surface", true)

	stm2 := stream.NewMockC2S("abc789", jTo)
	stm2.SetPresence(xmpp.NewPresence(jTo, jTo, xmpp.AvailableType))

	r.Bind(context.Background(), stm2)

	msgID := uuid.New().String()
	msg := xmpp.NewMessageType(msgID, xmpp.ChatType)
	msg.SetFromJID(jFrom)
	msg.SetToJID(jTo)
	body := xmpp.NewElementName("body")
	body.SetText("Hi buddy!")
	msg.AppendElement(body)

	_, _ = conn.inboundWrite([]byte(msg.String()))

	// to full jid...
	elem := stm2.ReceiveElement()
	require.Equal(t, "message", elem.Name())
	require.Equal(t, msgID, elem.ID())

	// to bare jid...
	msg.SetToJID(jTo.ToBareJID())
	_, _ = conn.inboundWrite([]byte(msg.String()))
	elem = stm2.ReceiveElement()
	require.Equal(t, "message", elem.Name())
	require.Equal(t, msgID, elem.ID())
}

func TestStream_SendToBlockedJID(t *testing.T) {
	r, userRep, blockListRep := setupTest("localhost")

	_ = userRep.UpsertUser(context.Background(), &model.User{Username: "user", Password: "password"})

	stm, conn := tUtilStreamInit(r, userRep, blockListRep)
	tUtilStreamOpen(conn)
	_ = conn.outboundRead() // read stream opening...
	_ = conn.outboundRead() // read stream features...

	tUtilStreamAuthenticate(conn, t)

	tUtilStreamOpen(conn)
	_ = conn.outboundRead() // read stream opening...
	_ = conn.outboundRead() // read stream features...

	tUtilStreamBind(conn, t)
	tUtilStreamStartSession(conn, t)

	require.Equal(t, bound, stm.getState())

	_ = blockListRep.InsertBlockListItem(context.Background(), &model.BlockListItem{
		Username: "user",
		JID:      "crag@localhost",
	})

	// send presence to a blocked JID...
	_, _ = conn.inboundWrite([]byte(`<presence to="crag@localhost"/>`))

	elem := conn.outboundRead()
	require.Equal(t, "presence", elem.Name())
	require.Equal(t, xmpp.ErrorType, elem.Type())
	require.NotNil(t, elem.Elements().Child("error"))
}

func tUtilStreamOpen(conn *fakeSocketConn) {
	s := `<?xml version="1.0"?>
		<stream:stream xmlns:stream="http://etherx.jabber.org/streams"
		version="1.0" xmlns="jabber:client" to="localhost" xml:lang="en" xmlns:xml="http://www.w3.org/XML/1998/namespace">`
	_, _ = conn.inboundWrite([]byte(s))
}

func tUtilStreamAuthenticate(conn *fakeSocketConn, t *testing.T) {
	_, _ = conn.inboundWrite([]byte(`<auth xmlns="urn:ietf:params:xml:ns:xmpp-sasl" mechanism="PLAIN">AHVzZXIAcGFzc3dvcmQ=</auth>`))

	elem := conn.outboundRead()
	require.Equal(t, "success", elem.Name())
}

func tUtilStreamBind(conn *fakeSocketConn, t *testing.T) {
	// bind a resource
	_, _ = conn.inboundWrite([]byte(`<iq type="set" id="bind_1">
<bind xmlns="urn:ietf:params:xml:ns:xmpp-bind">
<resource>desktop</resource>
</bind>
</iq>`))

	elem := conn.outboundRead()
	require.Equal(t, "iq", elem.Name())
	require.NotNil(t, elem.Elements().Child("bind"))
}

func tUtilStreamStartSession(conn *fakeSocketConn, t *testing.T) {
	// open session
	_, _ = conn.inboundWrite([]byte(`<iq type="set" id="seq-0001">
<session xmlns="urn:ietf:params:xml:ns:xmpp-session"/>
</iq>`))

	elem := conn.outboundRead()
	require.Equal(t, "iq", elem.Name())
	require.NotNil(t, xmpp.ResultType, elem.Type())

	time.Sleep(time.Millisecond * 100) // wait until stream internal state changes
}

func tUtilStreamInit(r router.Router, userRep repository.User, blockListRep repository.BlockList) (*inStream, *fakeSocketConn) {
	conn := newFakeSocketConn()
	tr := transport.NewSocketTransport(conn)
	stm := newStream(
		"abc123",
		tUtilInStreamDefaultConfig(),
		tr,
		tUtilInitModules(r),
		&component.Components{},
		r,
		userRep,
		blockListRep)
	return stm.(*inStream), conn
}

func tUtilInStreamDefaultConfig() *streamConfig {
	return &streamConfig{
		connectTimeout:   time.Second,
		keepAlive:        time.Second,
		maxStanzaSize:    8192,
		resourceConflict: Reject,
		compression:      CompressConfig{Level: compress.DefaultCompression},
		sasl:             []string{"plain", "digest_md5", "scram_sha_1", "scram_sha_256", "scram_sha_512"},
	}
}

func tUtilInitModules(r router.Router) *module.Modules {
	modules := map[string]struct{}{}
	modules["roster"] = struct{}{}
	modules["blocking_command"] = struct{}{}

	repContainer, _ := storage.New(&storage.Config{Type: storage.Memory})
	return module.New(&module.Config{Enabled: modules}, r, repContainer, "alloc-123")
}
