package querycraft

type BulkBuilder interface {
	// Bulk Insert
	BulkInsert(table string, data any, opts ...BulkOption) error
	BulkUpdate(table string, data any, opts ...BulkOption) error
	BulkDelete(table string, conditions []map[string]any) error
	BulkUpsert(table string, data any, conflictColumns []string, opts ...BulkOption) error

	// Bulk Update
	BulkUpdateByKey(table string, data any, keyColumn string) error

	// Batch processing
	ProcessInBatches(query SelectBuilder, batchSize int, processor func(batch any) error) error

	// CSV Import/Export
	ImportCSV(table string, csvPath string, mapping map[string]string) error
	ExportCSV(query SelectBuilder, csvPath string) error
}

// BulkOption - опции для bulk операций
type BulkOption func(*BulkConfig)

type BulkConfig struct {
	BatchSize    int
	OnConflict   ConflictAction
	IgnoreErrors bool
	MaxRetries   int
}

type ConflictAction int

const (
	ConflictIgnore ConflictAction = iota
	ConflictUpdate
	ConflictReplace
	ConflictError
)
