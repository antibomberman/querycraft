package main

import (
	"fmt"
	"log"
	"time"

	"github.com/antibomberman/querycraft"
	_ "github.com/go-sql-driver/mysql"
)

func ExampleTr() {
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

	fmt.Println("=== Transaction Examples ===")

	// Example 1: Basic transaction
	tx, err := QC.Begin()
	if err != nil {
		log.Fatal(err)
	}

	// Insert user in transaction
	id, err := tx.Insert("users").
		Columns("name", "email", "age", "created_at", "updated_at").
		Values("Transaction User", "tx@example.com", 33, time.Now(), time.Now()).
		ExecReturnID()
	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}
	fmt.Printf("Inserted user in transaction with ID: %dn", id)

	// Commit transaction
	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Transaction committed successfully")

	// Example 2: Transaction with rollback
	tx2, err := QC.Begin()
	if err != nil {
		log.Fatal(err)
	}

	// Insert user in transaction
	_, err = tx2.Insert("users").
		Columns("name", "email", "age", "created_at", "updated_at").
		Values("Rollback User", "rollback@example.com", 25, time.Now(), time.Now()).
		ExecReturnID()
	if err != nil {
		tx2.Rollback()
		log.Fatal(err)
	}
	fmt.Println("Inserted user for rollback")

	// Rollback transaction
	err = tx2.Rollback()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Transaction rolled back successfully")

	// Example 3: Nested operations in transaction
	tx3, err := QC.Begin()
	if err != nil {
		log.Fatal(err)
	}

	// Insert multiple users
	for i := 0; i < 5; i++ {
		_, err = tx3.Insert("users").
			Columns("name", "email", "age", "created_at", "updated_at").
			Values(
				fmt.Sprintf("Nested User %d", i),
				fmt.Sprintf("nested%d@example.com", i),
				20+i,
				time.Now(),
				time.Now(),
			).
			ExecReturnID()
		if err != nil {
			tx3.Rollback()
			log.Fatal(err)
		}
	}

	// Update a user
	_, err = tx3.Update("users").
		Set("name", "Updated Nested User").
		WhereEq("email", "nested0@example.com").
		Exec()
	if err != nil {
		tx3.Rollback()
		log.Fatal(err)
	}

	// Delete a user
	_, err = tx3.Delete("users").
		WhereEq("email", "nested4@example.com").
		Exec()
	if err != nil {
		tx3.Rollback()
		log.Fatal(err)
	}

	// Commit all changes
	err = tx3.Commit()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Nested transaction operations committed successfully")

	// Example 4: Transaction with select operations
	tx4, err := QC.Begin()
	if err != nil {
		log.Fatal(err)
	}

	// Insert a user
	id, err = tx4.Insert("users").
		Columns("name", "email", "age", "created_at", "updated_at").
		Values("Select Test User", "select@example.com", 28, time.Now(), time.Now()).
		ExecReturnID()
	if err != nil {
		tx4.Rollback()
		log.Fatal(err)
	}

	// Select the user within the transaction
	var user User
	err = tx4.Select("*").From("users").WhereEq("id", id).One(&user)
	if err != nil {
		tx4.Rollback()
		log.Fatal(err)
	}
	fmt.Printf("Selected user in transaction: %s (%s)n", user.Name.String, user.Email)

	// Commit transaction
	err = tx4.Commit()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Transaction with select committed successfully")
}
