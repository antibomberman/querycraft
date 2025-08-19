package transaction_tests

import (
	"testing"

	"github.com/antibomberman/querycraft"
	"github.com/antibomberman/querycraft/dialect"
	"github.com/stretchr/testify/assert"
)

func TestTransactionSelect(t *testing.T) {
	// Note: Для полноценного тестирования транзакций потребуется мок sqlx.Tx,
	// который выходит за рамки этих тестов.
	// Здесь мы просто проверим, что методы доступны и возвращают правильные типы.

	// Создаем мок транзакции (в реальности это будет *sqlx.Tx)
	// Для тестов мы используем тот же мок, что и для других операций
	tx := querycraft.NewTransaction(nil, nil, &dialect.MySQLDialect{})

	// Проверяем, что метод Select возвращает SelectBuilder
	builder := tx.Select("id", "name").From("users")
	assert.NotNil(t, builder)
}

func TestTransactionInsert(t *testing.T) {
	tx := querycraft.NewTransaction(nil, nil, &dialect.MySQLDialect{})

	// Проверяем, что метод Insert возвращает InsertBuilder
	builder := tx.Insert("users")
	assert.NotNil(t, builder)
}

func TestTransactionUpdate(t *testing.T) {
	tx := querycraft.NewTransaction(nil, nil, &dialect.MySQLDialect{})

	// Проверяем, что метод Update возвращает UpdateBuilder
	builder := tx.Update("users")
	assert.NotNil(t, builder)
}

func TestTransactionDelete(t *testing.T) {
	tx := querycraft.NewTransaction(nil, nil, &dialect.MySQLDialect{})

	// Проверяем, что метод Delete возвращает DeleteBuilder
	builder := tx.Delete("users")
	assert.NotNil(t, builder)
}

func TestTransactionRaw(t *testing.T) {
	tx := querycraft.NewTransaction(nil, nil, &dialect.MySQLDialect{})

	// Проверяем, что метод Raw возвращает Raw
	raw := tx.Raw("SELECT * FROM users WHERE id = ?", 1)
	assert.NotNil(t, raw)
}

func TestTransactionUpsert(t *testing.T) {
	tx := querycraft.NewTransaction(nil, nil, &dialect.MySQLDialect{})

	// Проверяем, что метод Upsert возвращает UpsertBuilder
	builder := tx.Upsert("users")
	assert.NotNil(t, builder)
}
