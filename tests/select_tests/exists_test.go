package select_tests

import (
	"testing"

	. "github.com/antibomberman/querycraft"
	"github.com/antibomberman/querycraft/dialect"
	"github.com/stretchr/testify/assert"
)

func TestWhereExists(t *testing.T) {
	mockDB := &MockSQLXExecutor{}

	// Создаем подзапрос
	subQuery := NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "1").
		From("orders").
		WhereRaw("`orders`.`user_id` = `users`.`id`")

	builder := NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")
	result := builder.From("users").WhereExists(subQuery)
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM users WHERE EXISTS (SELECT 1 FROM orders WHERE `orders`.`user_id` = `users`.`id`)"
	var expectedArgs []interface{}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestWhereNotExists(t *testing.T) {
	mockDB := &MockSQLXExecutor{}

	// Создаем подзапрос
	subQuery := NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "1").
		From("orders").
		WhereRaw("`orders`.`user_id` = `users`.`id`")

	builder := NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")
	result := builder.From("users").WhereNotExists(subQuery)
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM users WHERE NOT EXISTS (SELECT 1 FROM orders WHERE `orders`.`user_id` = `users`.`id`)"
	var expectedArgs []interface{}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}
