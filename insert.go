package querycraft

import (
	"context"
	"database/sql"
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
