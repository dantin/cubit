package xep0077

import (
	"context"
	"crypto/tls"
	"testing"

	c2srouter "github.com/dantin/cubit/c2s/router"
	"github.com/dantin/cubit/model"
	"github.com/dantin/cubit/router"
	"github.com/dantin/cubit/router/host"
	memorystorage "github.com/dantin/cubit/storage/memory"
	"github.com/dantin/cubit/stream"
	"github.com/dantin/cubit/xmpp"
	"github.com/dantin/cubit/xmpp/jid"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestModule_XEP0077_Matching(t *testing.T) {
	r, s := setupTest("example.org")

	j, _ := jid.New("user", "example.org", "desktop", true)

	x := New(&Config{}, nil, r, s)
	defer func() { _ = x.Shutdown() }()

	// test MatchesIQ
	iq := xmpp.NewIQType(uuid.New().String(), xmpp.SetType)
	iq.SetFromJID(j)

	require.False(t, x.MatchesIQ(iq))
	iq.AppendElement(xmpp.NewElementNamespace("query", registerNamespace))
	require.True(t, x.MatchesIQ(iq))
}

func TestModule_XEP0077_InvalidToJID(t *testing.T) {
	r, s := setupTest("example.org")

	j1, _ := jid.New("alice", "example.org", "desktop", true)
	j2, _ := jid.New("bob", "example.org", "desktop", true)

	stm1 := stream.NewMockC2S(uuid.New().String(), j1)
	r.Bind(context.Background(), stm1)

	x := New(&Config{}, nil, r, s)
	defer func() { _ = x.Shutdown() }()

	iq := xmpp.NewIQType(uuid.New().String(), xmpp.SetType)
	iq.SetFromJID(j1)
	iq.SetToJID(j2.ToBareJID())
	stm1.SetAuthenticated(true)

	x.ProcessIQ(context.Background(), iq)
	elem := stm1.ReceiveElement()
	require.Equal(t, xmpp.ErrForbidden.Error(), elem.Error().Elements().All()[0].Name())

	iq2 := xmpp.NewIQType(uuid.New().String(), xmpp.SetType)
	iq2.SetFromJID(j1)
	iq2.SetToJID(j1.ToBareJID())
}

func TestModule_XEP0077_NotAuthenticatedErrors(t *testing.T) {
	r, s := setupTest("example.org")

	j, _ := jid.New("user", "example.org", "desktop", true)

	stm := stream.NewMockC2S(uuid.New().String(), j)
	r.Bind(context.Background(), stm)

	x := New(&Config{}, nil, r, s)
	defer func() { _ = x.Shutdown() }()

	iq := xmpp.NewIQType(uuid.New().String(), xmpp.ResultType)
	iq.SetFromJID(j)
	iq.SetToJID(j.ToBareJID())

	x.ProcessIQ(context.Background(), iq)
	elem := stm.ReceiveElement()
	require.Equal(t, xmpp.ErrBadRequest.Error(), elem.Error().Elements().All()[0].Name())

	iq.SetType(xmpp.GetType)
	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ErrNotAllowed.Error(), elem.Error().Elements().All()[0].Name())

	// allow registration...
	x = New(&Config{AllowRegistration: true}, nil, r, s)
	defer func() { _ = x.Shutdown() }()

	q := xmpp.NewElementNamespace("query", registerNamespace)
	q.AppendElement(xmpp.NewElementName("q2"))
	iq.AppendElement(q)

	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ErrBadRequest.Error(), elem.Error().Elements().All()[0].Name())

	q.ClearElements()
	iq.SetType(xmpp.SetType)
	stm.SetValue(xep077RegisteredCtxKey, true)

	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ErrNotAcceptable.Error(), elem.Error().Elements().All()[0].Name())
}

func TestModule_XEP0077_AuthenticatedErrors(t *testing.T) {
	r, s := setupTest("example.org")

	srvJid, _ := jid.New("", "example.org", "", true)
	j, _ := jid.New("user", "example.org", "desktop", true)

	stm := stream.NewMockC2S(uuid.New().String(), j)
	r.Bind(context.Background(), stm)

	stm.SetAuthenticated(true)

	x := New(&Config{}, nil, r, s)
	defer func() { _ = x.Shutdown() }()

	iq := xmpp.NewIQType(uuid.New().String(), xmpp.ResultType)
	iq.SetFromJID(j)
	iq.SetToJID(j.ToBareJID())
	iq.SetToJID(srvJid)

	x.ProcessIQ(context.Background(), iq)
	elem := stm.ReceiveElement()
	require.Equal(t, xmpp.ErrBadRequest.Error(), elem.Error().Elements().All()[0].Name())

	iq.SetType(xmpp.SetType)
	iq.AppendElement(xmpp.NewElementNamespace("query", registerNamespace))
	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ErrBadRequest.Error(), elem.Error().Elements().All()[0].Name())
}

