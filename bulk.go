package querycraft

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"reflect"
	"strings"

	"github.com/antibomberman/querycraft/dialect"
)

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

type bulkBuilder struct {
	db      SQLXExecutor
	dialect dialect.Dialect
	ctx     context.Context
}

func NewBulkBuilder(db SQLXExecutor, dialect dialect.Dialect) BulkBuilder {
	return &bulkBuilder{
		db:      db,
		dialect: dialect,
		ctx:     context.Background(),
	}
}

func (b *bulkBuilder) WithContext(ctx context.Context) BulkBuilder {
	b.ctx = ctx
	return b
}

// Bulk Insert
func (b *bulkBuilder) BulkInsert(table string, data any, opts ...BulkOption) error {
	config := &BulkConfig{
		BatchSize:  1000,
		OnConflict: ConflictError,
	}
	for _, opt := range opts {
		opt(config)
	}

	// Convert data to slice of maps
	rows, err := b.convertToMapSlice(data)
	if err != nil {
		return err
	}

	// Process in batches
	for i := 0; i < len(rows); i += config.BatchSize {
		end := i + config.BatchSize
		if end > len(rows) {
			end = len(rows)
		}
		batch := rows[i:end]

		// Get columns from first row
		if len(batch) == 0 {
			continue
		}

		var columns []string
		for col := range batch[0] {
			columns = append(columns, col)
		}

		// Prepare values
		var values []any
		for _, row := range batch {
			for _, col := range columns {
				values = append(values, row[col])
			}
		}

		// Generate SQL
		query := b.generateBulkInsertSQL(table, columns, len(batch))

		// Handle conflicts
		switch config.OnConflict {
		case ConflictIgnore:
			query = fmt.Sprintf("%s %s", query, "IGNORE")
		case ConflictReplace:
			// For MySQL, this would be REPLACE INTO
			query = strings.Replace(query, "INSERT", "REPLACE", 1)
		}

		// Execute
		_, err := b.db.ExecContext(b.ctx, query, values...)
		if err != nil && !config.IgnoreErrors {
			return err
		}
	}

	return nil
}

func (b *bulkBuilder) BulkUpdate(table string, data any, opts ...BulkOption) error {
	config := &BulkConfig{
		BatchSize: 1000,
	}
	for _, opt := range opts {
		opt(config)
	}

	// Convert data to slice of maps
	rows, err := b.convertToMapSlice(data)
	if err != nil {
		return err
	}

	// Process in batches
	for i := 0; i < len(rows); i += config.BatchSize {
		end := i + config.BatchSize
		if end > len(rows) {
			end = len(rows)
		}
		batch := rows[i:end]

		// Get columns from first row
		if len(batch) == 0 {
			continue
		}

		var columns []string
		for col := range batch[0] {
			columns = append(columns, col)
		}

		// Prepare values
		var values []any
		for _, row := range batch {
			for _, col := range columns {
				values = append(values, row[col])
			}
		}

		// Generate SQL
		query := b.generateBulkUpdateSQL(table, columns, len(batch))

		// Execute
		_, err := b.db.ExecContext(b.ctx, query, values...)
		if err != nil && !config.IgnoreErrors {
			return err
		}
	}

	return nil
}

func (b *bulkBuilder) BulkDelete(table string, conditions []map[string]any) error {
	if len(conditions) == 0 {
		return nil
	}

	// Generate SQL
	query, args := b.dialect.BulkDelete(table, conditions)

	// Execute
	_, err := b.db.ExecContext(b.ctx, query, args...)
	return err
}

func (b *bulkBuilder) BulkUpsert(table string, data any, conflictColumns []string, opts ...BulkOption) error {
	config := &BulkConfig{
		BatchSize: 1000,
	}
	for _, opt := range opts {
		opt(config)
	}

	// Convert data to slice of maps
	rows, err := b.convertToMapSlice(data)
	if err != nil {
		return err
	}

	// Process in batches
	for i := 0; i < len(rows); i += config.BatchSize {
		end := i + config.BatchSize
		if end > len(rows) {
			end = len(rows)
		}
		batch := rows[i:end]

		// Get columns from first row
		if len(batch) == 0 {
			continue
		}

		var columns []string
		for col := range batch[0] {
			columns = append(columns, col)
		}

		// Prepare values
		var values []any
		for _, row := range batch {
			for _, col := range columns {
				values = append(values, row[col])
			}
		}

		// Generate SQL
		query := b.generateBulkUpsertSQL(table, columns, conflictColumns, len(batch))

		// Execute
		_, err := b.db.ExecContext(b.ctx, query, values...)
		if err != nil && !config.IgnoreErrors {
			return err
		}
	}

	return nil
}

