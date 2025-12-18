package storage

import (
	"context"

	"github.com/msorokin-hash/passkeeper/internal/entity"
)

// Closer defines an interface for closing the underlying storage connection.
type Closer interface {
	Close() error
}

// UserStorage defines operations related to managing user accounts.
type UserStorage interface {
	CreateUser(ctx context.Context, user *entity.User) (id string, err error)
	GetUserByLogin(ctx context.Context, login string) (user *entity.User, err error)
	GetUserByID(ctx context.Context, id string) (user *entity.User, err error)
}

// VaultStorage defines operations for storing and managing encrypted vault data.
type VaultStorage interface {
	CreateItem(ctx context.Context, userID string, item *entity.VaultItem) (string, error)
	GetItem(ctx context.Context, id string, userID string) (*entity.VaultItem, error)
	DeleteItem(ctx context.Context, id string, userID string) error
	GetItemsByType(ctx context.Context, dataType string, userID string) ([]entity.VaultItem, error)
	UpdateItem(ctx context.Context, id string, userID string, item *entity.VaultItem) error
}

//go:generate mockgen -source=storage.go -destination=mocks/storage_mock.go -package=mocks
type Storage interface {
	UserStorage
	VaultStorage
	Closer
}
