package select_tests

import (
	"github.com/antibomberman/querycraft/tests/test_utils"
	"testing"

	. "github.com/antibomberman/querycraft"
	"github.com/antibomberman/querycraft/dialect"
	"github.com/stretchr/testify/assert"
)

func TestWhereGroup(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	result := builder.From("users").Where("status", "=", "active").
		WhereGroup(func(b SelectBuilder) SelectBuilder {
			return b.Where("age", ">=", 18).Where("age", "<=", 65)
		})

	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `users` WHERE `status` = ? AND (`age` >= ? AND `age` <= ?)"
	expectedArgs := []any{"active", 18, 65}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestOrWhereGroup(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	result := builder.From("users").Where("status", "=", "active").
		OrWhereGroup(func(b SelectBuilder) SelectBuilder {
			return b.Where("role", "=", "admin").OrWhere("role", "=", "moderator")
		})

	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `users` WHERE `status` = ? OR (`role` = ? OR `role` = ?)"
	expectedArgs := []any{"active", "admin", "moderator"}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}
