package querycraft

import (
	"context"
	"database/sql"
)

// RawQuery - интерфейс для сырых SQL запросов
type Raw interface {
	// Выполнение с разными результатами
	One(dest any) error
	All(dest any) error
	Row() (map[string]any, error)
	Rows() ([]map[string]any, error)
	Exec() (sql.Result, error)

	// Утилиты
	WithContext(ctx context.Context) Raw
	Args() []any
	Query() string
}
