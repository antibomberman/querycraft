package querycraft

import (
	"context"
	"database/sql"

	"github.com/antibomberman/querycraft/dialect"
	"github.com/jmoiron/sqlx"
)

// Transaction interface provides transaction-specific operations
type Transaction interface {
	// Builders
	Select(columns ...string) SelectBuilder
	Insert(table string) InsertBuilder
	Upsert(table string) UpsertBuilder
	Update(table string) UpdateBuilder
	Delete(table string) DeleteBuilder

	// Raw queries
	Raw(query string, args ...any) Raw

	// Bulk operations
	Bulk() BulkBuilder

	// Schema operations
	Schema() SchemaBuilder

	// Transaction control
	Commit() error
	Rollback() error
	GetTx() *sqlx.Tx

	// Context
	WithContext(ctx context.Context) Transaction

	// Logging
	SetLogger(logger Logger) Transaction
}

type transaction struct {
	tx      *sqlx.Tx
	db      *sqlx.DB
	dialect dialect.Dialect
	ctx     context.Context
	logger  Logger
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
	builder := NewSelectBuilder(t.tx, t.dialect, columns...)
	// Set logger if available
	if t.logger != nil {
		if sb, ok := builder.(*selectBuilder); ok {
			sb.setLogger(t.logger)
		}
	}
	return builder
}

func (t *transaction) Insert(table string) InsertBuilder {
	builder := NewInsertBuilder(t.tx, t.dialect, table)
	// Set logger if available
	if t.logger != nil {
		if ib, ok := builder.(*insertBuilder); ok {
			ib.setLogger(t.logger)
		}
	}
	return builder
}

func (t *transaction) Upsert(table string) UpsertBuilder {
	builder := NewUpsertBuilder(t.tx, t.dialect, table)
	// Set logger if available
	if t.logger != nil {
		if ub, ok := builder.(*upsertBuilder); ok {
			ub.setLogger(t.logger)
		}
	}
	return builder
}

func (t *transaction) Update(table string) UpdateBuilder {
	builder := NewUpdateBuilder(t.tx, t.dialect, table)
	// Set logger if available
	if t.logger != nil {
		if ub, ok := builder.(*updateBuilder); ok {
			ub.setLogger(t.logger)
		}
	}
	return builder
}

func (t *transaction) Delete(table string) DeleteBuilder {
	builder := NewDeleteBuilder(t.tx, t.dialect, table)
	// Set logger if available
	if t.logger != nil {
		if db, ok := builder.(*deleteBuilder); ok {
			db.setLogger(t.logger)
		}
	}
	return builder
}

func (t *transaction) Raw(query string, args ...any) Raw {
	builder := NewRaw(t.tx, query, args...)
	// Set logger if available
	if t.logger != nil {
		if rb, ok := builder.(*rawQuery); ok {
			rb.setLogger(t.logger)
		}
	}
	return builder
}

func (t *transaction) Begin() (Transaction, error) {
	// Nested transactions are not supported in most databases
	return nil, sql.ErrTxDone
}

func (t *transaction) GetDB() *sqlx.DB {
	return t.db
}

func (t *transaction) Bulk() BulkBuilder {
	builder := NewBulkBuilder(t.tx, t.dialect)
	// Set logger if available
	if t.logger != nil {
		if bb, ok := builder.(*bulkBuilder); ok {
			bb.setLogger(t.logger)
		}
	}
	return builder
}

func (t *transaction) Schema() SchemaBuilder {
	builder := NewSchemaBuilder(t.tx, t.dialect)
	// Set logger if available
	if t.logger != nil {
		if sb, ok := builder.(*schemaBuilder); ok {
			sb.setLogger(t.logger)
		}
	}
	return builder
}

func (t *transaction) Migration() MigrationManager {
	// В транзакции мы не можем управлять миграциями,
	// так как это может привести к блокировке таблиц миграций
	// Возвращаем nil или паникуем, чтобы показать, что это не поддерживается
	panic("Migration is not supported in transaction")
}

func (t *transaction) SetLogger(logger Logger) Transaction {
	t.logger = logger
	return t
}
