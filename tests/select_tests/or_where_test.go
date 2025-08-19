package select_tests

import (
	"github.com/antibomberman/querycraft/tests/test_utils"
	"testing"

	. "github.com/antibomberman/querycraft"
	"github.com/antibomberman/querycraft/dialect"
	"github.com/stretchr/testify/assert"
)

func TestOrWhere(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	result := builder.From("users").Where("status", "=", "active").OrWhere("role", "=", "admin")
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `users` WHERE `status` = ? OR `role` = ?"
	expectedArgs := []interface{}{"active", "admin"}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestOrWhereEq(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	result := builder.From("users").Where("status", "=", "active").OrWhereEq("role", "admin")
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `users` WHERE `status` = ? OR `role` = ?"
	expectedArgs := []interface{}{"active", "admin"}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestOrWhereIn(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	result := builder.From("users").Where("status", "=", "active").OrWhereIn("id", 1, 2, 3)
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `users` WHERE `status` = ? OR `id` IN (?, ?, ?)"
	expectedArgs := []interface{}{"active", 1, 2, 3}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestOrWhereNull(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	result := builder.From("users").Where("status", "=", "active").OrWhereNull("deleted_at")
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `users` WHERE `status` = ? OR `deleted_at` IS NULL"
	expectedArgs := []interface{}{"active"}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestOrWhereRaw(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	result := builder.From("users").Where("status", "=", "active").OrWhereRaw("`role` = 'admin' OR `role` = 'moderator'")
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `users` WHERE `status` = ? OR `role` = 'admin' OR `role` = 'moderator'"
	expectedArgs := []interface{}{"active"}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}
