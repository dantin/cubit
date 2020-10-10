package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/dantin/cubit/log"
	"github.com/dantin/cubit/model"
	roomsmodel "github.com/dantin/cubit/model/rooms"
	"github.com/dantin/cubit/util/pool"
)

type mySQLRoom struct {
	*mySQLStorage
	pool *pool.BufferPool
}

func newRoom(db *sql.DB) *mySQLRoom {
	return &mySQLRoom{
		mySQLStorage: newStorage(db),
		pool:         pool.NewBufferPool(),
	}
}

func (r *mySQLRoom) FetchRoom(ctx context.Context, username string) (*roomsmodel.Room, error) {
	var (
		res  roomsmodel.Room
		name string
		user string
		id   int
	)
	err := sq.Select("id", "name", "username").
		From("rooms").
		Where(sq.Eq{"username": username}).
		Limit(1).
		RunWith(r.db).QueryRowContext(ctx).Scan(&id, &name, &user)
	switch err {
	case nil:
		streams, err := r.scanVideoStreams(ctx, model.Usr.String(), id)
		if err != nil {
			return nil, err
		}
		res.ID = id
		res.Name = name
		res.Username = user
		res.Streams = streams
		return &res, nil
	case sql.ErrNoRows:
		return nil, nil
	default:
		return nil, err
	}
}

func (r *mySQLRoom) FetchRooms(ctx context.Context, page int, pageSize int) ([]roomsmodel.Room, error) {
	log.Debugf("fetch rooms on page %d size %d", page, pageSize)

	q := sq.Select("id", "name", "username").
		From("rooms").
		Where(sq.Eq{"`type`": roomsmodel.Normal.String()}).
		OrderBy("`id` ASC").
		Limit(uint64(pageSize)).
		Offset(uint64(page * pageSize))

	rows, err := q.RunWith(r.db).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	// fetch video streams and build.
	var ids []int
	rooms := make(map[int]*roomsmodel.Room)
	for rows.Next() {
		var (
			id             int
			name, username string
		)
		if err := rows.Scan(&id, &name, &username); err != nil {
			return nil, err
		}
		ids = append(ids, id)
		rooms[id] = &roomsmodel.Room{ID: id, Name: name, Username: username, Streams: nil}
	}

	vs, err := r.scanVideoStreams(ctx, model.Admin.String(), ids...)
	if err != nil {
		return nil, err
	}

	for _, v := range vs {
		room := rooms[v.RoomID]
		room.Streams = append(room.Streams, v)
	}

	var res []roomsmodel.Room
	for _, r := range rooms {
		res = append(res, *r)
	}
	return res, nil
}

// CountRooms  returns current size of rooms.
func (r *mySQLRoom) CountRooms(ctx context.Context) (int, error) {
	q := sq.Select("COUNT(*)").
		From("rooms").
		Where(sq.Eq{"`type`": roomsmodel.Normal.String()})

	var count int
	err := q.RunWith(r.db).QueryRowContext(ctx).Scan(&count)
	switch err {
	case nil:
		return count, nil
	default:
		return 0, err
	}
}

// FetchQCStream retrieves qc room stream entites from storage.
func (r *mySQLRoom) FetchQCStream(ctx context.Context, username string) (*roomsmodel.VideoStream, error) {
	var id int
	err := sq.Select("id").
		From("rooms").
		Where(sq.Eq{"`type`": roomsmodel.QC.String()}).
		RunWith(r.db).QueryRowContext(ctx).Scan(&id)
	switch err {
	case nil:
		res, err := r.scanVideoStreams(ctx, username, id)
		if err != nil {
			return nil, err
		}
		if len(res) != 1 {
			return nil, errors.New("multiple qc found")
		}
		return &res[0], nil
	case sql.ErrNoRows:
		return nil, nil
	default:
		return nil, err
	}
}

func (r *mySQLRoom) scanVideoStreams(ctx context.Context, target string, ids ...int) ([]roomsmodel.VideoStream, error) {
	q := sq.Select("routes", "`type`", "room_id").
		From("room_video_streams").
		Where(sq.Eq{"room_id": ids})

	rows, err := q.RunWith(r.db).QueryContext(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var res []roomsmodel.VideoStream
	for rows.Next() {
		var (
			roomID         int
			t, streamsJSON string
		)
		if err := rows.Scan(&streamsJSON, &t, &roomID); err != nil {
			return nil, err
		}
		vt := roomsmodel.ParseVideoTypeString(t)
		if vt == roomsmodel.Unknown {
			continue
		}
		if len(streamsJSON) > 0 {
			var routes map[string]string
			if err := json.NewDecoder(strings.NewReader(streamsJSON)).Decode(&routes); err != nil {
				return nil, err
			}
			if streamURI, ok := routes[target]; ok {
				res = append(res, roomsmodel.VideoStream{Type: vt, Stream: streamURI, RoomID: roomID})
			}
		}
	}
	return res, nil
}
