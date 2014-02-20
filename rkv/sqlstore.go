package rkv

import (
	"database/sql"
	"fmt"
)

var (
	tableCreateStmt = `CREATE TABLE IF NOT EXISTS %s (
      name varchar(255) NOT NULL,
      rev bigint(20) NOT NULL,
      timestamp timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
      content blob NOT NULL,
      PRIMARY KEY (name)
    ) DEFAULT CHARSET=utf8;`
)

type SQLStore struct {
	db    *sql.DB
	table string
}

func NewSQLStore(driver, source, table string) (*SQLStore, error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	_, err = db.Exec(fmt.Sprintf(tableCreateStmt, table))
	if err != nil {
		return nil, err
	}
	return &SQLStore{db, table}, nil
}

func (s *SQLStore) Get(key string) (Value, error) {
	var p Value
	err := s.db.QueryRow(fmt.Sprintf("SELECT name, rev, timestamp, content from %s WHERE name = ?", s.table), key).Scan(&key, &p.Rev, &p.Timestamp, &p.Content)
	if err == sql.ErrNoRows {
		err = nil
	}
	return p, err
}

func (s *SQLStore) Put(key string, np Value) error {
	// try update first
	r, err := s.db.Exec(fmt.Sprintf("UPDATE %s SET rev=?, content=? WHERE name=? and rev=?", s.table), np.Rev, np.Content, key, np.Rev-1)
	if err != nil {
		return err
	}
	rows, err := r.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 1 {
		return nil
	}

	// either rev not match, or not existed
	// just use this insert for both cases
	r, err = s.db.Exec(fmt.Sprintf("INSERT INTO %s (name, rev, content) VALUES (?,?,?)", s.table), key, np.Rev, np.Content)
	if err != nil {
		return ErrRevNotMatch
	}
	rows, err = r.RowsAffected()
	if rows != 1 {
		return ErrRevNotMatch
	}
	if err != nil {
		return err
	}
	return nil

}
