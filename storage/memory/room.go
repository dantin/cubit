package memorystorage

import (
	"context"

	roomsmodel "github.com/dantin/cubit/model/rooms"
)

// Room represents an in-memory room storage.
type Room struct {
	*memoryStorage
}

// NewRoom returns an instance of Room in-memory storage.
func NewRoom() *Room {
	return &Room{memoryStorage: newStorage()}
}

// FetchRoom retrieves a room entity from storage.
func (r *Room) FetchRoom(ctx context.Context, username string) (*roomsmodel.Room, error) {
	return nil, nil
}

// FetchRooms retrieves room entites from storage.
func (r *Room) FetchRooms(ctx context.Context, page int, pageSize int) ([]roomsmodel.Room, error) {
	return nil, nil
}

// CountRooms  returns current size of rooms.
func (r *Room) CountRooms(ctx context.Context) (int, error) {
	return 0, nil
}

// FetchQCStream retrieves qc room stream entites from storage.
func (r *Room) FetchQCStream(ctx context.Context, username string) (*roomsmodel.VideoStream, error) {
	return nil, nil
}
