package select_tests

import (
	"testing"

	. "github.com/antibomberman/querycraft"
	"github.com/antibomberman/querycraft/dialect"
	"github.com/antibomberman/querycraft/tests"
	"github.com/stretchr/testify/assert"
)

func TestSelectWithReservedTableName(t *testing.T) {
	mockDB := &tests.MockSQLXExecutor{}
	builder := NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	// Тест с зарезервированным именем таблицы
	result := builder.From("order")
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `order`"
	var expectedArgs []interface{}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestSelectWithTableAlias(t *testing.T) {
	mockDB := &tests.MockSQLXExecutor{}
	builder := NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	// Тест с алиасом таблицы
	result := builder.From("order as o")
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `order` as o"
	var expectedArgs []interface{}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestSelectWithColumnAliases(t *testing.T) {
	mockDB := &tests.MockSQLXExecutor{}
	builder := NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "id as order_id", "product_name as product", "quantity as qty")

	result := builder.From("order")
	sql, args := result.ToSQL()

	expectedSQL := "SELECT id as order_id, product_name as product, quantity as qty FROM `order`"
	var expectedArgs []interface{}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestSelectWithReservedColumnNames(t *testing.T) {
	mockDB := &tests.MockSQLXExecutor{}
	builder := NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "`order` as order_number", "`group` as group_name")

	result := builder.From("order")
	sql, args := result.ToSQL()

	expectedSQL := "SELECT `order` as order_number, `group` as group_name FROM `order`"
	var expectedArgs []interface{}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestJoinWithReservedTableNames(t *testing.T) {
	mockDB := &tests.MockSQLXExecutor{}
	builder := NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	// Тест JOIN с зарезервированными именами таблиц
	result := builder.From("order as o").
		Join("user as u", "`o`.`user_id` = `u`.`id`").
		Where("status", "=", "completed")

	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `order` as o INNER JOIN `user` as u ON `o`.`user_id` = `u`.`id` WHERE `status` = ?"
	expectedArgs := []interface{}{"completed"}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestComplexQueryWithAliases(t *testing.T) {
	mockDB := &tests.MockSQLXExecutor{}
	builder := NewSelectBuilder(mockDB, &dialect.MySQLDialect{},
		"`o`.`id` as order_id",
		"`u`.`name` as customer_name",
		"`o`.`total` as order_total",
		"COUNT(`oi`.`id`) as item_count")

	result := builder.From("order as o").
		Join("user as u", "`o`.`user_id` = `u`.`id`").
		LeftJoin("order_item as oi", "`o`.`id` = `oi`.`order_id`").
		Where("created_at", ">=", "2023-01-01").
		GroupBy("`o`.`id`", "`u`.`name`", "`o`.`total`").
		OrderByDesc("created_at").
		Limit(10)

	sql, args := result.ToSQL()

	// Проверяем, что SQL содержит все необходимые части
	assert.Contains(t, sql, "SELECT `o`.`id` as order_id, `u`.`name` as customer_name, `o`.`total` as order_total, COUNT(`oi`.`id`) as item_count")
	assert.Contains(t, sql, "FROM `order` as o")
	assert.Contains(t, sql, "INNER JOIN `user` as u")
	assert.Contains(t, sql, "LEFT JOIN `order_item` as oi")
	assert.Contains(t, sql, "WHERE `created_at` >= ?")
	assert.Contains(t, sql, "GROUP BY `o`.`id`, `u`.`name`, `o`.`total`")
	assert.Contains(t, sql, "ORDER BY `created_at` DESC")
	assert.Contains(t, sql, "LIMIT 10")
	assert.Len(t, args, 1)
	assert.Equal(t, "2023-01-01", args[0])
}
