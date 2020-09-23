package repository

import (
	"context"

	"github.com/dantin/cubit/model"
)

// Room defines room repository operations
type Room interface {
	// UpsertRoom inserts a new room entity into storage, or updates it if previously inserted.
	UpsertRoom(ctx context.Context, room *model.Room) error

	// DeleteRoom deletes a room entity from storage.
	DeleteRoom(ctx context.Context, username string) error

	// FetchRoom retrieves a room entity from storage.
	FetchRoom(ctx context.Context, username string) (*model.Room, error)

	// FetchRooms retrieves room entites from storage.
	FetchRooms(ctx context.Context, p int) (*model.Room, error)
}
