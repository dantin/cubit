package memorystorage

import (
	"context"
	"testing"

	"github.com/dantin/cubit/model"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage_InsertUser(t *testing.T) {
	u := model.User{Username: "user", Password: "password", Role: "user"}
	s := NewUser()

	EnableMockedError()
	err := s.UpsertUser(context.Background(), &u)
	require.Equal(t, ErrMocked, err)
	DisableMockedError()

	err = s.UpsertUser(context.Background(), &u)
	require.Nil(t, err)
}

func TestMemoryStorage_UserExists(t *testing.T) {
	s := NewUser()

	EnableMockedError()
	_, err := s.UserExists(context.Background(), "user")
	require.Equal(t, ErrMocked, err)
	DisableMockedError()

	ok, err := s.UserExists(context.Background(), "user")
	require.Nil(t, err)
	require.False(t, ok)
}

func TestMemoryStorage_FetchUser(t *testing.T) {
	u := model.User{Username: "user", Password: "password", Role: "user"}
	s := NewUser()

	_ = s.UpsertUser(context.Background(), &u)

	EnableMockedError()
	_, err := s.FetchUser(context.Background(), "user")
	require.Equal(t, ErrMocked, err)
	DisableMockedError()

	usr, _ := s.FetchUser(context.Background(), "none")
	require.Nil(t, usr)

	usr, _ = s.FetchUser(context.Background(), "user")
	require.NotNil(t, usr)
}

func TestMemoryStorage_DeleteUser(t *testing.T) {
	u := model.User{Username: "user", Password: "password", Role: "user"}
	s := NewUser()

	_ = s.UpsertUser(context.Background(), &u)

	EnableMockedError()
	require.Equal(t, ErrMocked, s.DeleteUser(context.Background(), "user"))
	DisableMockedError()

	require.Nil(t, s.DeleteUser(context.Background(), "user"))

	usr, _ := s.FetchUser(context.Background(), "user")
	require.Nil(t, usr)
}
