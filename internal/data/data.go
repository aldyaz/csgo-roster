package data

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type key int

const (
	txKey key = 0
)

// Queryer represents the database commands interface
type Queryer interface {
	PrepareNamed(query string) (*sqlx.NamedStmt, error)
	Rebind(query string) string
	Select(dest interface{}, query string, args ...interface{}) error
	Get(dest interface{}, query string, args ...interface{}) error
}

// Manager represents the manager to manage the data consistency
type Manager struct {
	db *sqlx.DB
}

// NewManager creates a new manager
func NewManager(db *sqlx.DB) *Manager {
	return &Manager{
		db: db,
	}
}

// RunInTransaction runs the f with the transaction queryable inside the context
func (m *Manager) RunInTransaction(ctx context.Context, f func(tctx context.Context) error) error {
	tx, err := m.db.Beginx()
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error when creating transction: %v", err)
	}

	ctx = newContext(ctx, tx)
	err = f(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("error when committing transaction: %v", err)
	}

	return nil
}

// newContext creates a new database context
func newContext(ctx context.Context, q Queryer) context.Context {
	ctx = context.WithValue(ctx, txKey, q)
	return ctx
}
