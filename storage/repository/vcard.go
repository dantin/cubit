package repository

import (
	"context"

	"github.com/dantin/cubit/xmpp"
)

// VCard defines storage operations for vCards
type VCard interface {

	// UpsertVCard inserts a new vCard element into storage, or updates it in case it's been previously inserted.
	UpsertVCard(ctx context.Context, vCard xmpp.XElement, username string) error

	// FetchVCard retrieves from storage a vCard element associated to a given user.
	FetchVCard(ctx context.Context, username string) (xmpp.XElement, error)
}
