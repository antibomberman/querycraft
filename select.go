package querycraft

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/antibomberman/querycraft/dialect"
	"strings"
)

// SelectBuilder - интерфейс для SELECT запросов
type SelectBuilder interface {
	// Основные методы
	From(table string) SelectBuilder

	// WHERE условия
	Where(column, operator string, value any) SelectBuilder
	WhereEq(column string, value any) SelectBuilder
	WhereIn(column string, values ...any) SelectBuilder
	WhereNotIn(column string, values ...any) SelectBuilder
	WhereNull(column string) SelectBuilder
	WhereNotNull(column string) SelectBuilder
	WhereBetween(column string, from, to any) SelectBuilder
	WhereNotBetween(column string, from, to any) SelectBuilder
	WhereRaw(condition string, args ...any) SelectBuilder

	WhereExists(subquery SelectBuilder) SelectBuilder
	WhereNotExists(subquery SelectBuilder) SelectBuilder

	// OR WHERE условия
	OrWhere(column, operator string, value any) SelectBuilder
	OrWhereEq(column string, value any) SelectBuilder
	OrWhereIn(column string, values ...any) SelectBuilder
	OrWhereNull(column string) SelectBuilder
	OrWhereRaw(condition string, args ...any) SelectBuilder

	// WHERE группировка
	WhereGroup(fn func(SelectBuilder) SelectBuilder) SelectBuilder
	OrWhereGroup(fn func(SelectBuilder) SelectBuilder) SelectBuilder

	// Условное добавление WHERE
	When(condition bool, column, operator string, value any) SelectBuilder
	WhenFunc(condition bool, fn func(SelectBuilder) SelectBuilder) SelectBuilder

	// JOIN операции
	Join(table, condition string) SelectBuilder
	InnerJoin(table, condition string) SelectBuilder
	LeftJoin(table, condition string) SelectBuilder
	RightJoin(table, condition string) SelectBuilder
	CrossJoin(table string) SelectBuilder
	OuterJoin(table, condition string) SelectBuilder

	// Сортировка и группировка
	OrderBy(column string) SelectBuilder
	OrderByDesc(column string) SelectBuilder
	OrderByRaw(expression string) SelectBuilder
	GroupBy(columns ...string) SelectBuilder
	Having(condition string, args ...any) SelectBuilder

	// Пагинация
	Limit(limit int) SelectBuilder
	Offset(offset int) SelectBuilder
	Page(page, perPage int) SelectBuilder

	// Выполнение запросов
	One(dest any) error
	All(dest any) error

	// Получение данных в разных форматах
	Row() (map[string]any, error)
	Rows() ([]map[string]any, error)
	RowsMapKey(keyColumn string) (map[any]map[string]any, error)

	// Получение отдельных значений
	Field(column string) (any, error)
	Pluck(column string) ([]any, error)

	// Агрегатные функции
	Count() (int64, error)
	CountColumn(column string) (int64, error)
	Sum(column string) (float64, error)
	Avg(column string) (float64, error)
	Max(column string) (any, error)
	Min(column string) (any, error)
	Exists() (bool, error)

	// Утилиты
	WithContext(ctx context.Context) SelectBuilder
	Clone() SelectBuilder

	ToSQL() (string, []any)
	Debug() string
	Explain() ([]map[string]any, error)
}

type FullTextBuilder interface {
	// PostgreSQL Full-Text Search
	Match(column, query string) SelectBuilder       // @@
	ToTSVector(column, config string) SelectBuilder // to_tsvector
	ToTSQuery(query, config string) SelectBuilder   // to_tsquery
	Rank(vector, query string) SelectBuilder        // ts_rank
	// MySQL Full-Text Search
	MatchAgainst(columns []string, query string, mode string) SelectBuilder
	// Generic text search
	Search(columns []string, term string) SelectBuilder
}

type OptimizationBuilder interface {
	// Hints
	UseIndex(indexes ...string) SelectBuilder
	ForceIndex(index string) SelectBuilder
	IgnoreIndex(indexes ...string) SelectBuilder

	// Locking
	ForUpdate() SelectBuilder
	ForShare() SelectBuilder
	LockInShareMode() SelectBuilder

	// Partitioning
	Partition(partitions ...string) SelectBuilder
}

type selectBuilder struct {
	db      SQLXExecutor
	dialect dialect.Dialect
	ctx     context.Context

	// Query parts
	columns    []string
	table      string
	joins      []string
	wheres     []string
	whereArgs  []any
	orders     []string
	groups     []string
	havings    []string
	havingArgs []any
	limit      *int
	offset     *int

	// For subqueries in where exists
	subqueries   []string
	subqueryArgs []any
}

