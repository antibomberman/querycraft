package dialect

// Dialect interface defines methods for generating SQL for different databases
type Dialect interface {
	// Placeholders
	PlaceholderFormat() string
	Rebind(sql string) string

	// SELECT
	SelectLimit(limit int) string
	SelectOffset(offset int) string
	SelectOrderBy(column string, desc bool) string

	// INSERT
	InsertIgnore() string
	InsertReplace() string
	InsertOnConflict(columns []string, updateColumns []string, updateExcluded []string) string

	// UPDATE
	UpdateLimit(limit int) string

	// DELETE
	DeleteLimit(limit int) string

	// UPSERT
	Upsert(columns []string, values []any, conflictColumns []string, updateColumns []string) (string, []any)

	// BULK
	BulkInsert(table string, columns []string, values []any, batchSize int) (string, []any)
	BulkUpdate(table string, columns []string, values []any, keyColumn string) (string, []any)
	BulkDelete(table string, conditions []map[string]any) (string, []any)

	// QUOTES
	QuoteIdentifier(name string) string
}
