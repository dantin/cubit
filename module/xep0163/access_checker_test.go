package xep0163

import (
	"context"
	"testing"

	pubsubmodel "github.com/dantin/cubit/model/pubsub"
	rostermodel "github.com/dantin/cubit/model/roster"
	memorystorage "github.com/dantin/cubit/storage/memory"
	"github.com/stretchr/testify/require"
)

func TestModule_XEP0163_AccessChecker_Open(t *testing.T) {
	ac := &accessChecker{
		host:        "user@example.org",
		nodeID:      "current_status",
		accessModel: pubsubmodel.Open,
		rosterRep:   memorystorage.NewRoster(),
	}

	err := ac.checkAccess(context.Background(), "alice@example.org")
	require.Nil(t, err)
}

func TestModule_XEP0163_AccessChecker_Outcast(t *testing.T) {
	ac := &accessChecker{
		host:        "user@example.org",
		nodeID:      "current_status",
		accessModel: pubsubmodel.Open,
		affiliation: &pubsubmodel.Affiliation{JID: "alice@example.org", Affiliation: pubsubmodel.Outcast},
		rosterRep:   memorystorage.NewRoster(),
	}

	err := ac.checkAccess(context.Background(), "alice@example.org")
	require.NotNil(t, err)
	require.Equal(t, errOutcastMember, err)
}

func TestModule_XEP0163_AccessChecker_PresenceSubscription(t *testing.T) {
	rosterRep := memorystorage.NewRoster()
	ac := &accessChecker{
		host:        "user@example.org",
		nodeID:      "current_status",
		accessModel: pubsubmodel.Presence,
		rosterRep:   rosterRep,
	}

	err := ac.checkAccess(context.Background(), "alice@example.org")
	require.NotNil(t, err)
	require.Equal(t, errPresenceSubscriptionRequired, err)

	_, _ = rosterRep.UpsertRosterItem(context.Background(), &rostermodel.Item{
		Username:     "user",
		JID:          "alice@example.org",
		Subscription: rostermodel.SubscriptionFrom,
	})

	err = ac.checkAccess(context.Background(), "alice@example.org")
	require.Nil(t, err)
}

func TestModule_XEP0163_AccessChecker_RosterGroup(t *testing.T) {
	rosterRep := memorystorage.NewRoster()
	ac := &accessChecker{
		host:                "user@example.org",
		nodeID:              "current_status",
		rosterAllowedGroups: []string{"Work"},
		accessModel:         pubsubmodel.Roster,
		rosterRep:           rosterRep,
	}

	err := ac.checkAccess(context.Background(), "alice@example.org")
	require.NotNil(t, err)
	require.Equal(t, errNotInRosterGroup, err)

	_, _ = rosterRep.UpsertRosterItem(context.Background(), &rostermodel.Item{
		Username:     "user",
		JID:          "alice@example.org",
		Groups:       []string{"Work"},
		Subscription: rostermodel.SubscriptionFrom,
	})

	err = ac.checkAccess(context.Background(), "alice@example.org")
	require.Nil(t, err)
}

func TestModule_XEP0163_AccessChecker_Member(t *testing.T) {
	ac := &accessChecker{
		host:        "user@example.org",
		nodeID:      "current_status",
		accessModel: pubsubmodel.WhiteList,
		affiliation: &pubsubmodel.Affiliation{JID: "alice@example.org", Affiliation: pubsubmodel.Member},
		rosterRep:   memorystorage.NewRoster(),
	}

	err := ac.checkAccess(context.Background(), "alice2@example.org")
	require.NotNil(t, err)
	require.Equal(t, errNotOnWhiteList, err)

	err = ac.checkAccess(context.Background(), "alice@example.org")
	require.Nil(t, err)
}
