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

// UpdateBuilder - интерфейс для UPDATE запросов
type UpdateBuilder interface {
	// Установка значений
	Set(column string, value any) UpdateBuilder
	SetRaw(expression string, args ...any) UpdateBuilder
	SetMap(values map[string]any) UpdateBuilder
	SetStruct(data any) UpdateBuilder
	Columns(columns ...string) UpdateBuilder

	// Инкремент/декремент
	Increment(column string, value ...int) UpdateBuilder
	Decrement(column string, value ...int) UpdateBuilder

	// WHERE условия (те же что и в SelectBuilder)
	Where(column, operator string, value any) UpdateBuilder
	WhereEq(column string, value any) UpdateBuilder
	WhereIn(column string, values ...any) UpdateBuilder
	WhereRaw(condition string, args ...any) UpdateBuilder

	// Условное обновление
	When(condition bool, column string, value any) UpdateBuilder
	WhenFunc(condition bool, fn func(UpdateBuilder) UpdateBuilder) UpdateBuilder

	// JOIN для UPDATE
	Join(table, condition string) UpdateBuilder
	LeftJoin(table, condition string) UpdateBuilder

	// Выполнение
	Exec() (sql.Result, error)

	// Утилиты
	WithContext(ctx context.Context) UpdateBuilder
	ToSQL() (string, []any)
	PrintSQL() UpdateBuilder
	Clone() UpdateBuilder
}

type updateBuilder struct {
	db      SQLXExecutor
	dialect dialect.Dialect
	ctx     context.Context
	logger  Logger

	table     string
	sets      []string
	setArgs   []any
	joins     []string
	wheres    []string
	whereArgs []any
	limit     *int
	columns   []string

	// Print SQL flag
	printSQL bool
}

func NewUpdateBuilder(db SQLXExecutor, dialect dialect.Dialect, table string) UpdateBuilder {
	return &updateBuilder{
		db:      db,
		dialect: dialect,
		ctx:     context.Background(),
		table:   table,
	}
}

func (u *updateBuilder) Set(column string, value any) UpdateBuilder {
	u.sets = append(u.sets, fmt.Sprintf("%s = %s", u.dialect.QuoteIdentifier(column), u.dialect.PlaceholderFormat()))
	u.setArgs = append(u.setArgs, value)
	return u
}

func (u *updateBuilder) SetRaw(expression string, args ...any) UpdateBuilder {
	u.sets = append(u.sets, expression)
	u.setArgs = append(u.setArgs, args...)
	return u
}

func (u *updateBuilder) SetMap(values map[string]any) UpdateBuilder {
	for col, val := range values {
		u.Set(col, val)
	}
	return u
}

func (u *updateBuilder) SetStruct(data any) UpdateBuilder {
	// Check for nil data
	if data == nil {
		return u
	}

	// Use reflection to extract fields and their values
	v := reflect.ValueOf(data)
	t := reflect.TypeOf(data)

	// Handle pointers
	if v.Kind() == reflect.Ptr {
		// Check for nil pointer
		if v.IsNil() {
			return u
		}
		v = v.Elem()
		t = t.Elem()
	}

	// Only process structs
	if v.Kind() != reflect.Struct {
		return u
	}

	// Create a map of column names for filtering if columns are specified
	columnMap := make(map[string]bool)
	useColumnFilter := len(u.columns) > 0
	if useColumnFilter {
		for _, col := range u.columns {
			columnMap[col] = true
		}
	}

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Get column name from struct tag or field name
		column := field.Name
		if tag := field.Tag.Get("db"); tag != "" {
			column = tag
		}

		// Skip fields with "-" tag
		if column == "-" {
			continue
		}

		// If columns are specified, only update those columns
		if useColumnFilter && !columnMap[column] {
			continue
		}

		// Set the value
		u.Set(column, value.Interface())
	}

	return u
}

func (u *updateBuilder) Columns(columns ...string) UpdateBuilder {
	u.columns = columns
	return u
}

func (u *updateBuilder) Increment(column string, value ...int) UpdateBuilder {
	inc := 1
	if len(value) > 0 {
		inc = value[0]
	}

	u.sets = append(u.sets, fmt.Sprintf("%s = %s + %s",
		u.dialect.QuoteIdentifier(column),
		u.dialect.QuoteIdentifier(column),
		u.dialect.PlaceholderFormat()))
	u.setArgs = append(u.setArgs, inc)
	return u
}

func (u *updateBuilder) Decrement(column string, value ...int) UpdateBuilder {
	dec := 1
	if len(value) > 0 {
		dec = value[0]
	}

	u.sets = append(u.sets, fmt.Sprintf("%s = %s - %s",
		u.dialect.QuoteIdentifier(column),
		u.dialect.QuoteIdentifier(column),
		u.dialect.PlaceholderFormat()))
	u.setArgs = append(u.setArgs, dec)
	return u
}