func NewSelectBuilder(db SQLXExecutor, dialect dialect.Dialect, columns ...string) SelectBuilder {
	return &selectBuilder{
		db:      db,
		dialect: dialect,
		ctx:     context.Background(),
		columns: columns,
	}
}

func (s *selectBuilder) From(table string) SelectBuilder {
	s.table = table
	return s
}

func (s *selectBuilder) Where(column, operator string, value any) SelectBuilder {
	s.wheres = append(s.wheres, fmt.Sprintf("%s %s %s", s.dialect.QuoteIdentifier(column), operator, s.dialect.PlaceholderFormat()))
	s.whereArgs = append(s.whereArgs, value)
	return s
}

func (s *selectBuilder) WhereEq(column string, value any) SelectBuilder {
	return s.Where(column, "=", value)
}

func (s *selectBuilder) WhereIn(column string, values ...any) SelectBuilder {
	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = s.dialect.PlaceholderFormat()
	}
	s.wheres = append(s.wheres, fmt.Sprintf("%s IN (%s)", s.dialect.QuoteIdentifier(column), strings.Join(placeholders, ", ")))
	s.whereArgs = append(s.whereArgs, values...)
	return s
}

func (s *selectBuilder) WhereNotIn(column string, values ...any) SelectBuilder {
	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = s.dialect.PlaceholderFormat()
	}
	s.wheres = append(s.wheres, fmt.Sprintf("%s NOT IN (%s)", s.dialect.QuoteIdentifier(column), strings.Join(placeholders, ", ")))
	s.whereArgs = append(s.whereArgs, values...)
	return s
}

func (s *selectBuilder) WhereNull(column string) SelectBuilder {
	s.wheres = append(s.wheres, fmt.Sprintf("%s IS NULL", s.dialect.QuoteIdentifier(column)))
	return s
}

func (s *selectBuilder) WhereNotNull(column string) SelectBuilder {
	s.wheres = append(s.wheres, fmt.Sprintf("%s IS NOT NULL", s.dialect.QuoteIdentifier(column)))
	return s
}

func (s *selectBuilder) WhereBetween(column string, from, to any) SelectBuilder {
	s.wheres = append(s.wheres, fmt.Sprintf("%s BETWEEN %s AND %s",
		s.dialect.QuoteIdentifier(column),
		s.dialect.PlaceholderFormat(),
		s.dialect.PlaceholderFormat()))
	s.whereArgs = append(s.whereArgs, from, to)
	return s
}

func (s *selectBuilder) WhereNotBetween(column string, from, to any) SelectBuilder {
	s.wheres = append(s.wheres, fmt.Sprintf("%s NOT BETWEEN %s AND %s",
		s.dialect.QuoteIdentifier(column),
		s.dialect.PlaceholderFormat(),
		s.dialect.PlaceholderFormat()))
	s.whereArgs = append(s.whereArgs, from, to)
	return s
}

func (s *selectBuilder) WhereRaw(condition string, args ...any) SelectBuilder {
	s.wheres = append(s.wheres, condition)
	s.whereArgs = append(s.whereArgs, args...)
	return s
}

func (s *selectBuilder) WhereExists(subquery SelectBuilder) SelectBuilder {
	// For simplicity in this implementation, we'll just add a placeholder
	// A full implementation would need to handle the subquery properly
	sql, args := subquery.ToSQL()
	s.wheres = append(s.wheres, fmt.Sprintf("EXISTS (%s)", sql))
	s.whereArgs = append(s.whereArgs, args...)
	return s
}

func (s *selectBuilder) WhereNotExists(subquery SelectBuilder) SelectBuilder {
	sql, args := subquery.ToSQL()
	s.wheres = append(s.wheres, fmt.Sprintf("NOT EXISTS (%s)", sql))
	s.whereArgs = append(s.whereArgs, args...)
	return s
}

func (s *selectBuilder) OrWhere(column, operator string, value any) SelectBuilder {
	s.wheres = append(s.wheres, fmt.Sprintf("OR %s %s %s", s.dialect.QuoteIdentifier(column), operator, s.dialect.PlaceholderFormat()))
	s.whereArgs = append(s.whereArgs, value)
	return s
}

func (s *selectBuilder) OrWhereEq(column string, value any) SelectBuilder {
	return s.OrWhere(column, "=", value)
}

