package querycraft

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/antibomberman/querycraft/dialect"
)

// InsertBuilder - интерфейс для INSERT запросов
type InsertBuilder interface {
	// Установка данных
	Columns(columns ...string) InsertBuilder
	Values(values ...any) InsertBuilder
	ValuesMap(values map[string]any) InsertBuilder
	ValuesMaps(values []map[string]any) InsertBuilder

	// Конфликты
	OnConflictDoNothing() InsertBuilder
	OnConflictDoUpdate(columns ...string) InsertBuilder
	Ignore() InsertBuilder
	Replace() InsertBuilder

	// INSERT FROM SELECT
	FromSelect(selectBuilder SelectBuilder) InsertBuilder

	// Выполнение
	Exec() (sql.Result, error)
	ExecReturnID() (int64, error)
	ExecReturnIDs() ([]int64, error)

	// Утилиты
	WithContext(ctx context.Context) InsertBuilder
	ToSQL() (string, []any)
	Debug() string
	Clone() InsertBuilder
}

type insertBuilder struct {
	db      SQLXExecutor
	dialect dialect.Dialect
	ctx     context.Context

	table      string
	columns    []string
	values     [][]any
	onConflict string
	fromSelect SelectBuilder
}

func NewInsertBuilder(db SQLXExecutor, dialect dialect.Dialect, table string) InsertBuilder {
	return &insertBuilder{
		db:      db,
		dialect: dialect,
		ctx:     context.Background(),
		table:   table,
	}
}

func (i *insertBuilder) Columns(columns ...string) InsertBuilder {
	i.columns = columns
	return i
}

func (i *insertBuilder) Values(values ...any) InsertBuilder {
	i.values = append(i.values, values)
	return i
}

func (i *insertBuilder) ValuesMap(values map[string]any) InsertBuilder {
	// Extract columns and values in consistent order
	var columns []string
	var vals []any

	for col, val := range values {
		columns = append(columns, col)
		vals = append(vals, val)
	}

	i.columns = columns
	i.values = [][]any{vals}
	return i
}

func (i *insertBuilder) ValuesMaps(values []map[string]any) InsertBuilder {
	if len(values) == 0 {
		return i
	}

	// Extract columns from first map
	var columns []string
	for col := range values[0] {
		columns = append(columns, col)
	}

	i.columns = columns

	// Add values for each map
	for _, m := range values {
		var vals []any
		for _, col := range columns {
			vals = append(vals, m[col])
		}
		i.values = append(i.values, vals)
	}

	return i
}

func (i *insertBuilder) OnConflictDoNothing() InsertBuilder {
	i.onConflict = "DO NOTHING"
	return i
}

func (i *insertBuilder) OnConflictDoUpdate(columns ...string) InsertBuilder {
	// For MySQL, this would be ON DUPLICATE KEY UPDATE
	// We'll implement this properly in the dialect
	i.onConflict = "UPDATE"
	return i
}

func (i *insertBuilder) FromSelect(selectBuilder SelectBuilder) InsertBuilder {
	i.fromSelect = selectBuilder
	return i
}

func (i *insertBuilder) WithContext(ctx context.Context) InsertBuilder {
	i.ctx = ctx
	return i
}

func (i *insertBuilder) Clone() InsertBuilder {
	// Create a deep copy
	clone := &insertBuilder{
		db:         i.db,
		dialect:    i.dialect,
		ctx:        i.ctx,
		table:      i.table,
		columns:    make([]string, len(i.columns)),
		onConflict: i.onConflict,
		fromSelect: i.fromSelect,
	}

	copy(clone.columns, i.columns)

	// Copy values
	clone.values = make([][]any, len(i.values))
	for j, row := range i.values {
		clone.values[j] = make([]any, len(row))
		copy(clone.values[j], row)
	}

	return clone
}

//Exec methods

func (i *insertBuilder) Ignore() InsertBuilder {
	// This will be handled by the dialect
	return i
}

func (i *insertBuilder) Replace() InsertBuilder {
	// This will be handled by the dialect
	return i
}
func (i *insertBuilder) buildSQL() (string, []any) {
	if i.fromSelect != nil {
		return i.buildFromSelectSQL()
	}

	return i.buildValuesSQL()
}

func (i *insertBuilder) buildValuesSQL() (string, []any) {
	var queryParts []string
	var args []any

	// INSERT keyword will be determined by dialect
	queryParts = append(queryParts, "INSERT INTO")
	queryParts = append(queryParts, i.table)

	if len(i.columns) > 0 {
		quotedColumns := make([]string, len(i.columns))
		for j, col := range i.columns {
			quotedColumns[j] = i.dialect.QuoteIdentifier(col)
		}
		queryParts = append(queryParts, fmt.Sprintf("(%s)", strings.Join(quotedColumns, ", ")))
	}

	if len(i.values) > 0 {
		var valueParts []string
		for _, row := range i.values {
			placeholders := make([]string, len(row))
			for j := range row {
				placeholders[j] = i.dialect.PlaceholderFormat()
			}
			valueParts = append(valueParts, fmt.Sprintf("(%s)", strings.Join(placeholders, ", ")))
			args = append(args, row...)
		}
		queryParts = append(queryParts, "VALUES", strings.Join(valueParts, ", "))
	}

	// Handle conflict resolution
	if i.onConflict != "" {
		switch i.onConflict {
		case "DO NOTHING":
			queryParts = append(queryParts, "ON CONFLICT DO NOTHING")
		case "UPDATE":
			// For MySQL, this would be ON DUPLICATE KEY UPDATE
			// For now, we'll add a placeholder
			queryParts = append(queryParts, "ON DUPLICATE KEY UPDATE")
		}
	}

	return strings.Join(queryParts, " "), args
}

func (i *insertBuilder) buildFromSelectSQL() (string, []any) {
	var queryParts []string
	var args []any

	// INSERT keyword will be determined by dialect
	queryParts = append(queryParts, "INSERT INTO")
	queryParts = append(queryParts, i.table)

	// Get SQL from select builder
	selectSQL, selectArgs := i.fromSelect.ToSQL()

	if len(i.columns) > 0 {
		quotedColumns := make([]string, len(i.columns))
		for j, col := range i.columns {
			quotedColumns[j] = i.dialect.QuoteIdentifier(col)
		}
		queryParts = append(queryParts, fmt.Sprintf("(%s)", strings.Join(quotedColumns, ", ")))
	}

	queryParts = append(queryParts, selectSQL)
	args = append(args, selectArgs...)

	return strings.Join(queryParts, " "), args
}

func (i *insertBuilder) ToSQL() (string, []any) {
	return i.buildSQL()
}

func (i *insertBuilder) Debug() string {
	sql, args := i.buildSQL()
	// Simple placeholder replacement for debugging
	for _, arg := range args {
		sql = strings.Replace(sql, i.dialect.PlaceholderFormat(), fmt.Sprintf("'%v'", arg), 1)
	}
	return sql
}

func (i *insertBuilder) Exec() (sql.Result, error) {
	sql, args := i.buildSQL()
	return i.db.ExecContext(i.ctx, sql, args...)
}

func (i *insertBuilder) ExecReturnID() (int64, error) {
	result, err := i.Exec()
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (i *insertBuilder) ExecReturnIDs() ([]int64, error) {
	// For multiple inserts, we can only return the first ID in most databases
	// This is a limitation of the SQL standard
	id, err := i.ExecReturnID()
	if err != nil {
		return nil, err
	}

	return []int64{id}, nil
}
