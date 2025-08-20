package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/antibomberman/querycraft"
	_ "github.com/go-sql-driver/mysql"
)

func ExampleInsert() {
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

	fmt.Println("=== Insert Examples ===")

	// Example 1: Insert using struct
	user := User{
		Name:      sql.NullString{String: "John Doe", Valid: true},
		Email:     "john@example.com",
		Age:       30,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	id, err := QC.Insert("users").Values(user).ExecReturnID()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Inserted user with ID: %d\n", id)

	// Example 2: Insert with specific columns
	id, err = QC.Insert("users").
		Columns("name", "email", "age", "created_at", "updated_at").
		Values("Jane Smith", "jane@example.com", 25, time.Now(), time.Now()).
		ExecReturnID()
	if err != nil {
		log.Fatal(2, err)
	}
	fmt.Printf("Inserted user with ID: %d\n", id)

	// Example 3: Insert using map
	id, err = QC.Insert("users").
		ValuesMap(map[string]any{
			"name":       "Bob Johnson",
			"email":      "bob@example.com",
			"age":        35,
			"created_at": time.Now(),
			"updated_at": time.Now(),
		}).
		ExecReturnID()
	if err != nil {
		log.Fatal(3, err)
	}
	fmt.Printf("Inserted user with ID: %d\n", id)

	// Example 4: Insert with ON CONFLICT handling
	id, err = QC.Insert("users").
		Columns("name", "email", "age", "created_at", "updated_at").
		Values("Duplicate Email", "john@example.com", 40, time.Now(), time.Now()).
		Ignore().
		ExecReturnID()
	if err != nil {
		log.Fatal(4, err)
	}
	// Example 4: Insert with ON CONFLICT handling
	id, err = QC.Insert("users").
		Columns("name", "email", "age", "created_at", "updated_at").
		Values("Duplicate Email", "john@example.com", 40, time.Now(), time.Now()).
		Ignore().
		ExecReturnID()
	if err != nil {
		log.Fatal(4, err)
	}
	fmt.Printf("Insert with conflict handling, ID: %d\n", id)

	// Example 5: Insert from select
	id, err = QC.Insert("users").
		Columns("name", "email", "age", "created_at", "updated_at").
		FromSelect(QC.Select("name", "email", "age", "created_at", "updated_at").From("users").WhereEq("email", "john@example.com")).
		Ignore().
		ExecReturnID()
	if err != nil {
		log.Fatal(5, err)
	}
	fmt.Printf("Inserted from select with ID: %d\n", id)
}
