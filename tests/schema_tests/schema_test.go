package schema_tests

import (
	"context"
	"database/sql"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"

	"github.com/antibomberman/querycraft"
	"github.com/antibomberman/querycraft/dialect"
)

func TestSchemaBuilder_CreateTable(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	d := &dialect.MySQLDialect{}
	schema := querycraft.NewSchemaBuilder(sqlxDB, d)

	tableName := "users"
	expectedSQL := regexp.QuoteMeta("CREATE TABLE `users` (`id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT, `name` VARCHAR(255) NOT NULL, `email` VARCHAR(255) NOT NULL, UNIQUE KEY `users_email_unique` (`email`))")

	mock.ExpectExec(expectedSQL).WillReturnResult(sqlmock.NewResult(1, 1))

	err = schema.CreateTable(tableName, func(table querycraft.TableBuilder) {
		table.ID()
		table.String("name").NotNull()
		table.String("email").NotNull().Unique()
	})

	assert.NoError(t, err)
	mock.ExpectClose()
	assert.NoError(t, db.Close())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSchemaBuilder_AlterTable(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	d := &dialect.MySQLDialect{}
	schema := querycraft.NewSchemaBuilder(sqlxDB, d)

	tableName := "users"
	expectedSQL := "ALTER TABLE `users` ADD COLUMN `age` INT NOT NULL, DROP COLUMN `email`"

	mock.ExpectExec(expectedSQL).WillReturnResult(sqlmock.NewResult(0, 0))

	err = schema.AlterTable(tableName, func(table querycraft.TableBuilder) {
		table.Integer("age").NotNull()
		table.DropColumn("email")
	})

	assert.NoError(t, err)
	mock.ExpectClose()
	assert.NoError(t, db.Close())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSchemaBuilder_DropTable(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	d := &dialect.MySQLDialect{}
	schema := querycraft.NewSchemaBuilder(sqlxDB, d)

	tableName := "users"
	expectedSQL := "DROP TABLE `users`"

	mock.ExpectExec(expectedSQL).WillReturnResult(sqlmock.NewResult(0, 0))

	err = schema.DropTable(tableName)

	assert.NoError(t, err)
	mock.ExpectClose()
	assert.NoError(t, db.Close())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSchemaBuilder_HasTable(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	d := &dialect.MySQLDialect{}
	schema := querycraft.NewSchemaBuilder(sqlxDB, d)

	tableName := "users"
	expectedSQL := regexp.QuoteMeta("SELECT COUNT(*) > 0 FROM information_schema.tables WHERE table_schema = DATABASE() AND table_name = 'users'")

	// Test case 1: Table exists
	rows := sqlmock.NewRows([]string{"1"}).AddRow(1)
	mock.ExpectQuery(expectedSQL).WillReturnRows(rows)

	has, err := schema.HasTable(tableName)
	assert.NoError(t, err)
	assert.True(t, has)

	// Test case 2: Table does not exist
	mock.ExpectQuery(expectedSQL).WillReturnError(sql.ErrNoRows)
	has, err = schema.HasTable(tableName)
	assert.NoError(t, err)
	assert.False(t, has)

	mock.ExpectClose()
	assert.NoError(t, db.Close())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSchemaBuilder_WithContext(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)

	sqlxDB := sqlx.NewDb(db, "sqlmock")
	d := &dialect.MySQLDialect{}
	schema := querycraft.NewSchemaBuilder(sqlxDB, d)

	ctx := context.WithValue(context.Background(), "key", "value")
	schemaWithCtx := schema.WithContext(ctx)

	assert.NotNil(t, schemaWithCtx)
	// We can't directly check the context, but we can ensure it returns the correct type
	_, ok := schemaWithCtx.(querycraft.SchemaBuilder)
	assert.True(t, ok)
}
