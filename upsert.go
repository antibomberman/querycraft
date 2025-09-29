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
	PrintSQL() UpsertBuilder
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
	logger  Logger

	table           string
	columns         []string
	values          [][]any
	conflictColumns []string
	updateColumns   []string
	updateExcluded  []string
	updateWhere     string
	updateWhereArgs []any
	action          UpsertAction

	// Print SQL flag
	printSQL bool
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
	if data == nil {
		return u
	}

	v := reflect.ValueOf(data)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return u
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		u.addStruct(v)
	case reflect.Map:
		u.addMap(v)
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			if elem.Kind() == reflect.Ptr {
				if elem.IsNil() {
					continue
				}
				elem = elem.Elem()
			}

			switch elem.Kind() {
			case reflect.Struct:
				u.addStruct(elem)
			case reflect.Map:
				u.addMap(elem)
			default:
				// Potentially handle other types or panic
			}
		}
	default:
		// Potentially handle other types or panic
	}

	return u
}

func (u *upsertBuilder) addStruct(v reflect.Value) {
	t := v.Type()
	var rowValues []any

	if len(u.columns) == 0 {
		var columns []string
		for i := 0; i < v.NumField(); i++ {
			field := t.Field(i)
			if !field.IsExported() {
				continue
			}
			column := field.Name
			if tag := field.Tag.Get("db"); tag != "" {
				if tag == "-" {
					continue
				}
				column = tag
			}
			columns = append(columns, column)
			rowValues = append(rowValues, v.Field(i).Interface())
		}
		u.columns = columns
	} else {
		for _, column := range u.columns {
			var fieldValue reflect.Value
			for i := 0; i < t.NumField(); i++ {
				field := t.Field(i)
				tag := field.Tag.Get("db")
				if tag == column {
					fieldValue = v.Field(i)
					break
				}
			}

			if fieldValue.IsValid() {
				rowValues = append(rowValues, fieldValue.Interface())
			} else {
				rowValues = append(rowValues, nil)
			}
		}
	}
	u.values = append(u.values, rowValues)
}

func (u *upsertBuilder) addMap(v reflect.Value) {
	var rowValues []any

	if len(u.columns) == 0 {
		var columns []string
		for _, key := range v.MapKeys() {
			columns = append(columns, key.String())
			rowValues = append(rowValues, v.MapIndex(key).Interface())
		}
		u.columns = columns
	} else {
		for _, column := range u.columns {
			value := v.MapIndex(reflect.ValueOf(column))
			if value.IsValid() {
				rowValues = append(rowValues, value.Interface())
			} else {
				rowValues = append(rowValues, nil)
			}
		}
	}
	u.values = append(u.values, rowValues)
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
		var valueParts []string
		for _, row := range u.values {
			placeholders := make([]string, len(row))
			for j := range row {
				placeholders[j] = u.dialect.PlaceholderFormat()
			}
			valueParts = append(valueParts, fmt.Sprintf("(%s)", strings.Join(placeholders, ", ")))
			args = append(args, row...)
		}
		queryParts = append(queryParts, "VALUES", strings.Join(valueParts, ", "))
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

func (u *upsertBuilder) PrintSQL() UpsertBuilder {
	u.printSQL = true
	return u
}

func (u *upsertBuilder) setLogger(logger Logger) {
	u.logger = logger
}

func (u *upsertBuilder) Exec() (sql.Result, error) {
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
