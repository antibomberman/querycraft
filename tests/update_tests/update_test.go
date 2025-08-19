package update_tests

import (
	"testing"

	"github.com/antibomberman/querycraft"
	"github.com/antibomberman/querycraft/dialect"
	"github.com/antibomberman/querycraft/tests"
	"github.com/stretchr/testify/assert"
)

func TestUpdateSet(t *testing.T) {
	mockDB := &tests.MockSQLXExecutor{}
	builder := querycraft.NewUpdateBuilder(mockDB, &dialect.MySQLDialect{}, "users")

	result := builder.Set("name", "John").Set("email", "john@example.com").Where("id", "=", 1)
	sql, args := result.ToSQL()

	expectedSQL := "UPDATE users SET `name` = ?, `email` = ? WHERE `id` = ?"
	expectedArgs := []interface{}{"John", "john@example.com", 1}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestUpdateSetMap(t *testing.T) {
	mockDB := &tests.MockSQLXExecutor{}
	builder := querycraft.NewUpdateBuilder(mockDB, &dialect.MySQLDialect{}, "users")

	data := map[string]interface{}{
		"name":  "John",
		"email": "john@example.com",
	}

	result := builder.SetMap(data).Where("id", "=", 1)
	sql, args := result.ToSQL()

	// Для map порядок столбцов может быть разным, поэтому проверяем только структуру запроса
	assert.Contains(t, sql, "UPDATE users SET")
	assert.Contains(t, sql, "`name` = ?")
	assert.Contains(t, sql, "`email` = ?")
	assert.Contains(t, sql, "WHERE `id` = ?")
	assert.Len(t, args, 3)
	assert.Contains(t, args, "John")
	assert.Contains(t, args, "john@example.com")
	assert.Contains(t, args, 1)
}

func TestUpdateSetRaw(t *testing.T) {
	mockDB := &tests.MockSQLXExecutor{}
	builder := querycraft.NewUpdateBuilder(mockDB, &dialect.MySQLDialect{}, "users")

	result := builder.SetRaw("`updated_at` = NOW()").Where("id", "=", 1)
	sql, args := result.ToSQL()

	expectedSQL := "UPDATE users SET `updated_at` = NOW() WHERE `id` = ?"
	expectedArgs := []interface{}{1}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestUpdateIncrement(t *testing.T) {
	mockDB := &tests.MockSQLXExecutor{}
	builder := querycraft.NewUpdateBuilder(mockDB, &dialect.MySQLDialect{}, "users")

	result := builder.Increment("login_count", 1).Where("id", "=", 1)
	sql, args := result.ToSQL()

	expectedSQL := "UPDATE users SET `login_count` = `login_count` + ? WHERE `id` = ?"
	expectedArgs := []interface{}{1, 1}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestUpdateDecrement(t *testing.T) {
	mockDB := &tests.MockSQLXExecutor{}
	builder := querycraft.NewUpdateBuilder(mockDB, &dialect.MySQLDialect{}, "users")

	result := builder.Decrement("balance", 100).Where("id", "=", 1)
	sql, args := result.ToSQL()

	expectedSQL := "UPDATE users SET `balance` = `balance` - ? WHERE `id` = ?"
	expectedArgs := []interface{}{100, 1}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestUpdateWhereIn(t *testing.T) {
	mockDB := &tests.MockSQLXExecutor{}
	builder := querycraft.NewUpdateBuilder(mockDB, &dialect.MySQLDialect{}, "users")

	result := builder.Set("status", "inactive").WhereIn("id", 1, 2, 3)
	sql, args := result.ToSQL()

	expectedSQL := "UPDATE users SET `status` = ? WHERE `id` IN (?, ?, ?)"
	expectedArgs := []interface{}{"inactive", 1, 2, 3}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestUpdateJoin(t *testing.T) {
	mockDB := &tests.MockSQLXExecutor{}
	builder := querycraft.NewUpdateBuilder(mockDB, &dialect.MySQLDialect{}, "users")

	result := builder.Join("orders", "`users`.`id` = `orders`.`user_id`").
		Set("users.status", "premium").
		Where("orders.total", ">", 1000)
	sql, args := result.ToSQL()

	expectedSQL := "UPDATE users JOIN orders ON `users`.`id` = `orders`.`user_id` SET `users.status` = ? WHERE `orders.total` > ?"
	expectedArgs := []interface{}{"premium", 1000}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}
