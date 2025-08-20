package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/antibomberman/querycraft"
	_ "github.com/go-sql-driver/mysql"
)

func ExamplePaginate() {
	// Подключение к базе данных (в данном примере используем MySQL из docker-compose)
	db, err := sql.Open("mysql", "test_user:test_password@tcp(127.0.0.1:3336)/test_db?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Создание QueryCraft с опциями
	qc, err := querycraft.New("mysql", db, querycraft.Options{
		LogEnabled:        true,
		LogLevel:          querycraft.LogLevelInfo,
		LogFormat:         querycraft.LogFormatText,
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

	// Вставка тестовых данных
	for i := 1; i <= 50; i++ {
		user := struct {
			Name      string    `db:"name"`
			Email     string    `db:"email"`
			Age       int       `db:"age"`
			CreatedAt time.Time `db:"created_at"`
			UpdatedAt time.Time `db:"updated_at"`
		}{
			Name:      fmt.Sprintf("User %d", i),
			Email:     fmt.Sprintf("user%d@example.com", i),
			Age:       20 + (i % 50),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		_, err := qc.Insert("users").Values(user).Exec()
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Println("=== Примеры пагинации ===")

	// Пример 1: Обычная пагинация
	fmt.Println("n--- Обычная пагинация (страница 2, 10 записей на странице) ---")
	paginationResult, err := qc.Select("*").From("users").OrderBy("id").Paginate(2, 10)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Всего записей: %dn", paginationResult.Total)
	fmt.Printf("Записей на странице: %dn", paginationResult.PerPage)
	fmt.Printf("Текущая страница: %dn", paginationResult.CurrentPage)
	fmt.Printf("Последняя страница: %dn", paginationResult.LastPage)
	fmt.Printf("С записи: %dn", paginationResult.From)
	fmt.Printf("По запись: %dn", paginationResult.To)
	fmt.Printf("Количество записей на текущей странице: %dn", len(paginationResult.Data))
	fmt.Println("Данные:")
	for _, user := range paginationResult.Data {
		fmt.Printf("  ID: %v, Name: %v, Email: %vn", user["id"], user["name"], user["email"])
	}

	// Пример 2: Кейсет пагинация
	fmt.Println("n--- Кейсет пагинация (первые 10 записей) ---")
	keysetResult, err := qc.Select("*").From("users").OrderBy("id").KeysetPaginate("id", nil, 10, "asc")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Есть еще записи: %vn", keysetResult.HasMore)
	fmt.Printf("Следующий курсор: %vn", keysetResult.NextCursor)
	fmt.Printf("Предыдущий курсор: %vn", keysetResult.PrevCursor)
	fmt.Printf("Количество записей: %dn", len(keysetResult.Data))
	fmt.Println("Данные:")
	for _, user := range keysetResult.Data {
		fmt.Printf("  ID: %v, Name: %v, Email: %vn", user["id"], user["name"], user["email"])
	}

	// Пример 3: Кейсет пагинация - следующая страница
	if keysetResult.HasMore && keysetResult.NextCursor != nil {
		fmt.Println("n--- Кейсет пагинация (следующие 10 записей) ---")
		nextKeysetResult, err := qc.Select("*").From("users").OrderBy("id").KeysetPaginate("id", keysetResult.NextCursor, 10, "asc")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Есть еще записи: %vn", nextKeysetResult.HasMore)
		fmt.Printf("Следующий курсор: %vn", nextKeysetResult.NextCursor)
		fmt.Printf("Предыдущий курсор: %vn", nextKeysetResult.PrevCursor)
		fmt.Printf("Количество записей: %dn", len(nextKeysetResult.Data))
		fmt.Println("Данные:")
		for _, user := range nextKeysetResult.Data {
			fmt.Printf("  ID: %v, Name: %v, Email: %vn", user["id"], user["name"], user["email"])
		}
	}

	fmt.Println("nПримеры пагинации завершены")
}
