package repository

import (
	"context"

	roomsmodel "github.com/dantin/cubit/model/rooms"
)

// Room defines room repository operations
type Room interface {
	// UpsertRoom inserts a new room entity into storage, or updates it if previously inserted.
	//UpsertRoom(ctx context.Context, room *roomsmodel.Room) error

	// DeleteRoom deletes a room entity from storage.
	//DeleteRoom(ctx context.Context, name string) error

	// BindRoom binds a room to user.
	//BindRoom(ctx context.Context, room *roomsmodel.Room, username string) error

	// UnbindRoom unbinds a room from user.
	//UnbindRoom(ctx context.Context, room *roomsmodel.Room, username string) error

	// UpsertVideoStream insert a new video stream to room whose name is 'name' into storage, or updates it if previously inserted.
	//UpsertVideoStream(ctx context.Context, room *roomsmodel.Room, to roomsmodel.VideoType, stream *roomsmodel.VideoStream) error

	// DeleteVideoStream deletes a video stream from storage.
	//DeleteVideoStream(ctx context.Context, room *roomsmodel.Room, to roomsmodel.VideoType) error

	// FetchRoom retrieves a room entity from storage.
	FetchRoom(ctx context.Context, username string) (*roomsmodel.Room, error)

	// FetchRooms retrieves room entites from storage.
	FetchRooms(ctx context.Context, page int, pageSize int) ([]roomsmodel.Room, error)

	// CountRooms  returns current size of rooms.
	CountRooms(ctx context.Context) (int, error)

	// FetchQCStream retrieves qc room stream entites from storage.
	FetchQCStream(ctx context.Context, username string) (*roomsmodel.VideoStream, error)
}
