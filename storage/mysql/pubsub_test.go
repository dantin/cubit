package mysql

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	pubsubmodel "github.com/dantin/cubit/model/pubsub"
	"github.com/dantin/cubit/xmpp"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func newPubSubMock() (*mySQLPubSub, sqlmock.Sqlmock) {
	s, sqlMock := newStorageMock()
	return &mySQLPubSub{
		mySQLStorage: s,
	}, sqlMock
}

func TestMySQLStorage_FetchPubSubHosts(t *testing.T) {
	var columns = []string{"host"}

	s, mock := newPubSubMock()
	mock.ExpectQuery("SELECT DISTINCT\\(host\\) FROM pubsub_nodes").
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow("user1@example.org").
			AddRow("user2@example.org"))

	hosts, err := s.FetchHosts(context.Background())

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
	require.NotNil(t, hosts)
	require.Equal(t, "user1@example.org", hosts[0])
	require.Equal(t, "user2@example.org", hosts[1])

	s, mock = newPubSubMock()
	mock.ExpectQuery("SELECT DISTINCT\\(host\\) FROM pubsub_nodes").
		WillReturnError(errMySQLStorage)

	hosts, err = s.FetchHosts(context.Background())

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, hosts)
	require.Equal(t, errMySQLStorage, err)
}

func TestMySQLStorage_UpsertPubSubHosts(t *testing.T) {
	s, mock := newPubSubMock()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO pubsub_nodes (.+) ON DUPLICATE KEY UPDATE (.+)").
		WithArgs("host", "name").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT id FROM pubsub_nodes WHERE (.+)").
		WithArgs("host", "name").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))
	mock.ExpectExec("DELETE FROM pubsub_node_options WHERE (.+)").
		WithArgs("1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	opts := pubsubmodel.Options{}

	optMap, _ := opts.Map()
	for i := 0; i < len(optMap); i++ {
		mock.ExpectExec("INSERT INTO pubsub_node_options (.+)").
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 1))
	}
	mock.ExpectCommit()

	node := pubsubmodel.Node{Host: "host", Name: "name", Options: opts}
	err := s.UpsertNode(context.Background(), &node)

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
}

func TestMySQLStorage_FetchPubSubNodes(t *testing.T) {
	var (
		nodeColumns   = []string{"name"}
		optionColumns = []string{"name", "value"}
	)

	optionRows := sqlmock.NewRows(optionColumns).
		AddRow("pubsub#access_model", "presence").
		AddRow("pubsub#publish_model", "publishers").
		AddRow("pubsub#send_last_published_item", "on_sub_and_presence")

	s, mock := newPubSubMock()
	mock.ExpectQuery("SELECT name FROM pubsub_nodes WHERE host = (.+)").
		WithArgs("demo@example.org").
		WillReturnRows(sqlmock.NewRows(nodeColumns).
			AddRow("name1").
			AddRow("name2"))
	mock.ExpectQuery("SELECT name, value FROM pubsub_node_options WHERE (.+)").
		WithArgs("demo@example.org", "name1").
		WillReturnRows(optionRows)
	mock.ExpectQuery("SELECT name, value FROM pubsub_node_options WHERE (.+)").
		WithArgs("demo@example.org", "name2").
		WillReturnRows(optionRows)

	nodes, err := s.FetchNodes(context.Background(), "demo@example.org")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
	require.NotNil(t, nodes)
	require.Len(t, nodes, 2)
	require.Equal(t, "name1", nodes[0].Name)
	require.Equal(t, "name2", nodes[1].Name)
}

