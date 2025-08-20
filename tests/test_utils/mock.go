package test_utils

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// MockSQLXExecutor - мок для SQLXExecutor
type MockSQLXExecutor struct {
	// Хранилище для мок данных
	tables map[string][]map[string]any
}

func NewMockSQLXExecutor() *MockSQLXExecutor {
	return &MockSQLXExecutor{
		tables: make(map[string][]map[string]any),
	}
}

func (m *MockSQLXExecutor) GetContext(ctx context.Context, dest any, query string, args ...any) error {
	// Для простоты просто возвращаем nil
	return nil
}

func (m *MockSQLXExecutor) SelectContext(ctx context.Context, dest any, query string, args ...any) error {
	// Для простоты просто возвращаем nil
	return nil
}

func (m *MockSQLXExecutor) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	// Для простоты просто возвращаем nil
	return nil, nil
}

func (m *MockSQLXExecutor) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	// Создаем пустой результат для Query
	rows := &sql.Rows{}
	return rows, nil
}

func (m *MockSQLXExecutor) QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error) {
	// Для простоты просто возвращаем nil
	return nil, nil
}

func (m *MockSQLXExecutor) QueryRowxContext(ctx context.Context, query string, args ...any) *sqlx.Row {
	// Для простоты просто возвращаем nil
	return nil
}

func (m *MockSQLXExecutor) DriverName() string {
	return "mysql"
}

func (m *MockSQLXExecutor) Rebind(query string) string {
	// Для MySQL используем ? как placeholder
	return query
}