func (u *updateBuilder) Where(column, operator string, value any) UpdateBuilder {
	u.wheres = append(u.wheres, fmt.Sprintf("%s %s %s", u.dialect.QuoteIdentifier(column), operator, u.dialect.PlaceholderFormat()))
	u.whereArgs = append(u.whereArgs, value)
	return u
}

func (u *updateBuilder) WhereEq(column string, value any) UpdateBuilder {
	return u.Where(column, "=", value)
}

func (u *updateBuilder) WhereIn(column string, values ...any) UpdateBuilder {
	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = u.dialect.PlaceholderFormat()
	}
	u.wheres = append(u.wheres, fmt.Sprintf("%s IN (%s)", u.dialect.QuoteIdentifier(column), strings.Join(placeholders, ", ")))
	u.whereArgs = append(u.whereArgs, values...)
	return u
}

func (u *updateBuilder) WhereRaw(condition string, args ...any) UpdateBuilder {
	u.wheres = append(u.wheres, condition)
	u.whereArgs = append(u.whereArgs, args...)
	return u
}

func (u *updateBuilder) When(condition bool, column string, value any) UpdateBuilder {
	if condition {
		return u.Where(column, "=", value)
	}
	return u
}

func (u *updateBuilder) WhenFunc(condition bool, fn func(UpdateBuilder) UpdateBuilder) UpdateBuilder {
	if condition {
		return fn(u)
	}
	return u
}

func (u *updateBuilder) Join(table, condition string) UpdateBuilder {
	u.joins = append(u.joins, fmt.Sprintf("JOIN %s ON %s", table, condition))
	return u
}

func (u *updateBuilder) LeftJoin(table, condition string) UpdateBuilder {
	u.joins = append(u.joins, fmt.Sprintf("LEFT JOIN %s ON %s", table, condition))
	return u
}

func (u *updateBuilder) buildSQL() (string, []any) {
	var queryParts []string
	var args []any

	// UPDATE
	queryParts = append(queryParts, "UPDATE", u.table)

	// JOIN
	if len(u.joins) > 0 {
		queryParts = append(queryParts, strings.Join(u.joins, " "))
	}

	// SET
	if len(u.sets) > 0 {
		queryParts = append(queryParts, "SET", strings.Join(u.sets, ", "))
		args = append(args, u.setArgs...)
	}

	// WHERE
	if len(u.wheres) > 0 {
		queryParts = append(queryParts, "WHERE", strings.Join(u.wheres, " "))
		args = append(args, u.whereArgs...)
	}

	// LIMIT
	if u.limit != nil {
		queryParts = append(queryParts, u.dialect.UpdateLimit(*u.limit))
	}

	return strings.Join(queryParts, " "), args
}

func (u *updateBuilder) ToSQL() (string, []any) {
	return u.buildSQL()
}

func (u *updateBuilder) PrintSQL() UpdateBuilder {
	u.printSQL = true
	return u
}

func (u *updateBuilder) setLogger(logger Logger) {
	u.logger = logger
}

func (u *updateBuilder) Exec() (sql.Result, error) {
	sql, args := u.buildSQL()

	// Print SQL if needed
	if u.printSQL {
		// Simple placeholder replacement for debugging
		formattedSQL := sql
		for _, arg := range args {
			formattedSQL = strings.Replace(formattedSQL, u.dialect.PlaceholderFormat(), fmt.Sprintf("'%v'", arg), 1)
		}
		fmt.Println(formattedSQL)
	}

	// Log query if logger is set
	var start time.Time
	if u.logger != nil {
		start = time.Now()
	}

	result, err := u.db.ExecContext(u.ctx, sql, args...)

	// Log query execution
	if u.logger != nil {
		duration := time.Since(start)
		u.logger.LogQuery(u.ctx, sql, args, duration, err)
	}

	return result, err
}

func (u *updateBuilder) WithContext(ctx context.Context) UpdateBuilder {
	u.ctx = ctx
	return u
}

func (u *updateBuilder) Clone() UpdateBuilder {
	// Create a deep copy
	clone := &updateBuilder{
		db:      u.db,
		dialect: u.dialect,
		ctx:     u.ctx,
		table:   u.table,
		sets:    make([]string, len(u.sets)),
		wheres:  make([]string, len(u.wheres)),
		joins:   make([]string, len(u.joins)),
		limit:   u.limit,
		columns: make([]string, len(u.columns)),
	}

	copy(clone.sets, u.sets)
	copy(clone.wheres, u.wheres)
	copy(clone.joins, u.joins)
	copy(clone.columns, u.columns)

	// Copy args slices
	clone.setArgs = make([]any, len(u.setArgs))
	copy(clone.setArgs, u.setArgs)

	clone.whereArgs = make([]any, len(u.whereArgs))
	copy(clone.whereArgs, u.whereArgs)

	return clone
}