func TestMySQLStorage_FetchPubSubSubscribedNodes(t *testing.T) {
	var (
		nodeColumns   = []string{"host", "name"}
		optionColumns = []string{"name", "value"}
	)

	optionRows := sqlmock.NewRows(optionColumns).
		AddRow("pubsub#access_model", "presence").
		AddRow("pubsub#publish_model", "publishers").
		AddRow("pubsub#send_last_published_item", "on_sub_and_presence")

	s, mock := newPubSubMock()
	mock.ExpectQuery("SELECT host, name FROM pubsub_nodes WHERE id IN (.+)").
		WithArgs("demo@example.org", pubsubmodel.Subscribed).
		WillReturnRows(sqlmock.NewRows(nodeColumns).
			AddRow("demo@example.org", "name1").
			AddRow("demo@example.org", "name2"))
	mock.ExpectQuery("SELECT name, value FROM pubsub_node_options WHERE (.+)").
		WithArgs("demo@example.org", "name1").
		WillReturnRows(optionRows)
	mock.ExpectQuery("SELECT name, value FROM pubsub_node_options WHERE (.+)").
		WithArgs("demo@example.org", "name2").
		WillReturnRows(optionRows)

	nodes, err := s.FetchSubscribedNodes(context.Background(), "demo@example.org")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
	require.NotNil(t, nodes)
	require.Len(t, nodes, 2)
	require.Equal(t, "name1", nodes[0].Name)
	require.Equal(t, "name2", nodes[1].Name)
}

func TestMySQLStorage_FetchPubSubNode(t *testing.T) {
	var cols = []string{"name", "value"}

	s, mock := newPubSubMock()
	mock.ExpectQuery("SELECT name, value FROM pubsub_node_options WHERE (.+)").
		WithArgs("demo@example.org", "name").
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow("pubsub#access_model", "presence").
			AddRow("pubsub#publish_model", "publishers").
			AddRow("pubsub#send_last_published_item", "on_sub_and_presence"))

	node, err := s.FetchNode(context.Background(), "demo@example.org", "name")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
	require.NotNil(t, node)
	require.Equal(t, node.Options.AccessModel, pubsubmodel.Presence)
	require.Equal(t, node.Options.SendLastPublishedItem, pubsubmodel.OnSubAndPresence)

	// error case
	s, mock = newPubSubMock()
	mock.ExpectQuery("SELECT name, value FROM pubsub_node_options WHERE (.+)").
		WithArgs("demo@example.org", "name").
		WillReturnError(errMySQLStorage)

	node, err = s.FetchNode(context.Background(), "demo@example.org", "name")
	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)
}

func TestMySQLStorage_DeletePubSubNode(t *testing.T) {
	s, mock := newPubSubMock()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM pubsub_nodes WHERE (.+)").
		WithArgs("demo@example.org", "status").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))
	mock.ExpectExec("DELETE FROM pubsub_nodes WHERE (.*)").
		WithArgs("1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM pubsub_node_options WHERE (.*)").
		WithArgs("1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM pubsub_items WHERE (.*)").
		WithArgs("1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM pubsub_affiliations WHERE (.*)").
		WithArgs("1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM pubsub_subscriptions WHERE (.*)").
		WithArgs("1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := s.DeleteNode(context.Background(), "demo@example.org", "status")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
}

func TestMySQLStorage_UpsertPubSubNodeItem(t *testing.T) {
	payload := xmpp.NewIQType(uuid.New().String(), xmpp.GetType)

	s, mock := newPubSubMock()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM pubsub_nodes WHERE (.+)").
		WithArgs("demo@example.org", "status").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))
	mock.ExpectExec("INSERT INTO pubsub_items (.+) ON DUPLICATE KEY UPDATE payload = (.+), publisher = (.+), updated_at = NOW()").
		WithArgs("1", "123", payload.String(), "demo@example.org", payload.String(), "demo@example.org").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT item_id FROM pubsub_items WHERE node_id = \\? ORDER BY created_at DESC LIMIT 1").
		WithArgs("1").
		WillReturnRows(sqlmock.NewRows([]string{"item_id"}).AddRow("1").AddRow("2"))
	mock.ExpectExec("DELETE FROM pubsub_items WHERE \\(node_id = \\? AND item_id NOT IN \\(.+\\)\\)").
		WithArgs("1", "1", "2").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := s.UpsertNodeItem(context.Background(), &pubsubmodel.Item{
		ID:        "123",
		Publisher: "demo@example.org",
		Payload:   payload,
	}, "demo@example.org", "status", 1)

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
}

