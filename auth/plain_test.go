package auth

import (
	"bytes"
	"context"
	"encoding/base64"
	"testing"

	"github.com/dantin/cubit/model"
	memorystorage "github.com/dantin/cubit/storage/memory"
	"github.com/dantin/cubit/xmpp"
	"github.com/stretchr/testify/require"
)

func TestAuthPlain_Authentication(t *testing.T) {
	var err error

	testStm, s := authTestSetup(&model.User{Username: "admin", Password: "password"})

	authr := NewPlain(testStm, s)
	require.Equal(t, authr.Mechanism(), "PLAIN")
	require.False(t, authr.UsesChannelBinding())

	elem := xmpp.NewElementNamespace("auth", "urn:ietf:params:xml:ns:xmpp-sasl")
	elem.SetAttribute("mechanism", "PLAIN")
	_ = authr.ProcessElement(context.Background(), elem)

	buf := new(bytes.Buffer)
	buf.WriteByte(0)
	buf.WriteString("admin")
	buf.WriteByte(0)
	buf.WriteString("password")
	elem.SetText(base64.StdEncoding.EncodeToString(buf.Bytes()))

	// storage error...
	memorystorage.EnableMockedError()
	require.Equal(t, authr.ProcessElement(context.Background(), elem), memorystorage.ErrMocked)
	memorystorage.DisableMockedError()

	// valid credentials...
	err = authr.ProcessElement(context.Background(), elem)
	require.Nil(t, err)
	require.Equal(t, "admin", authr.Username())
	require.True(t, authr.Authenticated())

	// already authenticated...
	err = authr.ProcessElement(context.Background(), elem)
	require.Nil(t, err)

	// malformed request
	authr.Reset()
	elem.SetText("")
	err = authr.ProcessElement(context.Background(), elem)
	require.Equal(t, ErrSASLMalformedRequest, err)

	// invalid payload
	authr.Reset()
	elem.SetText("bad formed base64")
	err = authr.ProcessElement(context.Background(), elem)
	require.Equal(t, ErrSASLIncorrectEncoding, err)

	// invalid payload
	buf.Reset()
	buf.WriteByte(0)
	buf.WriteString("admin")
	buf.WriteByte(0)
	buf.WriteString("password")
	buf.WriteByte(0)
	elem.SetText(base64.StdEncoding.EncodeToString(buf.Bytes()))

	authr.Reset()
	err = authr.ProcessElement(context.Background(), elem)
	require.Equal(t, ErrSASLIncorrectEncoding, err)

	// invalid user
	buf.Reset()
	buf.WriteByte(0)
	buf.WriteString("noname")
	buf.WriteByte(0)
	buf.WriteString("passwd")
	elem.SetText(base64.StdEncoding.EncodeToString(buf.Bytes()))

	authr.Reset()
	err = authr.ProcessElement(context.Background(), elem)
	require.Equal(t, ErrSASLNotAuthorized, err)

	// incorrect password
	buf.Reset()
	buf.WriteByte(0)
	buf.WriteString("admin")
	buf.WriteByte(0)
	buf.WriteString("12345")
	elem.SetText(base64.StdEncoding.EncodeToString(buf.Bytes()))

	authr.Reset()
	err = authr.ProcessElement(context.Background(), elem)
	require.Equal(t, ErrSASLNotAuthorized, err)
}
