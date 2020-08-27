package memorystorage

import (
	"context"
	"testing"

	rostermodel "github.com/dantin/cubit/model/roster"
	"github.com/dantin/cubit/xmpp"
	"github.com/dantin/cubit/xmpp/jid"
	"github.com/stretchr/testify/require"
)

func TestMemoryStorage_InsertRosterItem(t *testing.T) {
	g := []string{"general", "friends"}
	ri := rostermodel.Item{
		Username:     "user",
		JID:          "contact",
		Name:         "a name",
		Subscription: "both",
		Ask:          false,
		Ver:          1,
		Groups:       g,
	}

	s := NewRoster()
	EnableMockedError()
	_, err := s.UpsertRosterItem(context.Background(), &ri)
	require.Equal(t, ErrMocked, err)
	DisableMockedError()

	_, err = s.UpsertRosterItem(context.Background(), &ri)
	require.Nil(t, err)

	ri.Subscription = "to"
	_, err = s.UpsertRosterItem(context.Background(), &ri)
	require.Nil(t, err)
}

func TestMemoryStorage_FetchRosterItem(t *testing.T) {
	g := []string{"work", "home"}
	ri := rostermodel.Item{
		Username:     "user",
		JID:          "contact",
		Name:         "a name",
		Subscription: "both",
		Ask:          false,
		Ver:          1,
		Groups:       g,
	}
	s := NewRoster()
	_, _ = s.UpsertRosterItem(context.Background(), &ri)

	EnableMockedError()
	_, err := s.FetchRosterItem(context.Background(), "user", "contact")
	require.Equal(t, ErrMocked, err)
	DisableMockedError()

	ri2, _ := s.FetchRosterItem(context.Background(), "user", "contact2")
	require.Nil(t, ri2)

	rc, _ := s.FetchRosterItem(context.Background(), "user", "contact")
	require.NotNil(t, rc)
	require.Equal(t, "user", rc.Username)
	require.Equal(t, "contact", rc.JID)
}

func TestMemoryStorage_FetchRosterItems(t *testing.T) {
	ri := rostermodel.Item{
		Username:     "user",
		JID:          "alice@example.org",
		Name:         "alice",
		Subscription: "both",
		Ask:          false,
		Ver:          1,
		Groups:       []string{"general", "friends"},
	}
	ri2 := rostermodel.Item{
		Username:     "user",
		JID:          "bob@example.org",
		Name:         "bob",
		Subscription: "both",
		Ask:          false,
		Ver:          2,
		Groups:       []string{"general", "buddies"},
	}
	ri3 := rostermodel.Item{
		Username:     "user",
		JID:          "carl@example.org",
		Name:         "carl",
		Subscription: "both",
		Ask:          false,
		Ver:          2,
		Groups:       []string{"family", "friends"},
	}

	s := NewRoster()
	_, _ = s.UpsertRosterItem(context.Background(), &ri)
	_, _ = s.UpsertRosterItem(context.Background(), &ri2)
	_, _ = s.UpsertRosterItem(context.Background(), &ri3)

	EnableMockedError()
	_, _, err := s.FetchRosterItems(context.Background(), "user")
	require.Equal(t, ErrMocked, err)
	DisableMockedError()

	ris, _, _ := s.FetchRosterItems(context.Background(), "user")
	require.Equal(t, 3, len(ris))

	ris, _, _ = s.FetchRosterItemsInGroups(context.Background(), "user", []string{"friends"})
	require.Equal(t, 2, len(ris))

	ris, _, _ = s.FetchRosterItemsInGroups(context.Background(), "user", []string{"buddies"})
	require.Equal(t, 1, len(ris))

	gr, _ := s.FetchRosterGroups(context.Background(), "user")
	require.Len(t, gr, 4)
	require.Contains(t, gr, "general")
	require.Contains(t, gr, "friends")
	require.Contains(t, gr, "family")
	require.Contains(t, gr, "buddies")
}

