package auth

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"hash"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dantin/cubit/model"
	"github.com/dantin/cubit/transport"
	"github.com/dantin/cubit/transport/compress"
	utilstring "github.com/dantin/cubit/util/string"
	"github.com/dantin/cubit/xmpp"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/pbkdf2"
)

func randomBytes(l int) []byte {
	b := make([]byte, l)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)

	}
	return b
}

type fakeTransport struct {
	cbBytes []byte
}

func (ft *fakeTransport) Read(p []byte) (n int, err error)        { return 0, nil }
func (ft *fakeTransport) Write(p []byte) (n int, err error)       { return 0, nil }
func (ft *fakeTransport) Close() error                            { return nil }
func (ft *fakeTransport) Type() transport.Type                    { return transport.Socket }
func (ft *fakeTransport) Flush() error                            { return nil }
func (ft *fakeTransport) SetWriteDeadline(_ time.Time) error      { return nil }
func (ft *fakeTransport) WriteString(s string) (n int, err error) { return 0, nil }
func (ft *fakeTransport) StartTLS(*tls.Config, bool)              { return }
func (ft *fakeTransport) EnableCompression(compress.Level)        { return }
func (ft *fakeTransport) ChannelBindingBytes(transport.ChannelBindingMechanism) []byte {
	return ft.cbBytes

}
func (ft *fakeTransport) PeerCertificates() []*x509.Certificate { return nil }

type scramAuthTestCase struct {
	id          int
	scramType   ScramType
	usesCb      bool
	cbBytes     []byte
	gs2BindFlag string
	authID      string
	n           string
	r           string
	password    string
	expectedErr error
}

type scramAuthResult struct {
	clientFinalMessage string
	v                  string
}

var tt = []scramAuthTestCase{

	// Success cases
	{
		// SCRAM-SHA-1
		id:          1,
		scramType:   ScramSHA1,
		usesCb:      false,
		gs2BindFlag: "n",
		n:           "admin",
		r:           "374dcb87-aeae-4da8-b452-022b0646afde",
		password:    "passwd",
	},
	{
		// SCRAM-SHA-256
		id:          2,
		scramType:   ScramSHA256,
		usesCb:      false,
		gs2BindFlag: "n",
		n:           "admin",
		r:           "374dcb87-aeae-4da8-b452-022b0646afde",
		password:    "passwd",
	},
	{
		// SCRAM-SHA-1-PLUS
		id:          3,
		scramType:   ScramSHA1,
		usesCb:      true,
		cbBytes:     randomBytes(23),
		gs2BindFlag: "p=tls-unique",
		authID:      "a=example.org",
		n:           "admin",
		r:           "374dcb87-aeae-4da8-b452-022b0646afde",
		password:    "passwd",
	},
	{
		// SCRAM-SHA-256-PLUS
		id:          4,
		scramType:   ScramSHA256,
		usesCb:      true,
		cbBytes:     randomBytes(32),
		gs2BindFlag: "p=tls-unique",
		authID:      "a=example.org",
		n:           "admin",
		r:           "374dcb87-aeae-4da8-b452-022b0646afde",
		password:    "passwd",
	},

	// Fail cases
	{
		// invalid user
		id:          5,
		scramType:   ScramSHA1,
		usesCb:      false,
		gs2BindFlag: "n",
		n:           "user",
		r:           "374dcb87-aeae-4da8-b452-022b0646afde",
		password:    "passwd",
		expectedErr: ErrSASLNotAuthorized,
	},
	{
		// invalid password
		id:          6,
		scramType:   ScramSHA1,
		usesCb:      false,
		gs2BindFlag: "n",
		n:           "admin",
		r:           "374dcb87-aeae-4da8-b452-022b0646afde",
		password:    "12345678",
		expectedErr: ErrSASLNotAuthorized,
	},
	{
		// not authorized gs2BindFlag
		id:          7,
		scramType:   ScramSHA1,
		usesCb:      false,
		gs2BindFlag: "y",
		n:           "admin",
		r:           "374dcb87-aeae-4da8-b452-022b0646afde",
		password:    "passwd",
		expectedErr: ErrSASLNotAuthorized,
	},
	{
		// invalid authID
		id:          8,
		scramType:   ScramSHA1,
		usesCb:      false,
		gs2BindFlag: "n",
		authID:      "b=example.org",
		n:           "admin",
		r:           "374dcb87-aeae-4da8-b452-022b0646afde",
		password:    "passwd",
		expectedErr: ErrSASLMalformedRequest,
	},
	{
		// not matching gs2BindFlag
		id:          9,
		scramType:   ScramSHA1,
		usesCb:      false,
		gs2BindFlag: "p=tls-unique",
		authID:      "a=example.org",
		n:           "admin",
		r:           "374dcb87-aeae-4da8-b452-022b0646afde",
		password:    "passwd",
		expectedErr: ErrSASLNotAuthorized,
	},
	{
		// not matching gs2BindFlag
		id:          10,
		scramType:   ScramSHA1,
		usesCb:      false,
		gs2BindFlag: "q=tls-unique",
		authID:      "a=example.org",
		n:           "admin",
		r:           "374dcb87-aeae-4da8-b452-022b0646afde",
		password:    "passwd",
		expectedErr: ErrSASLMalformedRequest,
	},
	{
		// empty username
		id:          11,
		scramType:   ScramSHA1,
		usesCb:      false,
		gs2BindFlag: "n",
		authID:      "a=example.org",
		n:           "",
		r:           "374dcb87-aeae-4da8-b452-022b0646afde",
		password:    "passwd",
		expectedErr: ErrSASLMalformedRequest,
	},
}

