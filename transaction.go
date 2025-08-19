package querycraft

import (
	"context"
	"database/sql"

	"github.com/antibomberman/querycraft/dialect"
	"github.com/jmoiron/sqlx"
)

// Transaction interface extends QueryCraft interface for transaction-specific operations
type Transaction interface {
	QueryCraft

	// Transaction control
	Commit() error
	Rollback() error
	GetTx() *sqlx.Tx

	// Context
	WithContext(ctx context.Context) Transaction
}

type transaction struct {
	tx      *sqlx.Tx
	db      *sqlx.DB
	dialect dialect.Dialect
	ctx     context.Context
}

func NewTransaction(tx *sqlx.Tx, db *sqlx.DB, dialect dialect.Dialect) Transaction {
	return &transaction{
		tx:      tx,
		db:      db,
		dialect: dialect,
		ctx:     context.Background(),
	}
}

func (t *transaction) WithContext(ctx context.Context) Transaction {
	t.ctx = ctx
	return t
}

func (t *transaction) Commit() error {
	return t.tx.Commit()
}

func (t *transaction) Rollback() error {
	return t.tx.Rollback()
}

func (t *transaction) GetTx() *sqlx.Tx {
	return t.tx
}

// Implement QueryCraft interface methods
func (t *transaction) Select(columns ...string) SelectBuilder {
	return NewSelectBuilder(t.tx, t.dialect, columns...)
}

func (t *transaction) Insert(table string) InsertBuilder {
	return NewInsertBuilder(t.tx, t.dialect, table)
}

func (t *transaction) Upsert(table string) UpsertBuilder {
	return NewUpsertBuilder(t.tx, t.dialect, table)
}

func (t *transaction) Update(table string) UpdateBuilder {
	return NewUpdateBuilder(t.tx, t.dialect, table)
}

func (t *transaction) Delete(table string) DeleteBuilder {
	return NewDeleteBuilder(t.tx, t.dialect, table)
}

func (t *transaction) Raw(query string, args ...any) Raw {
	return NewRaw(t.tx, query, args...)
}

func (t *transaction) Begin() (Transaction, error) {
	// Nested transactions are not supported in most databases
	return nil, sql.ErrTxDone
}

func (t *transaction) GetDB() *sqlx.DB {
	return t.db
}

func (t *transaction) Bulk() BulkBuilder {
	return NewBulkBuilder(t.tx, t.dialect)
}
