package select_tests

import (
	"testing"

	"github.com/antibomberman/querycraft"
	"github.com/antibomberman/querycraft/dialect"
	"github.com/antibomberman/querycraft/tests"
	"github.com/stretchr/testify/assert"
)

func TestJoin(t *testing.T) {
	mockDB := &tests.MockSQLXExecutor{}
	builder := querycraft.NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	// Тест Join (INNER JOIN)
	result := builder.From("users").Join("orders", "`users`.`id` = `orders`.`user_id`")
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `users` INNER JOIN `orders` ON `users`.`id` = `orders`.`user_id`"
	var expectedArgs []interface{}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestInnerJoin(t *testing.T) {
	mockDB := &tests.MockSQLXExecutor{}
	builder := querycraft.NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	result := builder.From("users").InnerJoin("orders", "`users`.`id` = `orders`.`user_id`")
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `users` INNER JOIN `orders` ON `users`.`id` = `orders`.`user_id`"
	var expectedArgs []interface{}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestLeftJoin(t *testing.T) {
	mockDB := &tests.MockSQLXExecutor{}
	builder := querycraft.NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	result := builder.From("users").LeftJoin("orders", "`users`.`id` = `orders`.`user_id`")
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `users` LEFT JOIN `orders` ON `users`.`id` = `orders`.`user_id`"
	var expectedArgs []interface{}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestRightJoin(t *testing.T) {
	mockDB := &tests.MockSQLXExecutor{}
	builder := querycraft.NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	result := builder.From("users").RightJoin("orders", "`users`.`id` = `orders`.`user_id`")
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `users` RIGHT JOIN `orders` ON `users`.`id` = `orders`.`user_id`"
	var expectedArgs []interface{}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestCrossJoin(t *testing.T) {
	mockDB := &tests.MockSQLXExecutor{}
	builder := querycraft.NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	result := builder.From("users").CrossJoin("orders")
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `users` CROSS JOIN `orders`"
	var expectedArgs []interface{}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestOuterJoin(t *testing.T) {
	mockDB := &tests.MockSQLXExecutor{}
	builder := querycraft.NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	result := builder.From("users").OuterJoin("orders", "`users`.`id` = `orders`.`user_id`")
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `users` OUTER JOIN `orders` ON `users`.`id` = `orders`.`user_id`"
	var expectedArgs []interface{}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}
