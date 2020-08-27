package memorystorage

import (
	"context"

	"github.com/dantin/cubit/xmpp"
)

// VCard represents an in-memory vCard storage.
type VCard struct {
	*memoryStorage
}

// NewVCard returns an instance of VCardin-memory storage.
func NewVCard() *VCard {
	return &VCard{memoryStorage: newStorage()}
}

// UpsertVCard inserts a new vCard element into storage, or updates it in case it's been previously inserted.
func (m *VCard) UpsertVCard(_ context.Context, vCard xmpp.XElement, username string) error {
	return m.saveEntity(vCardsKey(username), vCard)
}

// FetchVCard retrieves from storage a vCard element associated to a given user.
func (m *VCard) FetchVCard(_ context.Context, username string) (xmpp.XElement, error) {
	var vCard xmpp.Element
	ok, err := m.getEntity(vCardsKey(username), &vCard)
	switch err {
	case nil:
		if ok {
			return &vCard, nil
		}
		return nil, nil
	default:
		return nil, err
	}
}

func vCardsKey(username string) string {
	return "vCards:" + username
}
