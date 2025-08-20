package querycraft

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/antibomberman/querycraft/dialect"
)

type DeleteBuilder interface {
	// WHERE условия (те же что и в SelectBuilder)
	Where(column, operator string, value any) DeleteBuilder
	WhereEq(column string, value any) DeleteBuilder
	WhereIn(column string, values ...any) DeleteBuilder
	WhereRaw(condition string, args ...any) DeleteBuilder

	// JOIN операции
	Join(table, condition string) DeleteBuilder

	// Ограничения
	Limit(limit int) DeleteBuilder
	OrderBy(column string) DeleteBuilder

	// Выполнение
	Exec() (sql.Result, error)

	// Утилиты
	WithContext(ctx context.Context) DeleteBuilder
	ToSQL() (string, []any)
	PrintSQL() DeleteBuilder
	Clone() DeleteBuilder
}

type deleteBuilder struct {
	db      SQLXExecutor
	dialect dialect.Dialect
	ctx     context.Context
	logger  Logger

	table     string
	joins     []string
	wheres    []string
	whereArgs []any
	orders    []string
	limit     *int

	// Print SQL flag
	printSQL bool
}

func NewDeleteBuilder(db SQLXExecutor, dialect dialect.Dialect, table string) DeleteBuilder {
	return &deleteBuilder{
		db:      db,
		dialect: dialect,
		ctx:     context.Background(),
		table:   table,
	}
}

func (d *deleteBuilder) Where(column, operator string, value any) DeleteBuilder {
	d.wheres = append(d.wheres, fmt.Sprintf("%s %s %s", d.dialect.QuoteIdentifier(column), operator, d.dialect.PlaceholderFormat()))
	d.whereArgs = append(d.whereArgs, value)
	return d
}

func (d *deleteBuilder) WhereEq(column string, value any) DeleteBuilder {
	return d.Where(column, "=", value)
}

func (d *deleteBuilder) WhereIn(column string, values ...any) DeleteBuilder {
	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = d.dialect.PlaceholderFormat()
	}
	d.wheres = append(d.wheres, fmt.Sprintf("%s IN (%s)", d.dialect.QuoteIdentifier(column), strings.Join(placeholders, ", ")))
	d.whereArgs = append(d.whereArgs, values...)
	return d
}

func (d *deleteBuilder) WhereRaw(condition string, args ...any) DeleteBuilder {
	d.wheres = append(d.wheres, condition)
	d.whereArgs = append(d.whereArgs, args...)
	return d
}

func (d *deleteBuilder) Join(table, condition string) DeleteBuilder {
	d.joins = append(d.joins, fmt.Sprintf("JOIN %s ON %s", table, condition))
	return d
}

func (d *deleteBuilder) Limit(limit int) DeleteBuilder {
	d.limit = &limit
	return d
}

func (d *deleteBuilder) OrderBy(column string) DeleteBuilder {
	d.orders = append(d.orders, d.dialect.SelectOrderBy(column, false))
	return d
}

func (d *deleteBuilder) buildSQL() (string, []any) {
	var queryParts []string
	var args []any

	// DELETE
	queryParts = append(queryParts, "DELETE FROM", d.table)

	// JOIN
	if len(d.joins) > 0 {
		queryParts = append(queryParts, strings.Join(d.joins, " "))
	}

	// WHERE
	if len(d.wheres) > 0 {
		queryParts = append(queryParts, "WHERE", strings.Join(d.wheres, " "))
		args = append(args, d.whereArgs...)
	}

	// ORDER BY
	if len(d.orders) > 0 {
		queryParts = append(queryParts, strings.Join(d.orders, " "))
	}

	// LIMIT
	if d.limit != nil {
		queryParts = append(queryParts, d.dialect.DeleteLimit(*d.limit))
	}

	return strings.Join(queryParts, " "), args
}

func (d *deleteBuilder) ToSQL() (string, []any) {
	return d.buildSQL()
}

func (d *deleteBuilder) PrintSQL() DeleteBuilder {
	d.printSQL = true
	return d
}

func (d *deleteBuilder) setLogger(logger Logger) {
	d.logger = logger
}

func (d *deleteBuilder) Exec() (sql.Result, error) {
	sql, args := d.buildSQL()

	// Print SQL if needed
	if d.printSQL {
		// Simple placeholder replacement for debugging
		formattedSQL := sql
		for _, arg := range args {
			formattedSQL = strings.Replace(formattedSQL, d.dialect.PlaceholderFormat(), fmt.Sprintf("'%v'", arg), 1)
		}
		fmt.Println(formattedSQL)
	}

	// Log query if logger is set
	var start time.Time
	if d.logger != nil {
		start = time.Now()
	}

	result, err := d.db.ExecContext(d.ctx, sql, args...)

	// Log query execution
	if d.logger != nil {
		duration := time.Since(start)
		d.logger.LogQuery(d.ctx, sql, args, duration, err)
	}

	return result, err
}

func (d *deleteBuilder) WithContext(ctx context.Context) DeleteBuilder {
	d.ctx = ctx
	return d
}

func (d *deleteBuilder) Clone() DeleteBuilder {
	// Create a deep copy
	clone := &deleteBuilder{
		db:      d.db,
		dialect: d.dialect,
		ctx:     d.ctx,
		table:   d.table,
		joins:   make([]string, len(d.joins)),
		wheres:  make([]string, len(d.wheres)),
		orders:  make([]string, len(d.orders)),
		limit:   d.limit,
	}

	copy(clone.joins, d.joins)
	copy(clone.wheres, d.wheres)
	copy(clone.orders, d.orders)

	// Copy args slices
	clone.whereArgs = make([]any, len(d.whereArgs))
	copy(clone.whereArgs, d.whereArgs)

	return clone
}
