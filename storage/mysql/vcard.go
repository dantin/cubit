package mysql

import (
	"context"
	"database/sql"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/dantin/cubit/xmpp"
)

type mySQLVCard struct {
	*mySQLStorage
}

func newVCard(db *sql.DB) *mySQLVCard {
	return &mySQLVCard{
		mySQLStorage: newStorage(db),
	}
}

func (s *mySQLVCard) UpsertVCard(ctx context.Context, vCard xmpp.XElement, username string) error {
	rawXML := vCard.String()
	q := sq.Insert("vcards").
		Columns("username", "vcard", "updated_at", "created_at").
		Values(username, rawXML, nowExpr, nowExpr).
		Suffix("ON DUPLICATE KEY UPDATE vcard = ?, updated_at = NOW()", rawXML)
	_, err := q.RunWith(s.db).ExecContext(ctx)
	return err
}
func (s *mySQLVCard) FetchVCard(ctx context.Context, username string) (xmpp.XElement, error) {
	var vCard string

	q := sq.Select("vcard").
		From("vcards").
		Where(sq.Eq{"username": username})

	err := q.RunWith(s.db).QueryRowContext(ctx).Scan(&vCard)

	switch err {
	case nil:
		parser := xmpp.NewParser(strings.NewReader(vCard), xmpp.DefaultMode, 0)
		return parser.ParseElement()
	case sql.ErrNoRows:
		return nil, nil
	default:
		return nil, err
	}
}