func TestModule_XEP0077_RegisterUser(t *testing.T) {
	r, s := setupTest("example.org")

	srvJid, _ := jid.New("", "example.org", "", true)
	j, _ := jid.New("user", "example.org", "desktop", true)

	stm := stream.NewMockC2S(uuid.New().String(), j)
	r.Bind(context.Background(), stm)

	x := New(&Config{AllowRegistration: true}, nil, r, s)
	defer func() { _ = x.Shutdown() }()

	iq := xmpp.NewIQType(uuid.New().String(), xmpp.GetType)
	iq.SetFromJID(j)
	iq.SetToJID(srvJid)

	q := xmpp.NewElementNamespace("query", registerNamespace)
	iq.AppendElement(q)

	x.ProcessIQ(context.Background(), iq)
	q2 := stm.ReceiveElement().Elements().ChildNamespace("query", registerNamespace)
	require.NotNil(t, q2.Elements().Child("username"))
	require.NotNil(t, q2.Elements().Child("password"))

	username := xmpp.NewElementName("username")
	password := xmpp.NewElementName("password")
	q.AppendElement(username)
	q.AppendElement(password)

	// empty fields
	iq.SetType(xmpp.SetType)
	x.ProcessIQ(context.Background(), iq)
	elem := stm.ReceiveElement()
	require.Equal(t, xmpp.ErrBadRequest.Error(), elem.Error().Elements().All()[0].Name())

	// already existing user...
	_ = s.UpsertUser(context.Background(), &model.User{Username: "existed", Password: "password"})
	username.SetText("existed")
	password.SetText("password")
	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ErrConflict.Error(), elem.Error().Elements().All()[0].Name())

	// storage error
	memorystorage.EnableMockedError()
	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ErrInternalServerError.Error(), elem.Error().Elements().All()[0].Name())
	memorystorage.DisableMockedError()

	username.SetText("user")
	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ResultType, elem.Type())

	usr, _ := s.FetchUser(context.Background(), "user")
	require.NotNil(t, usr)
}

func TestModule_XEP0077_CancelRegistration(t *testing.T) {
	r, s := setupTest("example.org")

	srvJid, _ := jid.New("", "example.org", "", true)
	j, _ := jid.New("user", "example.org", "desktop", true)

	stm := stream.NewMockC2S("s-123", j)
	r.Bind(context.Background(), stm)

	stm.SetAuthenticated(true)

	x := New(&Config{}, nil, r, s)
	defer func() { _ = x.Shutdown() }()

	_ = s.UpsertUser(context.Background(), &model.User{Username: "user", Password: "password"})

	iq := xmpp.NewIQType(uuid.New().String(), xmpp.SetType)
	iq.SetFromJID(j)
	iq.SetToJID(srvJid)

	q := xmpp.NewElementNamespace("query", registerNamespace)
	q.AppendElement(xmpp.NewElementName("remove"))

	iq.AppendElement(q)
	x.ProcessIQ(context.Background(), iq)
	elem := stm.ReceiveElement()
	require.Equal(t, xmpp.ErrNotAllowed.Error(), elem.Error().Elements().All()[0].Name())

	x = New(&Config{AllowCancel: true}, nil, r, s)
	defer func() { _ = x.Shutdown() }()

	q.AppendElement(xmpp.NewElementName("remove2"))
	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ErrBadRequest.Error(), elem.Error().Elements().All()[0].Name())
	q.ClearElements()
	q.AppendElement(xmpp.NewElementName("remove"))

	// storage error
	memorystorage.EnableMockedError()
	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ErrInternalServerError.Error(), elem.Error().Elements().All()[0].Name())
	memorystorage.DisableMockedError()

	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ResultType, elem.Type())

	usr, _ := s.FetchUser(context.Background(), "user")
	require.Nil(t, usr)
}

func TestModule_XEP0077_ChangePassword(t *testing.T) {
	r, s := setupTest("example.org")

	srvJid, _ := jid.New("", "example.org", "", true)
	j, _ := jid.New("user", "example.org", "desktop", true)

	stm := stream.NewMockC2S(uuid.New().String(), j)
	r.Bind(context.Background(), stm)

	stm.SetAuthenticated(true)

	x := New(&Config{}, nil, r, s)
	defer func() { _ = x.Shutdown() }()

	_ = s.UpsertUser(context.Background(), &model.User{Username: "user", Password: "password"})

	iq := xmpp.NewIQType(uuid.New().String(), xmpp.SetType)
	iq.SetFromJID(j)
	iq.SetToJID(srvJid)

	q := xmpp.NewElementNamespace("query", registerNamespace)
	username := xmpp.NewElementName("username")
	username.SetText("alice")
	password := xmpp.NewElementName("password")
	password.SetText("passwd")
	q.AppendElement(username)
	q.AppendElement(password)
	iq.AppendElement(q)

	x.ProcessIQ(context.Background(), iq)
	elem := stm.ReceiveElement()
	require.Equal(t, xmpp.ErrNotAllowed.Error(), elem.Error().Elements().All()[0].Name())

	x = New(&Config{AllowChange: true}, nil, r, s)
	defer func() { _ = x.Shutdown() }()

	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ErrNotAllowed.Error(), elem.Error().Elements().All()[0].Name())

	username.SetText("user")
	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ErrNotAuthorized.Error(), elem.Error().Elements().All()[0].Name())

	// secure channel...
	stm.SetSecured(true)

	// storage error
	memorystorage.EnableMockedError()
	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ErrInternalServerError.Error(), elem.Error().Elements().All()[0].Name())
	memorystorage.DisableMockedError()

	x.ProcessIQ(context.Background(), iq)
	elem = stm.ReceiveElement()
	require.Equal(t, xmpp.ResultType, elem.Type())

	usr, _ := s.FetchUser(context.Background(), "user")
	require.NotNil(t, usr)
	require.Equal(t, "passwd", usr.Password)
}

func setupTest(domain string) (router.Router, *memorystorage.User) {
	hosts, _ := host.New([]host.Config{{Name: domain, Certificate: tls.Certificate{}}})
	userRep := memorystorage.NewUser()
	r, _ := router.New(
		hosts,
		c2srouter.New(userRep, memorystorage.NewBlockList()),
		nil,
	)
	return r, userRep
}
