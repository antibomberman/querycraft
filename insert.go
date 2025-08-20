package querycraft

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"

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

	// Утилиты
	WithContext(ctx context.Context) InsertBuilder
	ToSQL() (string, []any)
	PrintSQL() InsertBuilder
	Clone() InsertBuilder
}

type insertBuilder struct {
	db      SQLXExecutor
	dialect dialect.Dialect
	ctx     context.Context
	logger  Logger

	table               string
	columns             []string
	values              [][]any
	onConflict          string
	onConflictDoNothing bool
	fromSelect          SelectBuilder

	// Print SQL flag
	printSQL bool
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
	if len(values) == 1 {
		val := values[0]
		if val == nil {
			i.values = append(i.values, values)
			return i
		}
		v := reflect.ValueOf(val)
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				i.values = append(i.values, values)
				return i
			}
			v = v.Elem()
		}

		switch v.Kind() {
		case reflect.Struct:
			i.addStruct(v)
			return i
		case reflect.Slice:
			if _, ok := v.Interface().([]byte); ok {
				break
			}
			elemType := v.Type().Elem()
			isStructSlice := false
			if elemType.Kind() == reflect.Struct {
				isStructSlice = true
			} else if elemType.Kind() == reflect.Ptr && elemType.Elem().Kind() == reflect.Struct {
				isStructSlice = true
			}

			if isStructSlice {
				for j := 0; j < v.Len(); j++ {
					elem := v.Index(j)
					if elem.Kind() == reflect.Ptr {
						if elem.IsNil() {
							// Or add a row of nils? For now, just skip.
							continue
						}
						elem = elem.Elem()
					}
					i.addStruct(elem)
				}
				return i
			}
		default:
			panic("unsupported type")
		}
	}

	i.values = append(i.values, values)
	return i
}

func (i *insertBuilder) addStruct(v reflect.Value) {
	t := v.Type()

	if len(i.columns) == 0 {
		var columns []string
		var rowValues []any
		for j := 0; j < t.NumField(); j++ {
			field := t.Field(j)
			dbTag := field.Tag.Get("db")
			if dbTag != "" && dbTag != "-" {
				columns = append(columns, dbTag)
				rowValues = append(rowValues, v.Field(j).Interface())
			}
		}
		i.columns = columns
		i.values = append(i.values, rowValues)
		return
	}

	var rowValues []any
	fieldsByTag := make(map[string]reflect.Value)
	for j := 0; j < t.NumField(); j++ {
		field := t.Field(j)
		dbTag := field.Tag.Get("db")
		if dbTag != "" && dbTag != "-" {
			fieldsByTag[dbTag] = v.Field(j)
		}
	}

	for _, column := range i.columns {
		if fieldValue, ok := fieldsByTag[column]; ok && fieldValue.IsValid() {
			rowValues = append(rowValues, fieldValue.Interface())
		} else {
			rowValues = append(rowValues, nil)
		}
	}
	i.values = append(i.values, rowValues)
}

func (i *insertBuilder) ValuesMap(values map[string]any) InsertBuilder {
	// Extract columns and values in consistent order
	var columns []string
	var vals []any

	// Check for nil map
	if values != nil {
		for col, val := range values {
			columns = append(columns, col)
			vals = append(vals, val)
		}
	}

	i.columns = columns
	i.values = [][]any{vals}
	return i
}

func (i *insertBuilder) ValuesMaps(values []map[string]any) InsertBuilder {
	// Check for nil slice
	if values == nil || len(values) == 0 {
		return i
	}

	// Check for nil map in first element
	if values[0] == nil {
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
		// Check for nil map
		if m == nil {
			continue
		}

		var vals []any
		for _, col := range columns {
			vals = append(vals, m[col])
		}
		i.values = append(i.values, vals)
	}

	return i
}