func (s *selectBuilder) OrWhereIn(column string, values ...any) SelectBuilder {
	placeholders := make([]string, len(values))
	for i := range values {
		placeholders[i] = s.dialect.PlaceholderFormat()
	}
	s.wheres = append(s.wheres, fmt.Sprintf("OR %s IN (%s)", s.dialect.QuoteIdentifier(column), strings.Join(placeholders, ", ")))
	s.whereArgs = append(s.whereArgs, values...)
	return s
}

func (s *selectBuilder) OrWhereNull(column string) SelectBuilder {
	s.wheres = append(s.wheres, fmt.Sprintf("OR %s IS NULL", s.dialect.QuoteIdentifier(column)))
	return s
}

func (s *selectBuilder) OrWhereRaw(condition string, args ...any) SelectBuilder {
	s.wheres = append(s.wheres, "OR "+condition)
	s.whereArgs = append(s.whereArgs, args...)
	return s
}

func (s *selectBuilder) WhereGroup(fn func(SelectBuilder) SelectBuilder) SelectBuilder {
	// Создаем новый билдер с теми же параметрами, но без columns
	groupBuilder := &selectBuilder{
		db:        s.db,
		dialect:   s.dialect,
		ctx:       s.ctx,
		wheres:    make([]string, 0),
		whereArgs: make([]any, 0),
	}

	// Применяем функцию к групповому билдеру
	builder := fn(groupBuilder)

	// Получаем условия из группового билдера напрямую
	if sb, ok := builder.(*selectBuilder); ok && len(sb.wheres) > 0 {
		// Объединяем условия с правильными AND между ними
		var whereParts []string
		for i, where := range sb.wheres {
			if i == 0 {
				whereParts = append(whereParts, where)
			} else {
				// Добавляем AND, если условие не начинается с AND, OR или (
				if strings.HasPrefix(where, "AND ") || strings.HasPrefix(where, "OR ") || strings.HasPrefix(where, "(") {
					whereParts = append(whereParts, where)
				} else {
					whereParts = append(whereParts, "AND "+where)
				}
			}
		}

		// Добавляем скобки с префиксом AND
		s.wheres = append(s.wheres, fmt.Sprintf("AND (%s)", strings.Join(whereParts, " ")))
		s.whereArgs = append(s.whereArgs, sb.whereArgs...)
	}

	return s
}

func (s *selectBuilder) OrWhereGroup(fn func(SelectBuilder) SelectBuilder) SelectBuilder {
	// Создаем новый билдер с теми же параметрами, но без columns
	groupBuilder := &selectBuilder{
		db:        s.db,
		dialect:   s.dialect,
		ctx:       s.ctx,
		wheres:    make([]string, 0),
		whereArgs: make([]any, 0),
	}

	// Применяем функцию к групповому билдеру
	builder := fn(groupBuilder)

	// Получаем условия из группового билдера напрямую
	if sb, ok := builder.(*selectBuilder); ok && len(sb.wheres) > 0 {
		// Объединяем условия с правильными AND между ними
		var whereParts []string
		for i, where := range sb.wheres {
			if i == 0 {
				whereParts = append(whereParts, where)
			} else {
				// Добавляем AND, если условие не начинается с AND, OR или (
				if strings.HasPrefix(where, "AND ") || strings.HasPrefix(where, "OR ") || strings.HasPrefix(where, "(") {
					whereParts = append(whereParts, where)
				} else {
					whereParts = append(whereParts, "AND "+where)
				}
			}
		}

		// Добавляем скобки с OR
		s.wheres = append(s.wheres, fmt.Sprintf("OR (%s)", strings.Join(whereParts, " ")))
		s.whereArgs = append(s.whereArgs, sb.whereArgs...)
	}

	return s
}

func (s *selectBuilder) When(condition bool, column, operator string, value any) SelectBuilder {
	if condition {
		return s.Where(column, operator, value)
	}
	return s
}

func (s *selectBuilder) WhenFunc(condition bool, fn func(SelectBuilder) SelectBuilder) SelectBuilder {
	if condition {
		// Сохраняем текущее состояние where
		originalWheres := make([]string, len(s.wheres))
		originalWhereArgs := make([]any, len(s.whereArgs))
		copy(originalWheres, s.wheres)
		copy(originalWhereArgs, s.whereArgs)

		// Применяем функцию
		result := fn(s)

		// Если были добавлены новые условия, добавляем AND между старыми и новыми
		if len(s.wheres) > len(originalWheres) {
			// Добавляем AND перед первым новым условием
			if len(originalWheres) > 0 {
				// Изменяем первое новое условие, добавляя AND
				newCondition := s.wheres[len(originalWheres)]
				if !strings.HasPrefix(newCondition, "OR ") && !strings.HasPrefix(newCondition, "AND ") {
					s.wheres[len(originalWheres)] = "AND " + newCondition
				}
			}
		}

		return result
	}
	return s
}