func TestAuthScram_Mechanisms(t *testing.T) {
	testTr := &fakeTransport{}
	testStm, s := authTestSetup(&model.User{Username: "admin", Password: "passwd"})

	authr := NewScram(testStm, testTr, ScramSHA1, false, s)
	require.Equal(t, authr.Mechanism(), "SCRAM-SHA-1")
	require.False(t, authr.UsesChannelBinding())

	authr2 := NewScram(testStm, testTr, ScramSHA1, true, s)
	require.Equal(t, authr2.Mechanism(), "SCRAM-SHA-1-PLUS")
	require.True(t, authr2.UsesChannelBinding())

	authr3 := NewScram(testStm, testTr, ScramSHA256, false, s)
	require.Equal(t, authr3.Mechanism(), "SCRAM-SHA-256")
	require.False(t, authr3.UsesChannelBinding())

	authr4 := NewScram(testStm, testTr, ScramSHA256, true, s)
	require.Equal(t, authr4.Mechanism(), "SCRAM-SHA-256-PLUS")
	require.True(t, authr4.UsesChannelBinding())

	authr5 := NewScram(testStm, testTr, ScramType(999), true, s)
	require.Equal(t, authr5.Mechanism(), "")
}

func TestAuthScram_BadPayload(t *testing.T) {
	testTr := &fakeTransport{}
	testStm, s := authTestSetup(&model.User{Username: "admin", Password: "passwd"})

	authr := NewScram(testStm, testTr, ScramSHA1, false, s)

	auth := xmpp.NewElementNamespace("auth", "uri:ietf:params:xml:ns:xmpp-sasl")
	auth.SetAttribute("mechanism", authr.Mechanism())

	// empty auth payload
	require.Equal(t, ErrSASLIncorrectEncoding, authr.ProcessElement(context.Background(), auth))

	// incorrect auth payload encoding
	authr.Reset()
	auth.SetText(".")
	require.Equal(t, ErrSASLIncorrectEncoding, authr.ProcessElement(context.Background(), auth))
}

func TestAuthScram_TestCases(t *testing.T) {
	for _, tc := range tt {
		err := processScramTestCase(t, &tc)
		if err != nil {
			require.Equal(t, tc.expectedErr, err, fmt.Sprintf("TC id: %d", tc.id))
		}
	}
}

