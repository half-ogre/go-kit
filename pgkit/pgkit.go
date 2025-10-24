package pgkit

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

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

// dbOptions holds both pool config and context options
type dbOptions struct {
	config *pgxpool.Config
	ctx    context.Context
}

// DBOption is a functional option for configuring NewDB
type DBOption func(*dbOptions)

// NewDB creates a new database connection pool from a connection string.
// It opens the connection pool and verifies it with a ping.
// Optional DBOption parameters can be provided to configure connection pooling.
func NewDB(connectionString string, opts ...DBOption) (DB, error) {
	config, err := pgxpool.ParseConfig(connectionString)
	if err != nil {
		return nil, kit.WrapError(err, "failed to parse database config")
	}

	options := &dbOptions{
		config: config,
		ctx:    context.Background(),
	}

	// Apply options
	for _, opt := range opts {
		opt(options)
	}

	pool, err := pgxpool.NewWithConfig(options.ctx, options.config)
	if err != nil {
		return nil, kit.WrapError(err, "failed to create connection pool")
	}

	if err := pool.Ping(options.ctx); err != nil {
		pool.Close()
		return nil, kit.WrapError(err, "failed to ping database")
	}

	return &poolDB{pool: pool}, nil
}

// WithPoolContext sets the context used for creating the connection pool
func WithPoolContext(ctx context.Context) DBOption {
	return func(opts *dbOptions) {
		opts.ctx = ctx
	}
}

// WithMaxConns sets the maximum number of connections in the pool
func WithMaxConns(n int32) DBOption {
	return func(opts *dbOptions) {
		opts.config.MaxConns = n
	}
}

// WithMinConns sets the minimum number of connections in the pool
func WithMinConns(n int32) DBOption {
	return func(opts *dbOptions) {
		opts.config.MinConns = n
	}
}

// WithMaxConnLifetime sets the maximum amount of time a connection may be reused
func WithMaxConnLifetime(d time.Duration) DBOption {
	return func(opts *dbOptions) {
		opts.config.MaxConnLifetime = d
	}
}

// WithMaxConnIdleTime sets the maximum amount of time a connection may be idle
func WithMaxConnIdleTime(d time.Duration) DBOption {
	return func(opts *dbOptions) {
		opts.config.MaxConnIdleTime = d
	}
}

// WithHealthCheckPeriod sets how frequently to check the health of idle connections
func WithHealthCheckPeriod(d time.Duration) DBOption {
	return func(opts *dbOptions) {
		opts.config.HealthCheckPeriod = d
	}
}

// poolDB wraps *pgxpool.Pool to implement the DB interface
type poolDB struct {
	pool *pgxpool.Pool
}

func (p *poolDB) QueryRow(query string, args ...any) Row {
	return p.pool.QueryRow(context.Background(), query, args...)
}

func (p *poolDB) Query(query string, args ...any) (Rows, error) {
	rows, err := p.pool.Query(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	return &pgxRows{rows: rows}, nil
}

func (p *poolDB) Exec(query string, args ...any) (sql.Result, error) {
	cmdTag, err := p.pool.Exec(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	return pgxResult{cmdTag: cmdTag}, nil
}

func (p *poolDB) Close() error {
	p.pool.Close()
	return nil
}

// pgxRows wraps pgx.Rows to implement the Rows interface
type pgxRows struct {
	rows pgx.Rows
}

func (p *pgxRows) Next() bool {
	return p.rows.Next()
}

func (p *pgxRows) Scan(dest ...any) error {
	return p.rows.Scan(dest...)
}

func (p *pgxRows) Close() error {
	p.rows.Close()
	return nil
}

func (p *pgxRows) Err() error {
	return p.rows.Err()
}

// pgxResult wraps pgconn.CommandTag to implement sql.Result
type pgxResult struct {
	cmdTag pgconn.CommandTag
}

func (p pgxResult) LastInsertId() (int64, error) {
	return 0, kit.WrapError(nil, "LastInsertId is not supported by pgx")
}

func (p pgxResult) RowsAffected() (int64, error) {
	return p.cmdTag.RowsAffected(), nil
}
