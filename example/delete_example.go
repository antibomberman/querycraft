package main

import (
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func ExampleDelete() {
	var err error
	// Clear existing data
	QC.Schema().ClearTable("users")

	// Insert sample data
	QC.Insert("users").
		Columns("name", "email", "age", "created_at", "updated_at").
		Values("John Doe", "john@example.com", 30, time.Now(), time.Now()).
		Exec()

	QC.Insert("users").
		Columns("name", "email", "age", "created_at", "updated_at").
		Values("Jane Smith", "jane@example.com", 25, time.Now(), time.Now()).
		Exec()

	QC.Insert("users").
		Columns("name", "email", "age", "created_at", "updated_at").
		Values("Bob Johnson", "bob@example.com", 17, time.Now(), time.Now()).
		Exec()

	fmt.Println("=== Delete Examples ===")

	// Example 1: Delete specific user
	result, err := QC.Delete("users").
		WhereEq("email", "bob@example.com").
		Exec()
	if err != nil {
		log.Fatal(err)
	}
	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Deleted %d users\n", rowsAffected)

	// Example 2: Delete with complex conditions
	result, err = QC.Delete("users").
		Where("age", "<", 18).
		Exec()
	if err != nil {
		log.Fatal(err)
	}
	rowsAffected, _ = result.RowsAffected()
	fmt.Printf("Deleted %d users with age < 18\n", rowsAffected)

	// Example 3: Delete with IN clause
	result, err = QC.Delete("users").
		WhereIn("email", "john@example.com", "jane@example.com").
		Exec()
	if err != nil {
		log.Fatal(err)
	}
	rowsAffected, _ = result.RowsAffected()
	fmt.Printf("Deleted %d users with IN clause\n", rowsAffected)

	// Insert more data for remaining examples
	QC.Insert("users").
		Columns("name", "email", "age", "created_at", "updated_at").
		Values("Alice Cooper", "alice@example.com", 28, time.Now(), time.Now()).
		Exec()

	QC.Insert("users").
		Columns("name", "email", "age", "created_at", "updated_at").
		Values("Charlie Brown", "charlie@example.com", 22, time.Now(), time.Now()).
		Exec()

	// Example 5: Delete with limit
	result, err = QC.Delete("users").
		Limit(1).
		Exec()
	if err != nil {
		log.Fatal(err)
	}
	rowsAffected, _ = result.RowsAffected()
	fmt.Printf("Deleted %d users with LIMIT\n", rowsAffected)
}