func (i *insertBuilder) OnConflictDoNothing() InsertBuilder {
	// Для MySQL используем INSERT IGNORE
	// Это будет обработано в диалекте
	i.onConflictDoNothing = true
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
		db:                  i.db,
		dialect:             i.dialect,
		ctx:                 i.ctx,
		table:               i.table,
		columns:             make([]string, len(i.columns)),
		onConflict:          i.onConflict,
		onConflictDoNothing: i.onConflictDoNothing,
		fromSelect:          i.fromSelect,
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
	i.onConflict = "IGNORE"
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
	insertKeyword := "INSERT INTO"

	// Handle conflict resolution for IGNORE
	if i.onConflict == "IGNORE" {
		insertKeyword = i.dialect.InsertIgnore()
	}

	queryParts = append(queryParts, insertKeyword)
	queryParts = append(queryParts, i.dialect.QuoteIdentifier(i.table))

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

	// Handle conflict resolution for UPDATE
	if i.onConflict == "UPDATE" {
		// For MySQL, this would be ON DUPLICATE KEY UPDATE
		// This should be handled by the dialect
		queryParts = append(queryParts, i.dialect.InsertOnConflict(i.columns, i.columns, nil))
	} else if i.onConflictDoNothing {
		// Handle ON CONFLICT DO NOTHING
		onConflictClause := i.dialect.InsertOnConflictDoNothing()
		if onConflictClause != "" {
			queryParts = append(queryParts, onConflictClause)
		}
	}

	return strings.Join(queryParts, " "), args
}

func (i *insertBuilder) buildFromSelectSQL() (string, []any) {
	var queryParts []string
	var args []any

	// INSERT keyword will be determined by dialect
	insertKeyword := "INSERT INTO"

	// Handle conflict resolution for IGNORE
	if i.onConflict == "IGNORE" {
		insertKeyword = i.dialect.InsertIgnore()
	}

	queryParts = append(queryParts, insertKeyword)
	queryParts = append(queryParts, i.dialect.QuoteIdentifier(i.table))

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

	// Handle conflict resolution for UPDATE
	if i.onConflict == "UPDATE" {
		queryParts = append(queryParts, i.dialect.InsertOnConflict(i.columns, i.columns, nil))
	} else if i.onConflict == "IGNORE" && insertKeyword == "INSERT INTO" {
		// Special case for databases that don't have INSERT IGNORE syntax
		// but support ON CONFLICT DO NOTHING
		onConflictClause := i.dialect.InsertOnConflictDoNothing()
		if onConflictClause != "" {
			queryParts = append(queryParts, onConflictClause)
		}
	} else if i.onConflict == "IGNORE" && insertKeyword == "INSERT IGNORE" {
		// For MySQL, INSERT IGNORE already handles conflicts, no additional clause needed
	} else if i.onConflictDoNothing {
		// Handle ON CONFLICT DO NOTHING
		onConflictClause := i.dialect.InsertOnConflictDoNothing()
		if onConflictClause != "" {
			queryParts = append(queryParts, onConflictClause)
		}
	}

	return strings.Join(queryParts, " "), args
}

func (i *insertBuilder) ToSQL() (string, []any) {
	return i.buildSQL()
}

func (i *insertBuilder) PrintSQL() InsertBuilder {
	i.printSQL = true
	return i
}

func (i *insertBuilder) setLogger(logger Logger) {
	i.logger = logger
}

func (i *insertBuilder) Exec() (sql.Result, error) {
	sql, args := i.buildSQL()

	// Print SQL if needed
	if i.printSQL {
		// Simple placeholder replacement for debugging
		formattedSQL := sql
		for _, arg := range args {
			formattedSQL = strings.Replace(formattedSQL, i.dialect.PlaceholderFormat(), fmt.Sprintf("'%v'", arg), 1)
		}
		fmt.Println(formattedSQL)
	}

	// Log query if logger is set
	var start time.Time
	if i.logger != nil {
		start = time.Now()
	}

	result, err := i.db.ExecContext(i.ctx, sql, args...)

	// Log query execution
	if i.logger != nil {
		duration := time.Since(start)
		i.logger.LogQuery(i.ctx, sql, args, duration, err)
	}

	return result, err
}

func (i *insertBuilder) ExecReturnID() (int64, error) {
	result, err := i.Exec()
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}
