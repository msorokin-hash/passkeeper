package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx"
	"github.com/msorokin-hash/passkeeper/internal/entity"
	errorscustom "github.com/msorokin-hash/passkeeper/internal/errors"
)

// CreateItem inserts a new vault item for the given user into PostgreSQL.
// Returns the generated item ID or an error if the insert fails.
func (p *PGStorage) CreateItem(ctx context.Context, userID string, item *entity.VaultItem) (string, error) {
	row := p.pool.QueryRow(
		ctx,
		`INSERT INTO user_data(user_id, encrypted_data, meta, data_type) VALUES($1, $2, $3, $4) RETURNING id;`,
		userID, item.EncryptedData, item.Meta, item.Type,
	)
	if err := row.Scan(&item.ID); err != nil {
		return "", fmt.Errorf("failed to create new item: %w", err)
	}

	return item.ID, nil
}

// DeleteItem removes a vault item by its ID for the given user.
// Returns ErrNoData if no rows were affected (item not found).
func (p *PGStorage) DeleteItem(ctx context.Context, id string, userID string) error {
	query, err := p.pool.Exec(ctx, `DELETE FROM user_data WHERE id = $1 AND user_id = $2;`, id, userID)
	if err != nil {
		return err
	}

	if query.RowsAffected() == 0 {
		return errorscustom.ErrNoData
	}

	return nil
}

// GetItem retrieves a single vault item by ID for the given user.
// Returns ErrNoData if the item does not exist.
func (p *PGStorage) GetItem(ctx context.Context, id string, userID string) (*entity.VaultItem, error) {
	var item entity.VaultItem

	row := p.pool.QueryRow(
		ctx,
		`SELECT encrypted_data, meta, data_type, created_at, updated_at FROM user_data WHERE id = $1 AND user_id = $2;`,
		id, userID,
	)
	if err := row.Scan(
		&item.EncryptedData, &item.Meta, &item.Type, &item.CreatedAt, &item.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errorscustom.ErrNoData
		}
		return nil, fmt.Errorf("failed to get data from db: %w", err)
	}

	item.ID = id
	item.UserID = userID

	return &item, nil
}

// GetItemsByType returns all vault items of the specified type for the given user.
// Returns an empty slice if no items are found.
func (p *PGStorage) GetItemsByType(ctx context.Context, dataType string, userID string) ([]entity.VaultItem, error) {
	var items []entity.VaultItem

	rows, err := p.pool.Query(ctx,
		`SELECT id, meta, created_at, updated_at FROM user_data WHERE user_id = $1 AND data_type = $2;`,
		userID, dataType,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get items by type: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var item entity.VaultItem
		if err = rows.Scan(&item.ID, &item.Meta, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to read data from db - item row: %w", err)
		}
		item.UserID = userID
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to get items by type: %w", err)
	}

	return items, nil
}

// UpdateItem updates an existing vault item for the given user.
// Returns ErrNoData if no rows were affected (item not found).
func (p *PGStorage) UpdateItem(ctx context.Context, id string, userID string, item *entity.VaultItem) error {
	query, err := p.pool.Exec(ctx,
		`UPDATE user_data SET encrypted_data = $1, meta = $2, updated_at = NOW() WHERE id = $3 AND user_id = $4;`,
		item.EncryptedData, item.Meta, id, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	if query.RowsAffected() == 0 {
		return errorscustom.ErrNoData
	}

	return nil
}
