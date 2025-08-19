package select_tests

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// MockSQLXExecutor - мок для SQLXExecutor
type MockSQLXExecutor struct{}

func (m *MockSQLXExecutor) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return nil
}

func (m *MockSQLXExecutor) SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return nil
}

func (m *MockSQLXExecutor) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return nil, nil
}

func (m *MockSQLXExecutor) QueryxContext(ctx context.Context, query string, args ...interface{}) (*sqlx.Rows, error) {
	return nil, nil
}

func (m *MockSQLXExecutor) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	return nil
}

func (m *MockSQLXExecutor) DriverName() string {
	return "mysql"
}

func (m *MockSQLXExecutor) Rebind(query string) string {
	// Для MySQL используем ? как placeholder
	return query
}
