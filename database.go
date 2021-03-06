package amyssh

import (
	"fmt"

	"database/sql"
	_ "github.com/go-sql-driver/mysql"
)

var _ = fmt.Printf // deleteme

type Connection struct {
	db *sql.DB
}

func dbStr(cfg DatabaseConfig) string {
	return fmt.Sprintf("%s:%s@(%s:%d)/%s", cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DbName)
}

func NewCon(cfg *Config) (*Connection, error) {
	con := Connection{}
	var err error
	con.db, err = sql.Open("mysql", dbStr(cfg.Database))

	if err != nil {
		return nil, err
	}
	return &con, nil
}

func generatePlaceholder(count int) string {
	if count > 0 {
		maxChars := count*2 - 1
		buf := make([]byte, maxChars)
		for i := 0; i < maxChars; i++ {
			if (i % 2) == 0 {
				buf[i] = '?'
			} else {
				buf[i] = ','
			}

		}
		return string(buf[:maxChars])
	}
	return ""
}

type KeyData struct {
}

func (con *Connection) FetchKeys(hostTags []string, userTags []string) (keys map[string]StringSet, err error) {
	// TODO: find better way to use prepared statement escaping

	hostLen := len(hostTags)
	params := make([]interface{}, hostLen+len(userTags))
	for i, v := range hostTags {
		params[i] = interface{}(v)
	}
	for i, v := range userTags {
		params[i+hostLen] = interface{}(v)
	}

	query := fmt.Sprintf("SELECT DISTINCT k.key_id, k.`key`, u.label FROM `keys` k "+
		"JOIN host_tags h ON h.label IN (%s) AND k.key_id = h.key_id "+
		"JOIN user_host_tags u ON u.label IN (%s) AND u.host_tag_id = h.host_tag_id",
		generatePlaceholder(len(hostTags)), generatePlaceholder(len(userTags)))

	row, err := con.db.Query(query, params...)
	if err != nil {
		return nil, err
	}

	userKeys := make(map[string]StringSet)

	var id, key, userLabel string
	for row.Next() {
		row.Scan(&id, &key, &userLabel)
		if userKeys[userLabel] == nil {
			userKeys[userLabel] = make(StringSet)
		}

		userKeys[userLabel][key] = struct{}{}
	}
	return userKeys, nil
}
