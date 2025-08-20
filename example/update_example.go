package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/antibomberman/querycraft"
	_ "github.com/go-sql-driver/mysql"
)

func ExampleUpdate() {
	var err error

	// Clear existing data
	QC.Schema().ClearTable("users")

	// Create users table
	QC.Schema().CreateTable("users", func(builder querycraft.TableBuilder) {
		builder.ID()
		builder.String("name", 100).Nullable()
		builder.String("email", 255).NotNull().Unique()
		builder.Integer("age").Default(0)
		builder.Timestamp("created_at").NotNull()
		builder.Timestamp("updated_at").NotNull()
	})

	// Insert sample data
	QC.Insert("users").
		Columns("name", "email", "age", "created_at", "updated_at").
		Values("John Doe", "john@example.com", 30, time.Now(), time.Now()).
		Exec()

	QC.Insert("users").
		Columns("name", "email", "age", "created_at", "updated_at").
		Values("Jane Smith", "jane@example.com", 25, time.Now(), time.Now()).
		Exec()

	fmt.Println("=== Update Examples ===")

	// Example 1: Update using Set
	result, err := QC.Update("users").
		Set("name", "John Updated").
		Set("age", 31).
		WhereEq("email", "john@example.com").
		Exec()
	if err != nil {
		log.Fatal(err)
	}
	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Updated %d rows\n", rowsAffected)

	// Example 2: Update using struct
	user := User{
		Name: sql.NullString{String: "Jane Updated", Valid: true},
		Age:  26,
	}
	result, err = QC.Update("users").
		SetStruct(user).
		WhereEq("email", "jane@example.com").
		Exec()
	if err != nil {
		log.Fatal(err)
	}
	rowsAffected, _ = result.RowsAffected()
	fmt.Printf("Updated %d rows using struct\n", rowsAffected)

	// Example 3: Update using map
	result, err = QC.Update("users").
		SetMap(map[string]any{
			"name":       "Bob Updated",
			"updated_at": time.Now(),
		}).
		Where("age", ">", 20).
		Exec()
	if err != nil {
		log.Fatal(err)
	}
	rowsAffected, _ = result.RowsAffected()
	fmt.Printf("Updated %d rows using map\n", rowsAffected)

	// Example 4: Increment operation
	result, err = QC.Update("users").
		Increment("age", 1).
		WhereEq("email", "john@example.com").
		Exec()
	if err != nil {
		log.Fatal(err)
	}
	rowsAffected, _ = result.RowsAffected()
	fmt.Printf("Incremented age for %d rows\n", rowsAffected)

	// Example 5: Decrement operation
	result, err = QC.Update("users").
		Decrement("age", 1).
		WhereEq("email", "jane@example.com").
		Exec()
	if err != nil {
		log.Fatal(err)
	}
	rowsAffected, _ = result.RowsAffected()
	fmt.Printf("Decremented age for %d rows\n", rowsAffected)
}
