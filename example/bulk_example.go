package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/antibomberman/querycraft"
	_ "github.com/go-sql-driver/mysql"
)

func ExampleBulk() {
	var err error
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

	fmt.Println("=== Bulk Examples ===")

	// Example 1: Bulk insert
	var bulkUsers []User
	for i := 0; i < 1000; i++ {
		bulkUsers = append(bulkUsers, User{
			Name:      sql.NullString{String: fmt.Sprintf("Bulk User %d", i), Valid: true},
			Email:     fmt.Sprintf("bulkuser%d@example.com", i),
			Age:       20 + (i % 50),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	bulk := QC.Bulk()
	err = bulk.BulkInsert("users", bulkUsers, querycraft.WithBatchSize(100))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Bulk inserted %d usersn", len(bulkUsers))

	// Example 2: Bulk update
	// First, update some users to have a specific pattern
	var updateUsers []User
	for i := 0; i < 100; i++ {
		updateUsers = append(updateUsers, User{
			ID:        int64(i + 1),
			Name:      sql.NullString{String: fmt.Sprintf("Updated User %d", i), Valid: true},
			Email:     fmt.Sprintf("updateduser%d@example.com", i),
			Age:       30 + (i % 30),
			UpdatedAt: time.Now(),
		})
	}

	err = bulk.BulkUpdate("users", updateUsers, querycraft.WithBatchSize(50))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Bulk updated %d usersn", len(updateUsers))

	// Example 3: Bulk delete
	conditions := []map[string]any{
		{"age": 25},
		{"age": 30},
		{"email": "bulkuser500@example.com"},
	}

	err = bulk.BulkDelete("users", conditions)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Bulk deleted users with specified conditionsn")

	// Example 4: Bulk upsert
	var upsertUsers []User
	for i := 900; i < 1100; i++ {
		upsertUsers = append(upsertUsers, User{
			ID:        int64(i),
			Name:      sql.NullString{String: fmt.Sprintf("Upserted User %d", i), Valid: true},
			Email:     fmt.Sprintf("upserteduser%d@example.com", i),
			Age:       25 + (i % 20),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	err = bulk.BulkUpsert("users", upsertUsers, []string{"email"}, querycraft.WithBatchSize(50))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Bulk upserted %d usersn", len(upsertUsers))

	// Example 5: Process in batches
	fmt.Println("Processing users in batches:")
	err = bulk.ProcessInBatches(
		QC.Select("*").From("users"),
		50,
		func(batch any) error {
			users, ok := batch.([]map[string]any)
			if !ok {
				return fmt.Errorf("unexpected batch type")
			}
			fmt.Printf("  Processing batch of %d usersn", len(users))
			return nil
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	// Example 6: Bulk update by key
	var keyUpdateUsers []User
	for i := 100; i < 200; i++ {
		keyUpdateUsers = append(keyUpdateUsers, User{
			ID:        int64(i),
			Name:      sql.NullString{String: fmt.Sprintf("Key Updated User %d", i), Valid: true},
			Age:       40 + (i % 10),
			UpdatedAt: time.Now(),
		})
	}

	err = bulk.BulkUpdateByKey("users", keyUpdateUsers, "id")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Bulk updated %d users by keyn", len(keyUpdateUsers))
}
