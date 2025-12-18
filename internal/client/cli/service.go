package cli

import (
	"context"

	"github.com/msorokin-hash/passkeeper/internal/entity"
)

// Service defines a unified interface that combines user authentication
// operations and vault data management.
type Service interface {
	UserService
	VaultService
}

// UserService describes methods responsible for user registration and
// authentication workflows.
type UserService interface {
	Register(ctx context.Context, login, password string) (token string, err error)
	Login(ctx context.Context, login, password string) (token string, err error)
}

// VaultService describes methods responsible for storing, retrieving,
// and managing encrypted user data inside the vault.
type VaultService interface {
	AddPassword(ctx context.Context, data string, meta entity.PasswordMeta) error
	AddText(ctx context.Context, data string, meta entity.TextMeta) error
	AddBankCard(ctx context.Context, data entity.BankCardData, meta entity.BankCardMeta) error
	AddFile(ctx context.Context, file string, comment string) error
	GetData(ctx context.Context, id string) (entity.DataType, any, error)
	GetAllByType(ctx context.Context, dataType entity.DataType) ([]entity.ItemInfo, error)
	DeleteData(ctx context.Context, id string) error
}
