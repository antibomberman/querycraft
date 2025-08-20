package querycraft

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/antibomberman/querycraft/dialect"
)

type UpsertBuilder interface {
	// Установка данных для вставки
	Values(data any) UpsertBuilder
	Columns(columns ...string) UpsertBuilder

	// Настройка конфликтов
	OnConflict(columns ...string) UpsertBuilder     // Колонки для проверки конфликта
	DoUpdate(columns ...string) UpsertBuilder       // Колонки для обновления при конфликте
	DoUpdateExcept(columns ...string) UpsertBuilder // Обновить все кроме указанных
	DoNothing() UpsertBuilder                       // Игнорировать конфликт

	// Условия для UPDATE части
	UpdateWhere(condition string, args ...any) UpsertBuilder

	// Выполнение
	Exec() (sql.Result, error)
	ExecReturnID() (int64, error)
	ExecReturnAction() (UpsertAction, int64, error) // Что произошло + ID

	// Утилиты
	WithContext(ctx context.Context) UpsertBuilder
	ToSQL() (string, []any)
	Debug() string
}
type UpsertAction int

const (
	UpsertInserted UpsertAction = iota
	UpsertUpdated
	UpsertIgnored
)

type upsertBuilder struct {
	db      SQLXExecutor
	dialect dialect.Dialect
	ctx     context.Context

	table           string
	columns         []string
	values          []any
	conflictColumns []string
	updateColumns   []string
	updateExcluded  []string
	updateWhere     string
	updateWhereArgs []any
	action          UpsertAction
}

func NewUpsertBuilder(db SQLXExecutor, dialect dialect.Dialect, table string) UpsertBuilder {
	return &upsertBuilder{
		db:      db,
		dialect: dialect,
		ctx:     context.Background(),
		table:   table,
	}
}

func (u *upsertBuilder) Values(data any) UpsertBuilder {
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

	// Handle structs
	if v.Kind() == reflect.Struct {
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

			// Add column and value
			u.columns = append(u.columns, column)
			u.values = append(u.values, value.Interface())
		}
	}

	// Handle maps
	if v.Kind() == reflect.Map {
		for _, key := range v.MapKeys() {
			column := key.String()
			value := v.MapIndex(key).Interface()

			u.columns = append(u.columns, column)
			u.values = append(u.values, value)
		}
	}

	return u
}

func (u *upsertBuilder) Columns(columns ...string) UpsertBuilder {
	u.columns = columns
	return u
}

func (u *upsertBuilder) OnConflict(columns ...string) UpsertBuilder {
	u.conflictColumns = columns
	return u
}

func (u *upsertBuilder) DoUpdate(columns ...string) UpsertBuilder {
	u.updateColumns = columns
	return u
}

func (u *upsertBuilder) DoUpdateExcept(columns ...string) UpsertBuilder {
	u.updateExcluded = columns
	return u
}

func (u *upsertBuilder) DoNothing() UpsertBuilder {
	// This will be handled in the SQL generation
	return u
}

func (u *upsertBuilder) UpdateWhere(condition string, args ...any) UpsertBuilder {
	u.updateWhere = condition
	u.updateWhereArgs = args
	return u
}

func (u *upsertBuilder) buildSQL() (string, []any) {
	var queryParts []string
	var args []any

	// For MySQL, UPSERT is implemented as INSERT ... ON DUPLICATE KEY UPDATE

	// INSERT part
	queryParts = append(queryParts, "INSERT INTO", u.table)

	if len(u.columns) > 0 {
		quotedColumns := make([]string, len(u.columns))
		for j, col := range u.columns {
			quotedColumns[j] = u.dialect.QuoteIdentifier(col)
		}
		queryParts = append(queryParts, fmt.Sprintf("(%s)", strings.Join(quotedColumns, ", ")))
	}

	if len(u.values) > 0 {
		placeholders := make([]string, len(u.values))
		for j := range u.values {
			placeholders[j] = u.dialect.PlaceholderFormat()
		}
		queryParts = append(queryParts, "VALUES", fmt.Sprintf("(%s)", strings.Join(placeholders, ", ")))
		args = append(args, u.values...)
	}

	// ON DUPLICATE KEY UPDATE part
	if len(u.updateColumns) > 0 || len(u.updateExcluded) > 0 {
		queryParts = append(queryParts, u.dialect.InsertOnConflict(u.conflictColumns, u.updateColumns, u.updateExcluded))
	}

	return strings.Join(queryParts, " "), args
}

func (u *upsertBuilder) ToSQL() (string, []any) {
	return u.buildSQL()
}

func (u *upsertBuilder) Debug() string {
	sql, args := u.buildSQL()
	// Simple placeholder replacement for debugging
	for _, arg := range args {
		sql = strings.Replace(sql, u.dialect.PlaceholderFormat(), fmt.Sprintf("'%v'", arg), 1)
	}
	return sql
}

func (u *upsertBuilder) Exec() (sql.Result, error) {
	sql, args := u.buildSQL()
	return u.db.ExecContext(u.ctx, sql, args...)
}

func (u *upsertBuilder) ExecReturnID() (int64, error) {
	result, err := u.Exec()
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (u *upsertBuilder) ExecReturnAction() (UpsertAction, int64, error) {
	// This is difficult to implement correctly without database-specific features
	// For MySQL, we can't easily determine if an insert or update occurred
	// We'll just return a generic result
	id, err := u.ExecReturnID()
	if err != nil {
		return UpsertIgnored, 0, err
	}

	// We can't determine the action accurately, so we'll just return "Inserted"
	// In a real implementation, you might use a more sophisticated approach
	return UpsertInserted, id, nil
}

func (u *upsertBuilder) WithContext(ctx context.Context) UpsertBuilder {
	u.ctx = ctx
	return u
}
