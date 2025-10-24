package pgkit

import (
	"context"
	"database/sql"
)

type FakeRow struct {
	ScanFake func(dest ...any) error
}

func (f *FakeRow) Scan(dest ...any) error {
	if f.ScanFake != nil {
		return f.ScanFake(dest...)
	}
	panic("Scan fake not implemented")
}

type FakeRows struct {
	NextFake  func() bool
	ScanFake  func(dest ...any) error
	CloseFake func() error
	ErrFake   func() error
}

func (f *FakeRows) Next() bool {
	if f.NextFake != nil {
		return f.NextFake()
	}
	panic("Next fake not implemented")
}

func (f *FakeRows) Scan(dest ...any) error {
	if f.ScanFake != nil {
		return f.ScanFake(dest...)
	}
	panic("Scan fake not implemented")
}

func (f *FakeRows) Close() error {
	if f.CloseFake != nil {
		return f.CloseFake()
	}
	panic("Close fake not implemented")
}

func (f *FakeRows) Err() error {
	if f.ErrFake != nil {
		return f.ErrFake()
	}
	panic("Err fake not implemented")
}

type FakeDB struct {
	QueryRowFake func(ctx context.Context, query string, args ...any) Row
	QueryFake    func(ctx context.Context, query string, args ...any) (Rows, error)
	ExecFake     func(ctx context.Context, query string, args ...any) (sql.Result, error)
	CloseFake    func() error
}

func (f *FakeDB) QueryRow(ctx context.Context, query string, args ...any) Row {
	if f.QueryRowFake != nil {
		return f.QueryRowFake(ctx, query, args...)
	}
	panic("QueryRow fake not implemented")
}

func (f *FakeDB) Query(ctx context.Context, query string, args ...any) (Rows, error) {
	if f.QueryFake != nil {
		return f.QueryFake(ctx, query, args...)
	}
	panic("Query fake not implemented")
}

func (f *FakeDB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if f.ExecFake != nil {
		return f.ExecFake(ctx, query, args...)
	}
	panic("Exec fake not implemented")
}

func (f *FakeDB) Close() error {
	if f.CloseFake != nil {
		return f.CloseFake()
	}
	panic("Close fake not implemented")
}

type FakeMigrator struct {
	RunMigrationsFake func(db DB, dirPath string) error
}

func (f *FakeMigrator) RunMigrations(db DB, dirPath string) error {
	if f.RunMigrationsFake != nil {
		return f.RunMigrationsFake(db, dirPath)
	}
	panic("RunMigrations fake not implemented")
}
