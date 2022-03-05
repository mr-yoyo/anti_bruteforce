//nolint:dupl
package db

import (
	"database/sql"
	"net"

	"github.com/jmoiron/sqlx"
	"github.com/mr-yoyo/anti_bruteforce/app/cmd/internal/domain"
)

func NewBlacklist() domain.IPListRepository {
	return &blackListRepository{GetDB()}
}

type blackListRepository struct {
	db *sqlx.DB
}

func (r *blackListRepository) Exists(address net.IP) (bool, error) {
	var result struct {
		Exists bool `db:"exists"`
	}

	err := r.db.Get(
		&result,
		`SELECT true AS exists FROM ip_blacklist WHERE address >>= $1::inet`,
		address.String(),
	)

	if err == sql.ErrNoRows {
		return false, nil
	}

	return result.Exists, err
}

func (r *blackListRepository) Add(address *net.IPNet) (*domain.AddressItem, error) {
	var result domain.AddressItem

	query := `
		INSERT INTO ip_blacklist (address) 
		SELECT $1::inet
		WHERE NOT EXISTS (
			SELECT id FROM ip_blacklist WHERE address >>= $1::inet 
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

func (r *blackListRepository) Delete(address *net.IPNet) (*domain.AddressItem, error) {
	var result domain.AddressItem
	query := `
		DELETE FROM ip_blacklist WHERE address = $1::inet 
		RETURNING id, address, added_at
	`

	row := r.db.QueryRow(query, address.String())
	err := row.Scan(&result.ID, &result.Address, &result.AddedAt)

	if err == sql.ErrNoRows {
		return nil, &domain.IPNotExistsError{}
	}

	return &result, err
}
