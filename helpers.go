package querycraft

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"time"
)

type SQLXExecutor interface {
	// Основные методы sqlx
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)

	// Для сырых запросов
	QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error)
	QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row

	// Утилиты
	DriverName() string
	Rebind(query string) string
}

// Logger - интерфейс для логирования запросов
type Logger interface {
	LogQuery(ctx context.Context, query string, args []any, duration time.Duration, err error)
}

// Debuggable - интерфейс для отладки
type Debuggable interface {
	ToSQL() (string, []any)
	Explain() ([]map[string]any, error)
}
