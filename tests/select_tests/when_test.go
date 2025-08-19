package select_tests

import (
	"testing"

	. "github.com/antibomberman/querycraft"
	"github.com/antibomberman/querycraft/dialect"
	"github.com/stretchr/testify/assert"
)

func TestWhen(t *testing.T) {
	mockDB := &MockSQLXExecutor{}

	// Тест когда условие истинно
	builder := NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")
	result := builder.From("users").When(true, "status", "=", "active")
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM users WHERE `status` = ?"
	expectedArgs := []interface{}{"active"}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)

	// Тест когда условие ложно
	builder2 := NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")
	result2 := builder2.From("users").When(false, "status", "=", "active")
	sql2, args2 := result2.ToSQL()

	expectedSQL2 := "SELECT * FROM users"
	var expectedArgs2 []interface{}

	assert.Equal(t, expectedSQL2, sql2)
	assert.Equal(t, expectedArgs2, args2)
}

func TestWhenFunc(t *testing.T) {
	mockDB := &MockSQLXExecutor{}

	// Тест когда условие истинно
	builder := NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")
	result := builder.From("users").WhenFunc(true, func(b SelectBuilder) SelectBuilder {
		return b.Where("status", "=", "active").Where("age", ">=", 18)
	})
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM users WHERE `status` = ? AND `age` >= ?"
	expectedArgs := []interface{}{"active", 18}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)

	// Тест когда условие ложно
	builder2 := NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")
	result2 := builder2.From("users").WhenFunc(false, func(b SelectBuilder) SelectBuilder {
		return b.Where("status", "=", "active").Where("age", ">=", 18)
	})
	sql2, args2 := result2.ToSQL()

	expectedSQL2 := "SELECT * FROM users"
	var expectedArgs2 []interface{}

	assert.Equal(t, expectedSQL2, sql2)
	assert.Equal(t, expectedArgs2, args2)
}
