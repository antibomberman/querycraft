package querycraft

import (
	"context"
	"database/sql"
)

type DeleteBuilder interface {
	// WHERE условия (те же что и в SelectBuilder)
	Where(column, operator string, value any) DeleteBuilder
	WhereEq(column string, value any) DeleteBuilder
	WhereIn(column string, values ...any) DeleteBuilder
	WhereRaw(condition string, args ...any) DeleteBuilder

	// JOIN для DELETE
	Join(table, condition string) DeleteBuilder
	LeftJoin(table, condition string) DeleteBuilder

	// Ограничения
	Limit(limit int) DeleteBuilder
	OrderBy(column string) DeleteBuilder

	// Выполнение
	Exec() (sql.Result, error)

	// Утилиты
	WithContext(ctx context.Context) DeleteBuilder
	ToSQL() (string, []any)
	Debug() string
	Clone() DeleteBuilder
}
