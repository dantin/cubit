package offline

import (
	"errors"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/dantin/cubit/xmpp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fakeReadCloser struct{}

func (rc *fakeReadCloser) Read(p []byte) (n int, err error) { return 0, nil }
func (rc *fakeReadCloser) Close() error                     { return nil }

type fakeHTTPClient struct {
	do func(req *http.Request) (*http.Response, error)
}

func (c *fakeHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return c.do(req)
}

func TestModule_Offline_HttpGateway_Route(t *testing.T) {
	g := newHTTPGateway("http://127.0.0.1:6666", "secret-key").(*httpGateway)
	fakeClient := &fakeHTTPClient{}
	g.client = fakeClient

	msg := xmpp.NewMessageType(uuid.New().String(), xmpp.ChatType)
	body := xmpp.NewElementName("body")
	body.SetText("This is an offline message!")
	msg.AppendElement(body)

	var reqBody string
	fakeClient.do = func(req *http.Request) (response *http.Response, e error) {
		require.Equal(t, http.MethodPost, req.Method)
		require.Equal(t, "secret-key", req.Header.Get("Authorization"))
		require.Equal(t, "application/xml", req.Header.Get("Content-Type"))

		b, _ := ioutil.ReadAll(req.Body)
		reqBody = string(b)
		return &http.Response{StatusCode: http.StatusOK, Body: &fakeReadCloser{}}, nil
	}

	err := g.Route(msg)
	require.Nil(t, err)
	require.Equal(t, msg.String(), reqBody)

	fakeClient.do = func(req *http.Request) (response *http.Response, e error) {
		return &http.Response{StatusCode: http.StatusInternalServerError, Body: &fakeReadCloser{}}, nil
	}
	require.NotNil(t, g.Route(msg))

	fakeClient.do = func(req *http.Request) (response *http.Response, e error) {
		return nil, errors.New("foo error")
	}
	require.NotNil(t, g.Route(msg))
}
