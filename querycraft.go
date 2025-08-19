package querycraft

import (
	"database/sql"
	"fmt"

	"github.com/antibomberman/querycraft/dialect"
	"github.com/jmoiron/sqlx"
)

type QueryCraft interface {
	// Builders
	Select(columns ...string) SelectBuilder
	Insert(table string) InsertBuilder
	Upsert(table string) UpsertBuilder
	Update(table string) UpdateBuilder
	Delete(table string) DeleteBuilder

	// Raw queries
	Raw(query string, args ...any) Raw

	// Transactions
	Begin() (Transaction, error)
	GetDB() *sqlx.DB

	// Bulk operations
	Bulk() BulkBuilder

	// Schema and migrations
	Schema() SchemaBuilder
	Migration() MigrationManager
}

type queryCraft struct {
	db         *sqlx.DB
	dialect    dialect.Dialect
	migrations MigrationManager
}

func New(driver string, db *sql.DB) (QueryCraft, error) {
	sqlxDB := sqlx.NewDb(db, driver)

	qc := &queryCraft{
		db: sqlxDB,
	}

	// Set dialect based on driver
	switch driver {
	case "mysql":
		qc.dialect = &dialect.MySQLDialect{}
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}

	// Initialize migration manager
	qc.migrations = NewMigrationManager(qc.db, qc.dialect)

	return qc, nil
}

func (qc *queryCraft) Select(columns ...string) SelectBuilder {
	return NewSelectBuilder(qc.db, qc.dialect, columns...)
}

func (qc *queryCraft) Insert(table string) InsertBuilder {
	return NewInsertBuilder(qc.db, qc.dialect, table)
}

func (qc *queryCraft) Upsert(table string) UpsertBuilder {
	return NewUpsertBuilder(qc.db, qc.dialect, table)
}

func (qc *queryCraft) Update(table string) UpdateBuilder {
	return NewUpdateBuilder(qc.db, qc.dialect, table)
}

func (qc *queryCraft) Delete(table string) DeleteBuilder {
	return NewDeleteBuilder(qc.db, qc.dialect, table)
}

func (qc *queryCraft) Raw(query string, args ...any) Raw {
	return NewRaw(qc.db, query, args...)
}

func (qc *queryCraft) Begin() (Transaction, error) {
	tx, err := qc.db.Beginx()
	if err != nil {
		return nil, err
	}
	return NewTransaction(tx, qc.db, qc.dialect), nil
}

func (qc *queryCraft) GetDB() *sqlx.DB {
	return qc.db
}

func (qc *queryCraft) Bulk() BulkBuilder {
	return NewBulkBuilder(qc.db, qc.dialect)
}

func (qc *queryCraft) Schema() SchemaBuilder {
	return NewSchemaBuilder(qc.db, qc.dialect)
}

func (qc *queryCraft) Migration() MigrationManager {
	return qc.migrations
}
