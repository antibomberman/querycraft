package select_tests

import (
	"github.com/antibomberman/querycraft/tests/test_utils"
	"testing"

	. "github.com/antibomberman/querycraft"
	"github.com/antibomberman/querycraft/dialect"
	"github.com/stretchr/testify/assert"
)

func TestSelectOrderFromUserWithJoin(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}

	// Тест для запроса qc.Select("order.*").
	//	From("user as u").
	//	Join("order", "order.user_id = u.id").
	//	Limit(500)
	builder := NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "order.*")
	result := builder.From("user as u").
		Join("order", "order.user_id = u.id").
		Limit(500)

	sql, args := result.ToSQL()

	expectedSQL := "SELECT order.* FROM `user` as u INNER JOIN `order` ON order.user_id = u.id LIMIT 500"
	var expectedArgs []any

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}

func TestSelectOrderLimitWhere(t *testing.T) {
	mockDB := &test_utils.MockSQLXExecutor{}

	// Тест для запроса qc.Select("order.limit as _lim").
	//	Where("_lim", "<>", "0")
	builder := NewSelectBuilder(mockDB, &dialect.MySQLDialect{}, "order.limit as _lim")
	result := builder.Where("_lim", "<>", "0")

	sql, args := result.ToSQL()

	expectedSQL := "SELECT order.limit as _lim WHERE `_lim` <> ?"
	expectedArgs := []any{"0"}

	assert.Equal(t, expectedSQL, sql)
	assert.Equal(t, expectedArgs, args)
}