func TestMemoryStorage_DeleteRosterItem(t *testing.T) {
	g := []string{"general", "friends"}
	ri := rostermodel.Item{
		Username:     "user",
		JID:          "contact",
		Name:         "a name",
		Subscription: "both",
		Ask:          false,
		Ver:          1,
		Groups:       g,
	}
	s := NewRoster()
	_, _ = s.UpsertRosterItem(context.Background(), &ri)

	gr, _ := s.FetchRosterGroups(context.Background(), "user")
	require.Len(t, gr, 2)
	require.Contains(t, gr, "general")
	require.Contains(t, gr, "friends")

	EnableMockedError()
	_, err := s.DeleteRosterItem(context.Background(), "user", "contact")
	require.Equal(t, ErrMocked, err)
	DisableMockedError()

	_, err = s.DeleteRosterItem(context.Background(), "user", "contact")
	require.Nil(t, err)

	_, err = s.DeleteRosterItem(context.Background(), "user2", "contact")
	require.Nil(t, err) // delete not existing roster item...

	ri2, _ := s.FetchRosterItem(context.Background(), "user", "contact")
	require.Nil(t, ri2)

	gr, _ = s.FetchRosterGroups(context.Background(), "user")
	require.Len(t, gr, 0)
}

func TestMemoryStorage_InsertRosterNotification(t *testing.T) {
	rn := rostermodel.Notification{
		Contact:  "alice",
		JID:      "bob@example.org",
		Presence: &xmpp.Presence{},
	}
	s := NewRoster()

	EnableMockedError()
	require.Equal(t, ErrMocked, s.UpsertRosterNotification(context.Background(), &rn))
	DisableMockedError()

	require.Nil(t, s.UpsertRosterNotification(context.Background(), &rn))
}

func TestMemoryStorage_FetchRosterNotifications(t *testing.T) {
	rn1 := rostermodel.Notification{
		Contact:  "user",
		JID:      "alice@example.org",
		Presence: &xmpp.Presence{},
	}
	rn2 := rostermodel.Notification{
		Contact:  "user",
		JID:      "bob@example.org",
		Presence: &xmpp.Presence{},
	}
	s := NewRoster()
	_ = s.UpsertRosterNotification(context.Background(), &rn1)
	_ = s.UpsertRosterNotification(context.Background(), &rn2)

	from, _ := jid.NewWithString("contact@example.org", true)
	to, _ := jid.NewWithString("user@example.org", true)
	rn2.Presence = xmpp.NewPresence(from, to, xmpp.SubscribeType)
	_ = s.UpsertRosterNotification(context.Background(), &rn2)

	EnableMockedError()
	_, err := s.FetchRosterNotifications(context.Background(), "user")
	require.Equal(t, ErrMocked, err)
	DisableMockedError()

	rns, err := s.FetchRosterNotifications(context.Background(), "user")
	require.Nil(t, err)
	require.Len(t, rns, 2)
	require.Equal(t, "alice@example.org", rns[0].JID)
	require.Equal(t, "bob@example.org", rns[1].JID)
}

func TestMemoryStorage_DeleteRosterNotification(t *testing.T) {
	rn1 := rostermodel.Notification{
		Contact:  "user",
		JID:      "alice@example.org",
		Presence: &xmpp.Presence{},
	}
	s := NewRoster()
	_ = s.UpsertRosterNotification(context.Background(), &rn1)

	EnableMockedError()
	require.Equal(t, ErrMocked, s.DeleteRosterNotification(context.Background(), "user", "alice@example.org"))
	DisableMockedError()

	require.Nil(t, s.DeleteRosterNotification(context.Background(), "user", "alice@example.org"))

	rns, err := s.FetchRosterNotifications(context.Background(), "romeo")
	require.Nil(t, err)
	require.Len(t, rns, 0)

	// delete not existing roster notification...
	require.Nil(t, s.DeleteRosterNotification(context.Background(), "none", "alice@example.org"))
}
