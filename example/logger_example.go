package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/antibomberman/querycraft"
	_ "github.com/go-sql-driver/mysql"
)

func ExampleLogger() {
	// Подключение к базе данных
	db, err := sql.Open("mysql", "test_user:test_password@tcp(127.0.0.1:3336)/test_db?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Создание QueryCraft с опциями
	qc, err := querycraft.New("mysql", db, querycraft.Options{
		LogEnabled:        true,
		LogLevel:          querycraft.LogLevelInfo,
		LogFormat:         querycraft.LogFormatJSON,
		LogSaveToFile:     true,
		LogPrintToConsole: true,
		LogDir:            "./storage/logs/sql/",
		LogAutoCleanDays:  7,
	})
	if err != nil {
		log.Fatal(err)
	}

	// Очистка существующих данных
	qc.Schema().ClearTable("users")

	// Создание таблицы пользователей
	qc.Schema().CreateTable("users", func(builder querycraft.TableBuilder) {
		builder.ID()
		builder.String("name", 100).Nullable()
		builder.String("email", 255).NotNull().Unique()
		builder.Integer("age").Default(0)
		builder.Timestamp("created_at").NotNull()
		builder.Timestamp("updated_at").NotNull()
	})

	// Пример 1: Вставка пользователя с PrintSQL()
	user := struct {
		Name      string    `db:"name"`
		Email     string    `db:"email"`
		Age       int       `db:"age"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}{
		Name:      "John Doe",
		Email:     "john@example.com",
		Age:       30,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	id, err := qc.Insert("users").Values(user).PrintSQL().ExecReturnID()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Вставлен пользователь с ID: %d\n", id)

	// Пример 2: Выборка пользователей с PrintSQL()
	var users []map[string]any
	err = qc.Select("*").From("users").WhereEq("email", "john@example.com").PrintSQL().All(&users)
	if err != nil {
		log.Fatal(err)
	}

	// Пример 3: Обновление пользователя
	_, err = qc.Update("users").
		Set("name", "John Smith").
		Set("age", 31).
		WhereEq("email", "john@example.com").
		PrintSQL().
		Exec()
	if err != nil {
		log.Fatal(err)
	}
}
