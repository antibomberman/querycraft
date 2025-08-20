package main

import (
	"log"
	"os"
	"time"

	"github.com/antibomberman/querycraft"
	_ "github.com/go-sql-driver/mysql"
)

func ExampleLogger() {
	// Создание директории для логов
	err := os.MkdirAll("./storage/logs/sql", 0755)
	if err != nil {
		log.Fatal(err)
	}

	// Создание логгера с опциями
	loggerOptions := querycraft.DefaultLoggerOptions()
	loggerOptions.Enabled = false
	//loggerOptions.LogDir = "./storage/logs/sql/"
	//loggerOptions.AutoCleanDays = 7

	logger := querycraft.NewFileLogger(loggerOptions)

	// Создание QueryCraft с логгером
	qc, err := querycraft.NewWithLogger("mysql", DB, logger)
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

	// Пример 1: Вставка пользователя
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

	_, err = qc.Insert("users").Values(user).ExecReturnID()
	if err != nil {
		log.Fatal(err)
	}

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
		Exec()
	if err != nil {
		log.Fatal(err)
	}

	// Пример 4: Удаление пользователя
	_, err = qc.Delete("users").
		WhereEq("email", "john@example.com").
		Exec()
	if err != nil {
		log.Fatal(err)
	}
}