func TestMySQLStorage_FetchPubSubNodeItems(t *testing.T) {
	var cols = []string{"item_id", "publisher", "payload"}

	s, mock := newPubSubMock()
	mock.ExpectQuery("SELECT item_id, publisher, payload FROM pubsub_items WHERE node_id = (.+)").
		WithArgs("demo@example.org", "status").
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow("1", "demo@example.org", "<message/>").
			AddRow("2", "tester@example.org", "<iq type='get'/>"))

	items, err := s.FetchNodeItems(context.Background(), "demo@example.org", "status")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
	require.Len(t, items, 2)
	require.Equal(t, "1", items[0].ID)
	require.Equal(t, "2", items[1].ID)

	// error case
	s, mock = newPubSubMock()
	mock.ExpectQuery("SELECT item_id, publisher, payload FROM pubsub_items WHERE node_id = (.+)").
		WithArgs("demo@example.org", "status").
		WillReturnError(errMySQLStorage)

	_, err = s.FetchNodeItems(context.Background(), "demo@example.org", "status")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)
}

func TestMySQLStorage_FetchPubSubNodeItemsWithID(t *testing.T) {
	var (
		cols        = []string{"item_id", "publisher", "payload"}
		identifiers = []string{"1", "2"}
	)

	s, mock := newPubSubMock()
	mock.ExpectQuery("SELECT item_id, publisher, payload FROM pubsub_items WHERE (.+ IN (.+)) ORDER BY created_at").
		WithArgs("demo@example.org", "status", "1", "2").
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow("1", "demo@example.org", "<message/>").
			AddRow("2", "tester@example.org", "<iq type='get'/>"))

	items, err := s.FetchNodeItemsWithIDs(context.Background(), "demo@example.org", "status", identifiers)

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
	require.Len(t, items, 2)
	require.Equal(t, "1", items[0].ID)
	require.Equal(t, "2", items[1].ID)

	// error case
	s, mock = newPubSubMock()
	mock.ExpectQuery("SELECT item_id, publisher, payload FROM pubsub_items WHERE (.+ IN (.+)) ORDER BY created_at").
		WithArgs("demo@example.org", "status", "1", "2").
		WillReturnError(errMySQLStorage)

	items, err = s.FetchNodeItemsWithIDs(context.Background(), "demo@example.org", "status", identifiers)

	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)
}

func TestMySQLStorage_UpsertPubSubNodeAffiliation(t *testing.T) {
	s, mock := newPubSubMock()

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM pubsub_nodes WHERE (.+)").
		WithArgs("demo@example.org", "status").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))
	mock.ExpectExec("INSERT INTO pubsub_affiliations (.+) VALUES (.+) ON DUPLICATE KEY UPDATE affiliation = (.+), updated_at = (.+)").
		WithArgs("1", "demo@example.org", "owner", "owner").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := s.UpsertNodeAffiliation(context.Background(), &pubsubmodel.Affiliation{
		JID:         "demo@example.org",
		Affiliation: "owner",
	}, "demo@example.org", "status")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
}

func TestMySQLStorage_FetchPubSubNodeAffiliation(t *testing.T) {
	var cols = []string{"jid", "affiliation"}

	s, mock := newPubSubMock()
	mock.ExpectQuery("SELECT jid, affiliation FROM pubsub_affiliations WHERE (.+)").
		WithArgs("demo@example.org", "status").
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow("demo@example.org", "owner").
			AddRow("pub@example.org", "publisher"))

	affiliations, err := s.FetchNodeAffiliations(context.Background(), "demo@example.org", "status")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
	require.Len(t, affiliations, 2)

	// error case
	s, mock = newPubSubMock()
	mock.ExpectQuery("SELECT jid, affiliation FROM pubsub_affiliations WHERE (.+)").
		WithArgs("demo@example.org", "status").
		WillReturnError(errMySQLStorage)

	affiliations, err = s.FetchNodeAffiliations(context.Background(), "demo@example.org", "status")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)
}