func (s *selectBuilder) Join(table, condition string) SelectBuilder {
	return s.InnerJoin(table, condition)
}

func (s *selectBuilder) InnerJoin(table, condition string) SelectBuilder {
	s.joins = append(s.joins, fmt.Sprintf("INNER JOIN %s ON %s", table, condition))
	return s
}

func (s *selectBuilder) LeftJoin(table, condition string) SelectBuilder {
	s.joins = append(s.joins, fmt.Sprintf("LEFT JOIN %s ON %s", table, condition))
	return s
}

func (s *selectBuilder) RightJoin(table, condition string) SelectBuilder {
	s.joins = append(s.joins, fmt.Sprintf("RIGHT JOIN %s ON %s", table, condition))
	return s
}

func (s *selectBuilder) CrossJoin(table string) SelectBuilder {
	s.joins = append(s.joins, fmt.Sprintf("CROSS JOIN %s", table))
	return s
}

func (s *selectBuilder) OuterJoin(table, condition string) SelectBuilder {
	s.joins = append(s.joins, fmt.Sprintf("OUTER JOIN %s ON %s", table, condition))
	return s
}

func (s *selectBuilder) OrderBy(column string) SelectBuilder {
	s.orders = append(s.orders, s.dialect.SelectOrderBy(column, false))
	return s
}

func (s *selectBuilder) OrderByDesc(column string) SelectBuilder {
	s.orders = append(s.orders, s.dialect.SelectOrderBy(column, true))
	return s
}

func (s *selectBuilder) OrderByRaw(expression string) SelectBuilder {
	s.orders = append(s.orders, expression)
	return s
}

func (s *selectBuilder) GroupBy(columns ...string) SelectBuilder {
	s.groups = append(s.groups, columns...)
	return s
}

func (s *selectBuilder) Having(condition string, args ...any) SelectBuilder {
	s.havings = append(s.havings, condition)
	s.havingArgs = append(s.havingArgs, args...)
	return s
}

func (s *selectBuilder) Limit(limit int) SelectBuilder {
	s.limit = &limit
	return s
}

func (s *selectBuilder) Offset(offset int) SelectBuilder {
	s.offset = &offset
	return s
}

func (s *selectBuilder) Page(page, perPage int) SelectBuilder {
	offset := (page - 1) * perPage
	return s.Limit(perPage).Offset(offset)
}

func (s *selectBuilder) WithContext(ctx context.Context) SelectBuilder {
	s.ctx = ctx
	return s
}

func (s *selectBuilder) Clone() SelectBuilder {
	// Create a deep copy
	clone := &selectBuilder{
		db:      s.db,
		dialect: s.dialect,
		ctx:     s.ctx,
		columns: make([]string, len(s.columns)),
		table:   s.table,
		joins:   make([]string, len(s.joins)),
		wheres:  make([]string, len(s.wheres)),
		orders:  make([]string, len(s.orders)),
		groups:  make([]string, len(s.groups)),
		havings: make([]string, len(s.havings)),
		limit:   s.limit,
		offset:  s.offset,
	}

	copy(clone.columns, s.columns)
	copy(clone.joins, s.joins)
	copy(clone.wheres, s.wheres)
	copy(clone.orders, s.orders)
	copy(clone.groups, s.groups)
	copy(clone.havings, s.havings)

	// Copy args slices
	clone.whereArgs = make([]any, len(s.whereArgs))
	copy(clone.whereArgs, s.whereArgs)

	clone.havingArgs = make([]any, len(s.havingArgs))
	copy(clone.havingArgs, s.havingArgs)

	return clone
}