// Bulk Update by key
func (b *bulkBuilder) BulkUpdateByKey(table string, data any, keyColumn string) error {
	// Convert data to slice of maps
	rows, err := b.convertToMapSlice(data)
	if err != nil {
		return err
	}

	if len(rows) == 0 {
		return nil
	}

	// Get all columns except the key column
	var columns []string
	for col := range rows[0] {
		if col != keyColumn {
			columns = append(columns, col)
		}
	}

	// Prepare values
	var values []any
	for _, row := range rows {
		// Add update values
		for _, col := range columns {
			values = append(values, row[col])
		}
		// Add key value
		values = append(values, row[keyColumn])
	}

	// Generate SQL
	query := b.generateBulkUpdateByKeySQL(table, columns, keyColumn, len(rows))

	// Execute
	_, err = b.db.ExecContext(b.ctx, query, values...)
	return err
}

// Process in batches
func (b *bulkBuilder) ProcessInBatches(query SelectBuilder, batchSize int, processor func(batch any) error) error {
	offset := 0
	for {
		// Clone the query and add limit/offset
		batchQuery := query.Clone()
		batchQuery = batchQuery.Limit(batchSize).Offset(offset)

		// Execute and get results
		rows, err := batchQuery.Rows()
		if err != nil {
			return err
		}

		// If no more rows, break
		if len(rows) == 0 {
			break
		}

		// Process batch
		err = processor(rows)
		if err != nil {
			return err
		}

		// If we got less than batchSize rows, we're done
		if len(rows) < batchSize {
			break
		}

		// Move to next batch
		offset += batchSize
	}

	return nil
}

// CSV Import/Export
func (b *bulkBuilder) ImportCSV(table string, csvPath string, mapping map[string]string) error {
	file, err := os.Open(csvPath)
	if err != nil {
		return err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return err
	}

	if len(records) < 1 {
		return nil // No data
	}

	// First row is headers
	headers := records[0]
	records = records[1:]

	// Map headers if mapping is provided
	columns := make([]string, len(headers))
	if len(mapping) > 0 {
		for i, header := range headers {
			if mapped, exists := mapping[header]; exists {
				columns[i] = mapped
			} else {
				columns[i] = header
			}
		}
	} else {
		copy(columns, headers)
	}

	// Prepare data as slice of maps
	var data []map[string]any
	for _, record := range records {
		row := make(map[string]any)
		for i, value := range record {
			row[columns[i]] = value
		}
		data = append(data, row)
	}

	// Bulk insert
	return b.BulkInsert(table, data)
}

func (b *bulkBuilder) ExportCSV(query SelectBuilder, csvPath string) error {
	// Execute query
	rows, err := query.Rows()
	if err != nil {
		return err
	}

	// Create file
	file, err := os.Create(csvPath)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	if len(rows) == 0 {
		return nil
	}

	// Write headers
	var headers []string
	for col := range rows[0] {
		headers = append(headers, col)
	}
	if err := writer.Write(headers); err != nil {
		return err
	}

	// Write data
	for _, row := range rows {
		var record []string
		for _, header := range headers {
			record = append(record, fmt.Sprintf("%v", row[header]))
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}

// Helper methods
func (b *bulkBuilder) convertToMapSlice(data any) ([]map[string]any, error) {
	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Slice:
		var result []map[string]any
		for i := 0; i < v.Len(); i++ {
			item := v.Index(i)
			if item.Kind() == reflect.Ptr {
				item = item.Elem()
			}

			if item.Kind() == reflect.Struct {
				m := b.structToMap(item)
				result = append(result, m)
			} else if item.Kind() == reflect.Map {
				// Assume it's already a map[string]any
				m := make(map[string]any)
				for _, key := range item.MapKeys() {
					if key.Kind() == reflect.String {
						m[key.String()] = item.MapIndex(key).Interface()
					}
				}
				result = append(result, m)
			}
		}
		return result, nil
	case reflect.Struct:
		m := b.structToMap(v)
		return []map[string]any{m}, nil
	case reflect.Map:
		// Assume it's a map[string]any
		m := make(map[string]any)
		for _, key := range v.MapKeys() {
			if key.Kind() == reflect.String {
				m[key.String()] = v.MapIndex(key).Interface()
			}
		}
		return []map[string]any{m}, nil
	default:
		return nil, fmt.Errorf("unsupported data type: %T", data)
	}
}

func (b *bulkBuilder) structToMap(v reflect.Value) map[string]any {
	m := make(map[string]any)
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get column name from db tag or field name
		column := field.Name
		if tag := field.Tag.Get("db"); tag != "" {
			column = tag
		}

		m[column] = value.Interface()
	}

	return m
}

