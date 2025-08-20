package select_tests

import (
	"github.com/antibomberman/querycraft/tests/test_utils"
	"testing"

	"github.com/antibomberman/querycraft"
	"github.com/antibomberman/querycraft/dialect"
	"github.com/stretchr/testify/assert"
)

func TestWhere(t *testing.T) {
	// Создаем мок для SQLXExecutor, который будет возвращать ожидаемый SQL
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	// Тест Where
	result := builder.From("users").Where("id", "=", 1)
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `users` WHERE `id` = ?"
	expectedArgs := []any{1}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestWhereEq(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	result := builder.From("users").WhereEq("id", 1)
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `users` WHERE `id` = ?"
	expectedArgs := []any{1}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestWhereIn(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	result := builder.From("users").WhereIn("id", 1, 2, 3)
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `users` WHERE `id` IN (?, ?, ?)"
	expectedArgs := []any{1, 2, 3}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestWhereNotIn(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	result := builder.From("users").WhereNotIn("id", 1, 2, 3)
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `users` WHERE `id` NOT IN (?, ?, ?)"
	expectedArgs := []any{1, 2, 3}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestWhereNull(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	result := builder.From("users").WhereNull("name")
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `users` WHERE `name` IS NULL"
	var expectedArgs []any

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestWhereNotNull(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	result := builder.From("users").WhereNotNull("name")
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `users` WHERE `name` IS NOT NULL"
	var expectedArgs []any

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestWhereBetween(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	result := builder.From("users").WhereBetween("age", 18, 65)
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `users` WHERE `age` BETWEEN ? AND ?"
	expectedArgs := []any{18, 65}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestWhereNotBetween(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	result := builder.From("users").WhereNotBetween("age", 18, 65)
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `users` WHERE `age` NOT BETWEEN ? AND ?"
	expectedArgs := []any{18, 65}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestWhereRaw(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	result := builder.From("users").WhereRaw("`name` = 'John' AND `age` > ?", 18)
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `users` WHERE `name` = 'John' AND `age` > ?"
	expectedArgs := []any{18}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}
