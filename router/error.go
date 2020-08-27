package router

import "errors"

var (
	// ErrNotExistingAccount will be returned if destination user does not exist.
	ErrNotExistingAccount = errors.New("route: account does not exist")

	// ErrResourceNotFound will be returned if destination resource does not match any of user's available resource.
	ErrResourceNotFound = errors.New("route: resource not found")

	// ErrNotAuthenticated will be returned if destination user is not authenticated at this time.
	ErrNotAuthenticated = errors.New("route: user not authenticated")

	// ErrBlockedJID will be returned if destination user is not authenticated at this time.
	ErrBlockedJID = errors.New("route: destination JID is blocked")

	// ErrFailedRemoteConnect will be returned if it couldn't establish a connection to the remote server.
	ErrFailedRemoteConnect = errors.New("route: failed remote connection")
)
