package querycraft

import (
	"database/sql"
	"fmt"

	"github.com/antibomberman/querycraft/dialect"
	"github.com/jmoiron/sqlx"
)

// Options represents the options for QueryCraft
type Options struct {
	// Logger options
	LogEnabled        bool
	LogLevel          LogLevel
	LogFormat         LogFormat
	LogSaveToFile     bool
	LogPrintToConsole bool
	LogDir            string
	LogAutoCleanDays  int
}

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
	logger     Logger
}

func NewWithLogger(driver string, db *sql.DB, logger Logger) (QueryCraft, error) {
	sqlxDB := sqlx.NewDb(db, driver)

	qc := &queryCraft{
		db:     sqlxDB,
		logger: logger,
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

// DefaultOptions returns the default options for QueryCraft
func DefaultOptions() Options {
	return Options{
		LogEnabled:        false,
		LogLevel:          LogLevelInfo,
		LogFormat:         LogFormatText,
		LogSaveToFile:     true,
		LogPrintToConsole: false,
		LogDir:            "./storage/logs/sql/",
		LogAutoCleanDays:  7,
	}
}

func New(driver string, db *sql.DB, opts ...Options) (QueryCraft, error) {
	var options Options
	if len(opts) > 0 {
		options = opts[0]
	} else {
		options = DefaultOptions()
	}

	// Create logger if logging is enabled
	var logger Logger
	if options.LogEnabled {
		loggerOptions := LoggerOptions{
			Enabled:        true,
			Level:          options.LogLevel,
			Format:         options.LogFormat,
			SaveToFile:     options.LogSaveToFile,
			PrintToConsole: options.LogPrintToConsole,
			LogDir:         options.LogDir,
			AutoCleanDays:  options.LogAutoCleanDays,
		}
		logger = NewFileLogger(loggerOptions)
	}

	return NewWithLogger(driver, db, logger)
}

func (qc *queryCraft) Select(columns ...string) SelectBuilder {
	builder := NewSelectBuilder(qc.db, qc.dialect, columns...)
	// Set logger if available
	if qc.logger != nil {
		if sb, ok := builder.(*selectBuilder); ok {
			sb.setLogger(qc.logger)
		}
	}
	return builder
}

func (qc *queryCraft) Insert(table string) InsertBuilder {
	builder := NewInsertBuilder(qc.db, qc.dialect, table)
	// Set logger if available
	if qc.logger != nil {
		if ib, ok := builder.(*insertBuilder); ok {
			ib.setLogger(qc.logger)
		}
	}
	return builder
}

func (qc *queryCraft) Upsert(table string) UpsertBuilder {
	builder := NewUpsertBuilder(qc.db, qc.dialect, table)
	// Set logger if available
	if qc.logger != nil {
		if ub, ok := builder.(*upsertBuilder); ok {
			ub.setLogger(qc.logger)
		}
	}
	return builder
}

func (qc *queryCraft) Update(table string) UpdateBuilder {
	builder := NewUpdateBuilder(qc.db, qc.dialect, table)
	// Set logger if available
	if qc.logger != nil {
		if ub, ok := builder.(*updateBuilder); ok {
			ub.setLogger(qc.logger)
		}
	}
	return builder
}

func (qc *queryCraft) Delete(table string) DeleteBuilder {
	builder := NewDeleteBuilder(qc.db, qc.dialect, table)
	// Set logger if available
	if qc.logger != nil {
		if db, ok := builder.(*deleteBuilder); ok {
			db.setLogger(qc.logger)
		}
	}
	return builder
}

func (qc *queryCraft) Raw(query string, args ...any) Raw {
	builder := NewRaw(qc.db, query, args...)
	// Set logger if available
	if qc.logger != nil {
		if rb, ok := builder.(*rawQuery); ok {
			rb.setLogger(qc.logger)
		}
	}
	return builder
}

func (qc *queryCraft) Begin() (Transaction, error) {
	tx, err := qc.db.Beginx()
	if err != nil {
		return nil, err
	}

	transaction := NewTransaction(tx, qc.db, qc.dialect)
	// Set logger if available
	if qc.logger != nil {
		transaction = transaction.SetLogger(qc.logger)
	}

	return transaction, nil
}

func (qc *queryCraft) GetDB() *sqlx.DB {
	return qc.db
}

func (qc *queryCraft) Bulk() BulkBuilder {
	builder := NewBulkBuilder(qc.db, qc.dialect)
	// Set logger if available
	if qc.logger != nil {
		if bb, ok := builder.(*bulkBuilder); ok {
			bb.setLogger(qc.logger)
		}
	}
	return builder
}

func (qc *queryCraft) Schema() SchemaBuilder {
	builder := NewSchemaBuilder(qc.db, qc.dialect)
	// Set logger if available
	if qc.logger != nil {
		if sb, ok := builder.(*schemaBuilder); ok {
			sb.setLogger(qc.logger)
		}
	}
	return builder
}

func (qc *queryCraft) Migration() MigrationManager {
	return qc.migrations
}

func (qc *queryCraft) SetLogger(logger Logger) QueryCraft {
	qc.logger = logger
	return qc
}
