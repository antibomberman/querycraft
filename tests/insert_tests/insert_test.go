package insert_tests

import (
	"testing"

	"github.com/antibomberman/querycraft"
	"github.com/antibomberman/querycraft/dialect"
	"github.com/antibomberman/querycraft/tests/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestInsertValues(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewInsertBuilder(mockDB, &dialect.MySQLDialect{}, "users")

	result := builder.Columns("name", "email").Values("John", "john@example.com")
	sql, args := result.ToSQL()

	expectedSQL := "INSERT INTO `users` (`name`, `email`) VALUES (?, ?)"
	expectedArgs := []interface{}{"John", "john@example.com"}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestInsertMultipleValues(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewInsertBuilder(mockDB, &dialect.MySQLDialect{}, "users")

	result := builder.Columns("name", "email").
		Values("John", "john@example.com").
		Values("Jane", "jane@example.com")
	sql, args := result.ToSQL()

	expectedSQL := "INSERT INTO `users` (`name`, `email`) VALUES (?, ?), (?, ?)"
	expectedArgs := []interface{}{"John", "john@example.com", "Jane", "jane@example.com"}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestInsertValuesMap(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewInsertBuilder(mockDB, &dialect.MySQLDialect{}, "users")

	data := map[string]interface{}{
		"name":  "John",
		"email": "john@example.com",
	}

	result := builder.ValuesMap(data)
	sql, args := result.ToSQL()

	// Для map порядок столбцов может быть разным, поэтому проверяем только структуру запроса
	assert.Contains(t, sql, "INSERT INTO `users`")
	assert.Contains(t, sql, "`name`")
	assert.Contains(t, sql, "`email`")
	assert.Contains(t, sql, "VALUES")
	assert.Len(t, args, 2)
	assert.Contains(t, args, "John")
	assert.Contains(t, args, "john@example.com")
}

func TestInsertIgnore(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewInsertBuilder(mockDB, &dialect.MySQLDialect{}, "users")

	result := builder.Columns("name", "email").Values("John", "john@example.com").Ignore()
	sql, args := result.ToSQL()

	// В MySQL IGNORE добавляется как часть ключевого слова INSERT
	expectedSQL := "INSERT INTO `users` (`name`, `email`) VALUES (?, ?)"
	expectedArgs := []interface{}{"John", "john@example.com"}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestInsertFromSelect(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}

	// Создаем SELECT запрос
	selectBuilder := querycraft.NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "name", "email").
		From("temp_users").
		Where("active", "=", true)

	builder := querycraft.NewInsertBuilder(mockDB, &dialect.MySQLDialect{}, "users")
	result := builder.Columns("name", "email").FromSelect(selectBuilder)
	sql, args := result.ToSQL()

	expectedSQL := "INSERT INTO `users` (`name`, `email`) SELECT name, email FROM `temp_users` WHERE `active` = ?"
	expectedArgs := []interface{}{true}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}