func (s *selectBuilder) buildSQL() (string, []any) {
	var queryParts []string
	var args []any

	// SELECT
	if len(s.columns) == 0 {
		queryParts = append(queryParts, "SELECT *")
	} else {
		queryParts = append(queryParts, fmt.Sprintf("SELECT %s", strings.Join(s.columns, ", ")))
	}

	// FROM
	if s.table != "" {
		queryParts = append(queryParts, fmt.Sprintf("FROM %s", s.table))
	}

	// JOIN
	if len(s.joins) > 0 {
		queryParts = append(queryParts, strings.Join(s.joins, " "))
	}

	// WHERE
	if len(s.wheres) > 0 {
		// Добавляем AND между условиями, если они не начинаются с AND, OR или (
		var whereParts []string
		for i, where := range s.wheres {
			if i == 0 {
				// Первое условие не нуждается в AND/OR
				whereParts = append(whereParts, where)
			} else {
				// Для последующих условий добавляем AND, если они не начинаются с AND, OR или (
				if strings.HasPrefix(where, "AND ") || strings.HasPrefix(where, "OR ") || strings.HasPrefix(where, "(") {
					whereParts = append(whereParts, where)
				} else {
					// Особая обработка для скобок - добавляем AND перед ними
					if strings.HasPrefix(where, "(") {
						whereParts = append(whereParts, "AND "+where)
					} else {
						whereParts = append(whereParts, "AND "+where)
					}
				}
			}
		}

		queryParts = append(queryParts, fmt.Sprintf("WHERE %s", strings.Join(whereParts, " ")))
		args = append(args, s.whereArgs...)
	}

	// GROUP BY
	if len(s.groups) > 0 {
		queryParts = append(queryParts, fmt.Sprintf("GROUP BY %s", strings.Join(s.groups, ", ")))
	}

	// HAVING
	if len(s.havings) > 0 {
		queryParts = append(queryParts, fmt.Sprintf("HAVING %s", strings.Join(s.havings, " ")))
		args = append(args, s.havingArgs...)
	}

	// ORDER BY
	if len(s.orders) > 0 {
		// Для OrderByRaw выражения уже содержат "ORDER BY", для OrderBy и OrderByDesc - нет
		// Нам нужно проверить, содержат ли выражения префикс "ORDER BY"
		var orderParts []string
		for _, order := range s.orders {
			if strings.HasPrefix(strings.ToUpper(order), "ORDER BY") {
				// Это выражение от OrderByRaw, используем как есть
				orderParts = append(orderParts, order)
			} else {
				// Это выражение от OrderBy или OrderByDesc, добавляем префикс
				orderParts = append(orderParts, "ORDER BY "+order)
			}
		}
		queryParts = append(queryParts, strings.Join(orderParts, " "))
	}

	// LIMIT
	if s.limit != nil {
		queryParts = append(queryParts, s.dialect.SelectLimit(*s.limit))
	}

	// OFFSET
	if s.offset != nil {
		queryParts = append(queryParts, s.dialect.SelectOffset(*s.offset))
	}

	return strings.Join(queryParts, " "), args
}

//Exec Methods

func (s *selectBuilder) ToSQL() (string, []any) {
	return s.buildSQL()
}

func (s *selectBuilder) Debug() string {
	sql, args := s.buildSQL()
	// Simple placeholder replacement for debugging
	for _, arg := range args {
		sql = strings.Replace(sql, s.dialect.PlaceholderFormat(), fmt.Sprintf("'%v'", arg), 1)
	}
	return sql
}

func (s *selectBuilder) Explain() ([]map[string]any, error) {
	sql, args := s.buildSQL()
	explainSQL := fmt.Sprintf("EXPLAIN %s", sql)

	rows, err := s.db.QueryxContext(s.ctx, explainSQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]any
	for rows.Next() {
		row := make(map[string]any)
		if err := rows.MapScan(row); err != nil {
			return nil, err
		}
		results = append(results, row)
	}

	return results, nil
}

func (s *selectBuilder) One(dest any) error {
	sql, args := s.buildSQL()
	return s.db.GetContext(s.ctx, dest, sql, args...)
}

func (s *selectBuilder) All(dest any) error {
	sql, args := s.buildSQL()
	return s.db.SelectContext(s.ctx, dest, sql, args...)
}

func (s *selectBuilder) Row() (map[string]any, error) {
	query, args := s.buildSQL()

	rows, err := s.db.QueryxContext(s.ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		row := make(map[string]any)
		if err := rows.MapScan(row); err != nil {
			return nil, err
		}
		return row, nil
	}

	return nil, sql.ErrNoRows
}

func (s *selectBuilder) Rows() ([]map[string]any, error) {
	sql, args := s.buildSQL()

	rows, err := s.db.QueryxContext(s.ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]any
	for rows.Next() {
		row := make(map[string]any)
		if err := rows.MapScan(row); err != nil {
			return nil, err
		}
		results = append(results, row)
	}

	return results, nil
}

