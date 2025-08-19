package querycraft

import (
	"context"
	"database/sql"
)

// UpdateBuilder - интерфейс для UPDATE запросов
type UpdateBuilder interface {
	// Установка значений
	Set(column string, value any) UpdateBuilder
	SetRaw(expression string, args ...any) UpdateBuilder
	SetMap(values map[string]any) UpdateBuilder
	SetStruct(data any) UpdateBuilder

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
	Debug() string
	Clone() UpdateBuilder
}