func TestMySQLStorage_DeletePubSubNodeAffiliation(t *testing.T) {
	s, mock := newPubSubMock()
	mock.ExpectExec("DELETE FROM pubsub_affiliations WHERE (.+)").
		WithArgs("pub@example.org", "demo@example.org", "status").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := s.DeleteNodeAffiliation(context.Background(), "pub@example.org", "demo@example.org", "status")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)

	// error case
	s, mock = newPubSubMock()
	mock.ExpectExec("DELETE FROM pubsub_affiliations WHERE (.+)").
		WithArgs("pub@example.org", "demo@example.org", "status").
		WillReturnError(errMySQLStorage)

	err = s.DeleteNodeAffiliation(context.Background(), "pub@example.org", "demo@example.org", "status")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)
}

func TestMySQLStorage_UpsertPubSubNodeSubscription(t *testing.T) {
	s, mock := newPubSubMock()
	mock.ExpectBegin()
	mock.ExpectQuery("SELECT id FROM pubsub_nodes WHERE (.+)").
		WithArgs("demo@example.org", "status").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("1"))
	mock.ExpectExec("INSERT INTO pubsub_subscriptions (.+) VALUES (.+) ON DUPLICATE KEY UPDATE (.+)").
		WithArgs("1", "101", "demo@example.org", "subscribed", "101", "subscribed").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := s.UpsertNodeSubscription(context.Background(), &pubsubmodel.Subscription{
		SubID:        "101",
		JID:          "demo@example.org",
		Subscription: "subscribed",
	}, "demo@example.org", "status")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
}

func TestMySQLStorage_FetchPubSubNodeSubscription(t *testing.T) {
	var cols = []string{"subid", "jid", "subscription"}

	s, mock := newPubSubMock()
	mock.ExpectQuery("SELECT subid, jid, subscription FROM pubsub_subscriptions WHERE (.+)").
		WithArgs("demo@example.org", "status").
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow("1", "demo@example.org", "subscribed").
			AddRow("2", "pub@example.org", "unsubscribed"))

	subscriptions, err := s.FetchNodeSubscriptions(context.Background(), "demo@example.org", "status")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)
	require.Len(t, subscriptions, 2)

	// error case
	s, mock = newPubSubMock()
	mock.ExpectQuery("SELECT subid, jid, subscription FROM pubsub_subscriptions WHERE (.+)").
		WithArgs("demo@example.org", "status").
		WillReturnError(errMySQLStorage)

	subscriptions, err = s.FetchNodeSubscriptions(context.Background(), "demo@example.org", "status")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)
}

func TestMySQLStorage_DeletePubSubNodeSubscription(t *testing.T) {
	s, mock := newPubSubMock()
	mock.ExpectExec("DELETE FROM pubsub_subscriptions WHERE (.+)").
		WithArgs("pub@example.org", "demo@example.org", "status").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := s.DeleteNodeSubscription(context.Background(), "pub@example.org", "demo@example.org", "status")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Nil(t, err)

	// error case
	s, mock = newPubSubMock()
	mock.ExpectExec("DELETE FROM pubsub_subscriptions WHERE (.+)").
		WithArgs("pub@example.org", "demo@example.org", "status").
		WillReturnError(errMySQLStorage)

	err = s.DeleteNodeSubscription(context.Background(), "pub@example.org", "demo@example.org", "status")

	require.Nil(t, mock.ExpectationsWereMet())
	require.Equal(t, errMySQLStorage, err)
}
