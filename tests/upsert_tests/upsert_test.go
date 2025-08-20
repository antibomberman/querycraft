package upsert_tests

import (
	"github.com/antibomberman/querycraft/tests/test_utils"
	"testing"

	"github.com/antibomberman/querycraft"
	"github.com/antibomberman/querycraft/dialect"
	"github.com/stretchr/testify/assert"
)

type User struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Email string `db:"email"`
}

func TestUpsertValuesStruct(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewUpsertBuilder(mockDB, &dialect.MySQLDialect{}, "users")

	user := User{
		ID:    1,
		Name:  "John",
		Email: "john@example.com",
	}

	result := builder.Values(user)
	sql, args := result.ToSQL()

	// Проверяем структуру SQL, а не точный порядок столбцов
	assert.Contains(t, sql, "INSERT INTO users")
	assert.Contains(t, sql, "`id`")
	assert.Contains(t, sql, "`name`")
	assert.Contains(t, sql, "`email`")
	assert.Contains(t, sql, "VALUES")
	assert.Len(t, args, 3)
	assert.Contains(t, args, 1)
	assert.Contains(t, args, "John")
	assert.Contains(t, args, "john@example.com")
}

func TestUpsertOnConflictDoUpdate(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewUpsertBuilder(mockDB, &dialect.MySQLDialect{}, "users")

	user := User{
		ID:    1,
		Name:  "John",
		Email: "john@example.com",
	}

	result := builder.Values(user).OnConflict("id").DoUpdate("name", "email")
	sql, args := result.ToSQL()

	// Проверяем структуру SQL
	assert.Contains(t, sql, "INSERT INTO users")
	assert.Contains(t, sql, "`id`")
	assert.Contains(t, sql, "`name`")
	assert.Contains(t, sql, "`email`")
	assert.Contains(t, sql, "VALUES")
	assert.Contains(t, sql, "ON DUPLICATE KEY UPDATE")
	assert.Contains(t, sql, "`name` = VALUES(`name`)")
	assert.Contains(t, sql, "`email` = VALUES(`email`)")
	assert.Len(t, args, 3)
	assert.Contains(t, args, 1)
	assert.Contains(t, args, "John")
	assert.Contains(t, args, "john@example.com")
}

func TestUpsertColumns(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewUpsertBuilder(mockDB, &dialect.MySQLDialect{}, "users")

	// Сначала устанавливаем столбцы, затем значения
	builder.Columns("name", "email")
	// Для метода Values с массивом значений нужно использовать reflect.Value
	// или передавать map. Лучше использовать map:
	data := map[string]any{
		"name":  "John",
		"email": "john@example.com",
	}
	result := builder.Values(data)
	sql, args := result.ToSQL()

	// Проверяем структуру SQL
	assert.Contains(t, sql, "INSERT INTO users")
	assert.Contains(t, sql, "`name`")
	assert.Contains(t, sql, "`email`")
	assert.Contains(t, sql, "VALUES")
	assert.Len(t, args, 2)
	assert.Contains(t, args, "John")
	assert.Contains(t, args, "john@example.com")
}

func TestUpsertWithMap(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewUpsertBuilder(mockDB, &dialect.MySQLDialect{}, "users")

	data := map[string]any{
		"id":    1,
		"name":  "John",
		"email": "john@example.com",
	}

	result := builder.Values(data)
	sql, args := result.ToSQL()

	// Проверяем структуру SQL, а не точный порядок столбцов
	assert.Contains(t, sql, "INSERT INTO users")
	assert.Contains(t, sql, "`id`")
	assert.Contains(t, sql, "`name`")
	assert.Contains(t, sql, "`email`")
	assert.Contains(t, sql, "VALUES")
	assert.Len(t, args, 3)
	assert.Contains(t, args, 1)
	assert.Contains(t, args, "John")
	assert.Contains(t, args, "john@example.com")
}
