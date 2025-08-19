package querycraft

import (
	"context"
	"database/sql"
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
