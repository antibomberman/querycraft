package select_tests

import (
	"github.com/antibomberman/querycraft/tests/test_utils"
	"testing"

	"github.com/antibomberman/querycraft"
	"github.com/antibomberman/querycraft/dialect"
	"github.com/stretchr/testify/assert"
)

func TestOrderBy(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	result := builder.From("users").OrderBy("name")
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `users` ORDER BY `name`"
	var expectedArgs []interface{}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestOrderByDesc(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	result := builder.From("users").OrderByDesc("name")
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `users` ORDER BY `name` DESC"
	var expectedArgs []interface{}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestOrderByRaw(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "*")

	result := builder.From("users").OrderByRaw("CASE WHEN `status` = 'active' THEN 1 ELSE 2 END")
	sql, args := result.ToSQL()

	expectedSQL := "SELECT * FROM `users` ORDER BY CASE WHEN `status` = 'active' THEN 1 ELSE 2 END"
	var expectedArgs []interface{}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestGroupBy(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "category", "COUNT(*) as count")

	result := builder.From("products").GroupBy("category")
	sql, args := result.ToSQL()

	expectedSQL := "SELECT category, COUNT(*) as count FROM `products` GROUP BY category"
	var expectedArgs []interface{}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestHaving(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}
	builder := querycraft.NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "category", "COUNT(*) as count")

	result := builder.From("products").GroupBy("category").Having("COUNT(*) > ?", 5)
	sql, args := result.ToSQL()

	expectedSQL := "SELECT category, COUNT(*) as count FROM `products` GROUP BY category HAVING COUNT(*) > ?"
	expectedArgs := []interface{}{5}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}
