package domain

import "time"

type AddressItem struct {
	ID      int       `db:"id"`
	Address string    `db:"address"`
	AddedAt time.Time `db:"added_at"`
}
