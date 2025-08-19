package raw_tests

import (
	"testing"

	"github.com/antibomberman/querycraft"
	"github.com/antibomberman/querycraft/tests"
	"github.com/stretchr/testify/assert"
)

func TestRawQuery(t *testing.T) {
	mockDB := &tests.MockSQLXExecutor{}

	raw := querycraft.NewRaw(mockDB, "SELECT * FROM users WHERE id = ?", 1)

	// Проверяем, что запрос и аргументы сохраняются правильно
	assert.Equal(t, "SELECT * FROM users WHERE id = ?", raw.Query())
	assert.Equal(t, []interface{}{1}, raw.Args())
}

func TestRawQueryWithoutArgs(t *testing.T) {
	mockDB := &tests.MockSQLXExecutor{}

	raw := querycraft.NewRaw(mockDB, "SELECT COUNT(*) FROM users")

	// Проверяем, что запрос и аргументы сохраняются правильно
	assert.Equal(t, "SELECT COUNT(*) FROM users", raw.Query())
	// Для пустых аргументов можем получить nil или пустой срез, проверяем длину
	assert.Len(t, raw.Args(), 0)
}

func TestRawQueryWithMultipleArgs(t *testing.T) {
	mockDB := &tests.MockSQLXExecutor{}

	raw := querycraft.NewRaw(mockDB, "SELECT * FROM users WHERE age BETWEEN ? AND ? AND status = ?", 18, 65, "active")

	// Проверяем, что запрос и аргументы сохраняются правильно
	assert.Equal(t, "SELECT * FROM users WHERE age BETWEEN ? AND ? AND status = ?", raw.Query())
	assert.Equal(t, []interface{}{18, 65, "active"}, raw.Args())
}