func (s *selectBuilder) RowsMapKey(keyColumn string) (map[any]map[string]any, error) {
	sql, args := s.buildSQL()

	rows, err := s.db.QueryxContext(s.ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	results := make(map[any]map[string]any)
	for rows.Next() {
		row := make(map[string]any)
		if err := rows.MapScan(row); err != nil {
			return nil, err
		}

		key, ok := row[keyColumn]
		if !ok {
			return nil, fmt.Errorf("key column %s not found in result", keyColumn)
		}

		results[key] = row
	}

	return results, nil
}

func (s *selectBuilder) Field(column string) (any, error) {
	originalColumns := s.columns
	s.columns = []string{column}

	defer func() {
		s.columns = originalColumns
	}()

	row, err := s.Row()
	if err != nil {
		return nil, err
	}

	value, ok := row[column]
	if !ok {
		return nil, fmt.Errorf("column %s not found in result", column)
	}

	return value, nil
}

func (s *selectBuilder) Pluck(column string) ([]any, error) {
	originalColumns := s.columns
	s.columns = []string{column}

	defer func() {
		s.columns = originalColumns
	}()

	rows, err := s.Rows()
	if err != nil {
		return nil, err
	}

	var results []any
	for _, row := range rows {
		value, ok := row[column]
		if !ok {
			return nil, fmt.Errorf("column %s not found in result", column)
		}
		results = append(results, value)
	}

	return results, nil
}

func (s *selectBuilder) Count() (int64, error) {
	return s.CountColumn("*")
}

func (s *selectBuilder) CountColumn(column string) (int64, error) {
	originalColumns := s.columns
	s.columns = []string{fmt.Sprintf("COUNT(%s) as count", column)}

	defer func() {
		s.columns = originalColumns
	}()

	var result struct {
		Count int64 `db:"count"`
	}

	err := s.One(&result)
	if err != nil {
		return 0, err
	}

	return result.Count, nil
}

func (s *selectBuilder) Sum(column string) (float64, error) {
	originalColumns := s.columns
	s.columns = []string{fmt.Sprintf("SUM(%s) as sum", column)}

	defer func() {
		s.columns = originalColumns
	}()

	var result struct {
		Sum *float64 `db:"sum"`
	}

	err := s.One(&result)
	if err != nil {
		return 0, err
	}

	if result.Sum == nil {
		return 0, nil
	}

	return *result.Sum, nil
}

func (s *selectBuilder) Avg(column string) (float64, error) {
	originalColumns := s.columns
	s.columns = []string{fmt.Sprintf("AVG(%s) as avg", column)}

	defer func() {
		s.columns = originalColumns
	}()

	var result struct {
		Avg *float64 `db:"avg"`
	}

	err := s.One(&result)
	if err != nil {
		return 0, err
	}

	if result.Avg == nil {
		return 0, nil
	}

	return *result.Avg, nil
}

func (s *selectBuilder) Max(column string) (any, error) {
	originalColumns := s.columns
	s.columns = []string{fmt.Sprintf("MAX(%s) as max", column)}

	defer func() {
		s.columns = originalColumns
	}()

	var result struct {
		Max any `db:"max"`
	}

	err := s.One(&result)
	if err != nil {
		return nil, err
	}

	return result.Max, nil
}

func (s *selectBuilder) Min(column string) (any, error) {
	originalColumns := s.columns
	s.columns = []string{fmt.Sprintf("MIN(%s) as min", column)}

	defer func() {
		s.columns = originalColumns
	}()

	var result struct {
		Min any `db:"min"`
	}

	err := s.One(&result)
	if err != nil {
		return nil, err
	}

	return result.Min, nil
}

func (s *selectBuilder) Exists() (bool, error) {
	originalLimit := s.limit
	s.limit = &[]int{1}[0] // Set limit to 1

	defer func() {
		s.limit = originalLimit
	}()

	sql, args := s.buildSQL()
	checkSQL := fmt.Sprintf("SELECT EXISTS(%s) as exists", sql)

	var result struct {
		Exists bool `db:"exists"`
	}

	err := s.db.GetContext(s.ctx, &result, checkSQL, args...)
	if err != nil {
		return false, err
	}

	return result.Exists, nil
}

// Testing methods - for testing purposes only
func (s *selectBuilder) Columns() []string {
	// Create a copy of the slice to avoid external modification
	columns := make([]string, len(s.columns))
	copy(columns, s.columns)
	return columns
}

func (s *selectBuilder) SetColumns(columns ...string) {
	s.columns = columns
}
