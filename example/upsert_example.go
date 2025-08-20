package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/antibomberman/querycraft"
	_ "github.com/go-sql-driver/mysql"
)

func ExampleUpsert() {
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

	QC.Insert("users").
		Columns("name", "email", "age", "created_at", "updated_at").
		Values("Bob Johnson", "bob@example.com", 35, time.Now(), time.Now()).
		Exec()

	fmt.Println("=== Upsert Examples ===")

	// Example 1: Upsert user (insert new)
	user := User{
		Name:      sql.NullString{String: "Alice Cooper", Valid: true},
		Email:     "alice@example.com",
		Age:       28,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err := QC.Upsert("users").
		Values(user).
		OnConflict("email").
		DoUpdate("name", "age", "updated_at").
		Exec()
	if err != nil {
		log.Fatal(err)
	}
	id, _ := result.LastInsertId()
	fmt.Printf("Upserted user with ID: %dn", id)

	// Example 2: Upsert user (update existing)
	userUpdated := User{
		Name:      sql.NullString{String: "John Smith", Valid: true},
		Email:     "john@example.com", // This email already exists
		Age:       31,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err = QC.Upsert("users").
		Values(userUpdated).
		OnConflict("email").
		DoUpdate("name", "age", "updated_at").
		Exec()
	if err != nil {
		log.Fatal(err)
	}
	id, _ = result.LastInsertId()
	fmt.Printf("Upserted user with ID: %dn", id)

	// Example 3: Upsert with DoNothing
	user2 := User{
		Name:      sql.NullString{String: "Charlie Brown", Valid: true},
		Email:     "charlie@example.com",
		Age:       22,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err = QC.Upsert("users").
		Values(user2).
		OnConflict("email").
		DoNothing().
		Exec()
	if err != nil {
		log.Fatal(err)
	}
	id, _ = result.LastInsertId()
	fmt.Printf("Upserted with DoNothing, ID: %dn", id)

	// Example 4: Upsert with DoUpdateExcept
	user3 := User{
		Name:      sql.NullString{String: "Jane Updated", Valid: true},
		Email:     "jane@example.com", // This email already exists
		Age:       27,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	result, err = QC.Upsert("users").
		Values(user3).
		OnConflict("email").
		DoUpdateExcept("created_at"). // Update all except created_at
		Exec()
	if err != nil {
		log.Fatal(err)
	}
	id, _ = result.LastInsertId()
	fmt.Printf("Upserted with DoUpdateExcept, ID: %dn", id)
}
