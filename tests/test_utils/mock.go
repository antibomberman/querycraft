package test_utils

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// MockSQLXExecutor - мок для SQLXExecutor
type MockSQLXExecutor struct {
	// Хранилище для мок данных
	tables map[string][]map[string]interface{}
}

func NewMockSQLXExecutor() *MockSQLXExecutor {
	return &MockSQLXExecutor{
		tables: make(map[string][]map[string]interface{}),
	}
}

func (m *MockSQLXExecutor) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	// Для простоты просто возвращаем nil
	return nil
}

func (m *MockSQLXExecutor) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	// Для простоты просто возвращаем nil
	return nil
}

func (m *MockSQLXExecutor) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	// Для простоты просто возвращаем nil
	return nil, nil
}

func (m *MockSQLXExecutor) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	// Создаем пустой результат для Query
	rows := &sql.Rows{}
	return rows, nil
}

func (m *MockSQLXExecutor) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	// Для простоты просто возвращаем nil
	return nil, nil
}

func (m *MockSQLXExecutor) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
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
