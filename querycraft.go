package querycraft

import (
	"database/sql"
	"fmt"
	
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
}

type queryCraft struct {
	db      *sqlx.DB
	dialect Dialect
}

func New(driver string, db *sql.DB) (QueryCraft, error) {
	sqlxDB := sqlx.NewDb(db, driver)
	
	qc := &queryCraft{
		db: sqlxDB,
	}
	
	// Set dialect based on driver
	switch driver {
	case "mysql":
		qc.dialect = &MySQLDialect{}
	default:
		return nil, fmt.Errorf("unsupported driver: %s", driver)
	}
	
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
	return NewTransaction(tx, qc.dialect), nil
}

func (qc *queryCraft) GetDB() *sqlx.DB {
	return qc.db
}