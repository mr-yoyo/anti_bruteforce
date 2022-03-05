package db

import (
	"log"
	"sync"

	"github.com/jmoiron/sqlx"
	"github.com/mr-yoyo/anti_bruteforce/app/cmd/internal/config"
)

type Conn struct {
	db   *sqlx.DB
	once sync.Once
}

var conn = &Conn{}

func (conn *Conn) connect() error {
	cfg := config.Get()

	db, err := sqlx.Open(
		cfg.GetString("db.driver"),
		cfg.GetStringSlice("db.dataSource")[0],
	)
	if err != nil {
		return err
	}

	db.SetMaxIdleConns(cfg.GetInt("db.maxIdleConns"))
	db.SetMaxOpenConns(cfg.GetInt("db.maxOpenConns"))
	db.SetConnMaxLifetime(cfg.GetDuration("db.connMaxLifetime"))

	conn.db = db

	return nil
}

func GetDB() *sqlx.DB {
	conn.once.Do(func() {
		err := conn.connect()
		if err != nil {
			log.Fatalf("cannot connect to db: %+v", err)
		}
	})
	return conn.db
}
