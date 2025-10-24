package pgkit

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/half-ogre/go-kit/kit"
)

// Row is an interface for scanning a single row result
type Row interface {
	Scan(dest ...any) error
}

// Rows is an interface for iterating over multiple row results
type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Close() error
	Err() error
}

// DB is an interface for database operations
type DB interface {
	QueryRow(query string, args ...any) Row
	Query(query string, args ...any) (Rows, error)
	Exec(query string, args ...any) (sql.Result, error)
	Close() error
}

// sqlDB wraps *sql.DB to implement the DB interface
type sqlDB struct {
	db *sql.DB
}

func (s *sqlDB) QueryRow(query string, args ...any) Row {
	return s.db.QueryRow(query, args...)
}

func (s *sqlDB) Query(query string, args ...any) (Rows, error) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	return &sqlRows{rows: rows}, nil
}

func (s *sqlDB) Exec(query string, args ...any) (sql.Result, error) {
	return s.db.Exec(query, args...)
}

func (s *sqlDB) Close() error {
	return s.db.Close()
}

// sqlRows wraps *sql.Rows to implement the Rows interface
type sqlRows struct {
	rows *sql.Rows
}

func (s *sqlRows) Next() bool {
	return s.rows.Next()
}

func (s *sqlRows) Scan(dest ...any) error {
	return s.rows.Scan(dest...)
}

func (s *sqlRows) Close() error {
	return s.rows.Close()
}

func (s *sqlRows) Err() error {
	return s.rows.Err()
}

// NewDB creates a new database connection from a connection string.
// It opens the connection and verifies it with a ping.
func NewDB(connectionString string) (DB, error) {
	db, err := sql.Open("pgx", connectionString)
	if err != nil {
		return nil, kit.WrapError(err, "failed to open database")
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, kit.WrapError(err, "failed to ping database")
	}

	return &sqlDB{db: db}, nil
}

// WrapDB wraps a *sql.DB to implement the DB interface.
// This is useful when you already have a *sql.DB instance.
// For most cases, use NewDB instead.
func WrapDB(db *sql.DB) DB {
	return &sqlDB{db: db}
}
