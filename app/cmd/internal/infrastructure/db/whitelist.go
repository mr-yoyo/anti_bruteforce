//nolint:dupl
package db

import (
	"database/sql"
	"net"

	"github.com/jmoiron/sqlx"
	"github.com/mr-yoyo/anti_bruteforce/app/cmd/internal/domain"
)

func NewWhitelist() domain.IPListRepository {
	return &whiteListRepository{GetDB()}
}

type whiteListRepository struct {
	db *sqlx.DB
}

func (r *whiteListRepository) Exists(address net.IP) (bool, error) {
	var result struct {
		Exists bool `db:"exists"`
	}

	err := r.db.Get(
		&result,
		`SELECT true AS exists FROM ip_whitelist WHERE address >>= $1::inet`,
		address.String(),
	)

	if err == sql.ErrNoRows {
		return false, nil
	}

	return result.Exists, err
}

func (r *whiteListRepository) Add(address *net.IPNet) (*domain.AddressItem, error) {
	var result domain.AddressItem

	query := `
		INSERT INTO ip_whitelist (address) 
		SELECT $1::inet
		WHERE NOT EXISTS (
			SELECT id FROM ip_whitelist WHERE address >>= $1::inet 
		)
		RETURNING id, address, added_at
	`

	row := r.db.QueryRow(query, address.String())
	err := row.Scan(&result.ID, &result.Address, &result.AddedAt)

	if err == sql.ErrNoRows {
		return nil, &domain.IPDuplicateError{}
	}

	return &result, err
}

func (r *whiteListRepository) Delete(address *net.IPNet) (*domain.AddressItem, error) {
	var result domain.AddressItem
	query := `
		DELETE FROM ip_whitelist WHERE address = $1::inet 
		RETURNING id, address, added_at
	`

	row := r.db.QueryRow(query, address.String())
	err := row.Scan(&result.ID, &result.Address, &result.AddedAt)

	if err == sql.ErrNoRows {
		return nil, &domain.IPNotExistsError{}
	}

	return &result, err
}
