package delete_tests

import (
	"testing"

	"github.com/antibomberman/querycraft"
	"github.com/antibomberman/querycraft/dialect"
	"github.com/antibomberman/querycraft/tests/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestDeleteWhere(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewDeleteBuilder(mockDB, &dialect.MySQLDialect{}, "users")

	result := builder.Where("id", "=", 1)
	sql, args := result.ToSQL()

	expectedSQL := "DELETE FROM users WHERE `id` = ?"
	expectedArgs := []interface{}{1}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestDeleteWhereEq(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewDeleteBuilder(mockDB, &dialect.MySQLDialect{}, "users")

	result := builder.WhereEq("id", 1)
	sql, args := result.ToSQL()

	expectedSQL := "DELETE FROM users WHERE `id` = ?"
	expectedArgs := []interface{}{1}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestDeleteWhereIn(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewDeleteBuilder(mockDB, &dialect.MySQLDialect{}, "users")

	result := builder.WhereIn("id", 1, 2, 3)
	sql, args := result.ToSQL()

	expectedSQL := "DELETE FROM users WHERE `id` IN (?, ?, ?)"
	expectedArgs := []interface{}{1, 2, 3}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestDeleteWhereRaw(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewDeleteBuilder(mockDB, &dialect.MySQLDialect{}, "users")

	result := builder.WhereRaw("`created_at` < NOW() - INTERVAL 30 DAY")
	sql, args := result.ToSQL()

	expectedSQL := "DELETE FROM users WHERE `created_at` < NOW() - INTERVAL 30 DAY"
	var expectedArgs []interface{}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestDeleteJoin(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewDeleteBuilder(mockDB, &dialect.MySQLDialect{}, "users")

	result := builder.Join("orders", "`users`.`id` = `orders`.`user_id`").Where("orders.status", "=", "cancelled")
	sql, args := result.ToSQL()

	expectedSQL := "DELETE FROM users JOIN orders ON `users`.`id` = `orders`.`user_id` WHERE `orders.status` = ?"
	expectedArgs := []interface{}{"cancelled"}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestDeleteWithLimit(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewDeleteBuilder(mockDB, &dialect.MySQLDialect{}, "users")

	result := builder.Where("active", "=", false).Limit(10)
	sql, args := result.ToSQL()

	expectedSQL := "DELETE FROM users WHERE `active` = ? LIMIT 10"
	expectedArgs := []interface{}{false}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestDeleteWithOrderBy(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewDeleteBuilder(mockDB, &dialect.MySQLDialect{}, "users")

	result := builder.Where("active", "=", false).OrderBy("created_at")
	sql, args := result.ToSQL()

	// Проверяем, что SQL содержит необходимые части
	assert.Contains(t, sql, "DELETE FROM users")
	assert.Contains(t, sql, "WHERE `active` = ?")
	assert.Contains(t, sql, "ORDER BY `created_at`")
	assert.Len(t, args, 1)
	assert.Equal(t, false, args[0])
}
