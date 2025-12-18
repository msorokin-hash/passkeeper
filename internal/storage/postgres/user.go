package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"github.com/msorokin-hash/passkeeper/internal/entity"
	errorscustom "github.com/msorokin-hash/passkeeper/internal/errors"
)

// CreateUser inserts a new user into the PostgreSQL storage.
// The login is normalized to lowercase before saving.
// Returns the generated user ID or an error.
func (p *PGStorage) CreateUser(ctx context.Context, user *entity.User) (id string, err error) {
	user.Login = strings.ToLower(user.Login)

	row := p.pool.QueryRow(
		ctx,
		"INSERT INTO users(login, password_hash, encrypted_secret) VALUES($1, $2, $3) RETURNING id;",
		user.Login, user.PasswordHash, user.EncryptedSecret,
	)
	if err := row.Scan(&user.ID); err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) {
			if pgError.Code == pgerrcode.UniqueViolation {
				return "", errorscustom.ErrLoginTaken
			}
		}
		return "", fmt.Errorf("failed to create new user: %w", err)
	}

	return user.ID, nil
}

// GetUserByID retrieves a user by their ID from PostgreSQL.
// Returns ErrNoUser if the user does not exist.
func (p *PGStorage) GetUserByID(ctx context.Context, id string) (user *entity.User, err error) {
	var u entity.User

	row := p.pool.QueryRow(
		ctx,
		"SELECT id, password_hash, encrypted_secret FROM users WHERE id = $1;",
		id,
	)
	if err := row.Scan(&u.ID, &u.PasswordHash, &u.EncryptedSecret); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorscustom.ErrNoUser
		}
		return nil, fmt.Errorf("failed to get user from db: %w", err)
	}

	u.ID = id

	return &u, nil
}

// GetUserByLogin retrieves a user by login from PostgreSQL.
// Returns ErrNoUser if the login is not registered.
func (p *PGStorage) GetUserByLogin(ctx context.Context, login string) (user *entity.User, err error) {
	var u entity.User

	row := p.pool.QueryRow(
		ctx,
		"SELECT id, password_hash, encrypted_secret FROM users WHERE login = $1;",
		login,
	)
	if err := row.Scan(&u.ID, &u.PasswordHash, &u.EncryptedSecret); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorscustom.ErrNoUser
		}
		return nil, fmt.Errorf("failed to get user from db: %w", err)
	}

	u.Login = login

	return &u, nil
}
