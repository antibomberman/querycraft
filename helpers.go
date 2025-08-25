package querycraft

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"
	"unicode/utf8"
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

func convertByteArrayToString(data map[string]any) map[string]any {
	if data == nil {
		return nil
	}
	result := make(map[string]any)
	for key, value := range data {
		switch v := value.(type) {
		case []byte:
			if utf8.Valid(v) {
				result[key] = string(v)
			} else {
				result[key] = v
			}
		case nil:
			result[key] = nil
		default:
			result[key] = value
		}
	}
	return result
}
func formatArg(arg any) string {
	switch v := arg.(type) {
	case string:
		return "'" + strings.Replace(v, "'", "''", -1) + "'"
	case bool:
		if v {
			return "true"
		}
		return "false"
	case nil:
		return "NULL"
	default:
		return fmt.Sprintf("%v", v)
	}
}