func (b *bulkBuilder) generateBulkInsertSQL(table string, columns []string, rowCount int) string {
	quotedColumns := make([]string, len(columns))
	for i, col := range columns {
		quotedColumns[i] = b.dialect.QuoteIdentifier(col)
	}

	// Create placeholders for one row
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = b.dialect.PlaceholderFormat()
	}
	rowPlaceholder := fmt.Sprintf("(%s)", strings.Join(placeholders, ", "))

	// Create all row placeholders
	var allPlaceholders []string
	for i := 0; i < rowCount; i++ {
		allPlaceholders = append(allPlaceholders, rowPlaceholder)
	}

	return fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		b.dialect.QuoteIdentifier(table),
		strings.Join(quotedColumns, ", "),
		strings.Join(allPlaceholders, ", "))
}

func (b *bulkBuilder) generateBulkUpdateSQL(table string, columns []string, rowCount int) string {
	quotedColumns := make([]string, len(columns))
	for i, col := range columns {
		quotedColumns[i] = b.dialect.QuoteIdentifier(col)
	}

	// Create placeholders for one row
	placeholders := make([]string, len(columns))
	for i := range placeholders {
		placeholders[i] = b.dialect.PlaceholderFormat()
	}
	rowPlaceholder := fmt.Sprintf("(%s)", strings.Join(placeholders, ", "))

	// Create all row placeholders
	var allPlaceholders []string
	for i := 0; i < rowCount; i++ {
		allPlaceholders = append(allPlaceholders, rowPlaceholder)
	}

	return fmt.Sprintf("UPDATE %s SET (%s) = VALUES (%s)",
		b.dialect.QuoteIdentifier(table),
		strings.Join(quotedColumns, ", "),
		strings.Join(quotedColumns, ", "))
}

func (b *bulkBuilder) generateBulkUpsertSQL(table string, columns []string, conflictColumns []string, rowCount int) string {
	insertSQL := b.generateBulkInsertSQL(table, columns, rowCount)

	// For MySQL, we'll use ON DUPLICATE KEY UPDATE
	var updates []string
	for _, col := range columns {
		isConflictColumn := false
		for _, conflictCol := range conflictColumns {
			if col == conflictCol {
				isConflictColumn = true
				break
			}
		}
		if !isConflictColumn {
			quotedCol := b.dialect.QuoteIdentifier(col)
			updates = append(updates, fmt.Sprintf("%s = VALUES(%s)", quotedCol, quotedCol))
		}
	}

	if len(updates) > 0 {
		insertSQL = fmt.Sprintf("%s ON DUPLICATE KEY UPDATE %s", insertSQL, strings.Join(updates, ", "))
	}

	return insertSQL
}

func (b *bulkBuilder) generateBulkUpdateByKeySQL(table string, columns []string, keyColumn string, rowCount int) string {
	quotedColumns := make([]string, len(columns))
	for i, col := range columns {
		quotedColumns[i] = b.dialect.QuoteIdentifier(col)
	}

	quotedKeyColumn := b.dialect.QuoteIdentifier(keyColumn)

	// Create SET clause
	var setParts []string
	for _, col := range quotedColumns {
		setParts = append(setParts, fmt.Sprintf("%s = %s", col, b.dialect.PlaceholderFormat()))
	}

	// Create WHERE clause for key
	whereClause := fmt.Sprintf("%s = %s", quotedKeyColumn, b.dialect.PlaceholderFormat())

	return fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		b.dialect.QuoteIdentifier(table),
		strings.Join(setParts, ", "),
		whereClause)
}

// BulkOption functions
func WithBatchSize(size int) BulkOption {
	return func(config *BulkConfig) {
		config.BatchSize = size
	}
}

func WithConflictAction(action ConflictAction) BulkOption {
	return func(config *BulkConfig) {
		config.OnConflict = action
	}
}

func WithIgnoreErrors(ignore bool) BulkOption {
	return func(config *BulkConfig) {
		config.IgnoreErrors = ignore
	}
}

func WithMaxRetries(retries int) BulkOption {
	return func(config *BulkConfig) {
		config.MaxRetries = retries
	}
}