func processScramTestCase(t *testing.T, tc *scramAuthTestCase) error {
	tr := &fakeTransport{}
	if tc.usesCb {
		tr.cbBytes = tc.cbBytes

	}
	testStm, s := authTestSetup(&model.User{Username: "admin", Password: "passwd"})

	authr := NewScram(testStm, tr, tc.scramType, tc.usesCb, s)

	auth := xmpp.NewElementNamespace("auth", saslNamespace)
	auth.SetAttribute("mechanism", authr.Mechanism())

	clientInitialMessage := fmt.Sprintf(`n=%s,r=%s`, tc.n, tc.r)
	gs2Header := fmt.Sprintf(`%s,%s,`, tc.gs2BindFlag, tc.authID)
	authPayload := gs2Header + clientInitialMessage
	auth.SetText(base64.StdEncoding.EncodeToString([]byte(authPayload)))

	err := authr.ProcessElement(context.Background(), auth)
	if err != nil {
		return err

	}
	challenge := testStm.ReceiveElement()
	require.NotNil(t, challenge)
	require.Equal(t, "challenge", challenge.Name())

	srvInitialMessage, err := base64.StdEncoding.DecodeString(challenge.Text())
	require.Nil(t, err)
	resp, err := parseScramResponse(challenge.Text())
	require.Nil(t, err)

	srvNonce := resp["r"]
	salt, err := base64.StdEncoding.DecodeString(resp["s"])
	require.Nil(t, err)

	iterations, _ := strconv.Atoi(resp["i"])

	buf := new(bytes.Buffer)
	buf.Write([]byte(gs2Header))
	if tc.usesCb {
		buf.Write(tc.cbBytes)

	}
	cBytes := base64.StdEncoding.EncodeToString(buf.Bytes())

	res := computeScramAuthResult(tc.scramType, clientInitialMessage, string(srvInitialMessage), srvNonce, cBytes, tc.password, salt, iterations)

	response := xmpp.NewElementNamespace("response", saslNamespace)
	response.SetText(base64.StdEncoding.EncodeToString([]byte(res.clientFinalMessage)))

	err = authr.ProcessElement(context.Background(), response)
	if err != nil {
		return err

	}

	success := testStm.ReceiveElement()
	require.Equal(t, "success", success.Name())

	vb64, err := base64.StdEncoding.DecodeString(success.Text())
	require.Nil(t, err)
	require.Equal(t, res.v, string(vb64))

	require.True(t, authr.Authenticated())
	require.Equal(t, tc.n, authr.Username())

	require.Nil(t, authr.ProcessElement(context.Background(), auth)) // test already authenticated...
	return nil
}

func computeScramAuthResult(scramType ScramType, clientInitialMessage, serverInitialMessage, srvNonce, cBytes, password string, salt []byte, iterations int) *scramAuthResult {
	clientFinalMessageBare := fmt.Sprintf("c=%s,r=%s", cBytes, srvNonce)

	saltedPassword := testScramAuthPbkdf2([]byte(password), salt, scramType, iterations)
	clientKey := testScramAuthHmac([]byte("Client Key"), saltedPassword, scramType)
	storedKey := testScramAuthHash(clientKey, scramType)
	authMessage := clientInitialMessage + "," + serverInitialMessage + "," + clientFinalMessageBare
	clientSignature := testScramAuthHmac([]byte(authMessage), storedKey, scramType)

	clientProof := make([]byte, len(clientKey))
	for i := 0; i < len(clientKey); i++ {
		clientProof[i] = clientKey[i] ^ clientSignature[i]

	}
	serverKey := testScramAuthHmac([]byte("Server Key"), saltedPassword, scramType)
	serverSignature := testScramAuthHmac([]byte(authMessage), serverKey, scramType)

	res := &scramAuthResult{}
	res.clientFinalMessage = clientFinalMessageBare + ",p=" + base64.StdEncoding.EncodeToString(clientProof)
	res.v = "v=" + base64.StdEncoding.EncodeToString(serverSignature)
	return res

}

func parseScramResponse(b64 string) (map[string]string, error) {
	s, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, err
	}
	ret := map[string]string{}
	s1 := strings.Split(string(s), ",")
	for _, s0 := range s1 {
		k, v := utilstring.SplitKeyAndValue(s0, '=')
		ret[k] = v
	}
	return ret, nil
}

func testScramAuthPbkdf2(b []byte, salt []byte, scramType ScramType, iterationCount int) []byte {
	switch scramType {
	case ScramSHA1:
		return pbkdf2.Key(b, salt, iterationCount, sha1.Size, sha1.New)
	case ScramSHA256:
		return pbkdf2.Key(b, salt, iterationCount, sha256.Size, sha256.New)

	}
	return nil
}

func testScramAuthHmac(b []byte, key []byte, scramType ScramType) []byte {
	var h func() hash.Hash
	switch scramType {
	case ScramSHA1:
		h = sha1.New
	case ScramSHA256:
		h = sha256.New
	}
	m := hmac.New(h, key)
	m.Write(b)
	return m.Sum(nil)
}

func testScramAuthHash(b []byte, scramType ScramType) []byte {
	var h hash.Hash
	switch scramType {
	case ScramSHA1:
		h = sha1.New()
	case ScramSHA256:
		h = sha256.New()
	}
	h.Write(b)
	return h.Sum(nil)
}
