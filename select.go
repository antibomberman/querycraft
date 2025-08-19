package querycraft

import "context"

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
