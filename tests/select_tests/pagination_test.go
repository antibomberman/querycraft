package select_tests

import (
	"testing"

	"github.com/antibomberman/querycraft"
	"github.com/antibomberman/querycraft/dialect"
	"github.com/stretchr/testify/assert"
)

func TestLimit(t *testing.T) {
	mockDB := &MockSQLXExecutor{}
	builder := querycraft.NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	result := builder.From("users").Limit(10)
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM users LIMIT 10"
	var expectedArgs []interface{}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestOffset(t *testing.T) {
	mockDB := &MockSQLXExecutor{}
	builder := querycraft.NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	result := builder.From("users").Offset(20)
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM users OFFSET 20"
	var expectedArgs []interface{}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestPage(t *testing.T) {
	mockDB := &MockSQLXExecutor{}
	builder := querycraft.NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	// Тест Page(3, 10) - третья страница по 10 записей
	result := builder.From("users").Page(3, 10)
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM users LIMIT 10 OFFSET 20"
	var expectedArgs []interface{}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}
