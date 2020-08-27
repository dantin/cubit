package auth

import (
	"context"
	"testing"

	"github.com/dantin/cubit/model"
	memorystorage "github.com/dantin/cubit/storage/memory"
	"github.com/dantin/cubit/stream"
	"github.com/dantin/cubit/xmpp/jid"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func authTestSetup(user *model.User) (*stream.MockC2S, *memorystorage.User) {
	s := memorystorage.NewUser()

	_ = s.UpsertUser(context.Background(), user)

	j, _ := jid.New("admin", "localhost", "res", true)

	testStm := stream.NewMockC2S(uuid.New().String(), j)
	testStm.SetJID(j)

	return testStm, s
}

func TestAuthError(t *testing.T) {
	require.Equal(t, "incorrect-encoding", ErrSASLIncorrectEncoding.(*SASLError).Error())
	require.Equal(t, "malformed-request", ErrSASLMalformedRequest.(*SASLError).Error())
	require.Equal(t, "not-authorized", ErrSASLNotAuthorized.(*SASLError).Error())
	require.Equal(t, "temporary-auth-failure", ErrSASLTemporaryAuthFailure.(*SASLError).Error())
	require.Equal(t, "incorrect-encoding", ErrSASLIncorrectEncoding.(*SASLError).Element().Name())
	require.Equal(t, "malformed-request", ErrSASLMalformedRequest.(*SASLError).Element().Name())
	require.Equal(t, "not-authorized", ErrSASLNotAuthorized.(*SASLError).Element().Name())
	require.Equal(t, "temporary-auth-failure", ErrSASLTemporaryAuthFailure.(*SASLError).Element().Name())

}
