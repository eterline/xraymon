package database

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/eterline/xraymon/internal/domain"
	"github.com/eterline/xraymon/internal/utils/usecase"
)

var allowedFields = usecase.NewWhitelist(
	"log",
	"api",
	"dns",
	"routing",
	"policy",
	"inbounds",
	"outbounds",
	"transport",
	"stats",
	"reverse",
	"fakedns",
	"metrics",
	"observatory",
	"burstObservatory",
)

func clearConfig(ccf *domain.CoreConfiguration) {
	for key := range *ccf {
		if !allowedFields.Allowed(key) {
			delete(*ccf, key)
		}
	}
}

type sqlConfig struct {
	db *sql.DB
}

func NewSQLiteConfig(db *sql.DB) (*sqlConfig, error) {
	c := &sqlConfig{db: db}
	if err := c.init(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *sqlConfig) init() error {
	const query = `
	CREATE TABLE IF NOT EXISTS CoreConfig (
		key   TEXT PRIMARY KEY,
		value BLOB NOT NULL
	);`
	_, err := c.db.Exec(query)
	return err
}

func (c *sqlConfig) SaveConfig(cfg domain.CoreConfiguration) error {
	tx, err := c.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`
		INSERT INTO CoreConfig(key, value)
		VALUES(?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	clearConfig(&cfg)

	for key, value := range cfg {
		if !json.Valid(value) {
			return fmt.Errorf("invalid JSON for key %q", key)
		}

		if _, err := stmt.Exec(key, value); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (c *sqlConfig) LoadConfig() (domain.CoreConfiguration, error) {
	rows, err := c.db.Query(`
		SELECT key, value FROM CoreConfig
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cfg := make(domain.CoreConfiguration)

	for rows.Next() {
		var key string
		var value []byte

		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}

		cfg[key] = json.RawMessage(value)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return cfg, nil
}
