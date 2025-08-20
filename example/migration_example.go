package main

import (
	"fmt"
	"log"
	"time"

	"github.com/antibomberman/querycraft"
	_ "github.com/go-sql-driver/mysql"
)

// Define a simple migration
type CreateUserTableMigration struct{}

func (m *CreateUserTableMigration) Up(schema querycraft.SchemaBuilder) error {
	return schema.CreateTable("users", func(builder querycraft.TableBuilder) {
		builder.ID()
		builder.String("name", 100).Nullable()
		builder.String("email", 255).NotNull().Unique()
		builder.Integer("age").Default(0)
		builder.Timestamp("created_at").NotNull()
		builder.Timestamp("updated_at").NotNull()
	})
}

func (m *CreateUserTableMigration) Down(schema querycraft.SchemaBuilder) error {
	return schema.DropTable("users")
}

// Another migration
type AddPhoneToUsersMigration struct{}

func (m *AddPhoneToUsersMigration) Up(schema querycraft.SchemaBuilder) error {
	return schema.AlterTable("users", func(builder querycraft.TableBuilder) {
		builder.String("phone", 20).Nullable().After("email")
	})
}

func (m *AddPhoneToUsersMigration) Down(schema querycraft.SchemaBuilder) error {
	return schema.AlterTable("users", func(builder querycraft.TableBuilder) {
		builder.DropColumn("phone")
	})
}

func ExampleMigrate() {
	var err error

	// Clear existing migrations table
	QC.Schema().ClearTable("migrations")
	fmt.Println("=== Migration Examples ===")

	// Example 1: Register migrations
	migrationManager := QC.Migration()

	err = migrationManager.RegisterMigration("create_users_table", &CreateUserTableMigration{})
	if err != nil {
		log.Fatal(err)
	}

	err = migrationManager.RegisterMigration("add_phone_to_users", &AddPhoneToUsersMigration{})
	if err != nil {
		log.Fatal(err)
	}

	// Example 2: Run migrations (Up)
	fmt.Println("Running migrations...")
	err = migrationManager.Up()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Migrations completed successfully")

	// Example 3: Check migration status
	statuses, err := migrationManager.Status()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Migration statuses (%d):n", len(statuses))
	for _, status := range statuses {
		fmt.Printf("  - %s: Applied=%vn", status.Name, status.Applied)
	}

	// Example 4: Insert data to test the schema
	QC.Insert("users").
		Columns("name", "email", "age", "phone", "created_at", "updated_at").
		Values("John Doe", "john@example.com", 30, "123-456-7890", time.Now(), time.Now()).
		Exec()

	// Example 5: Rollback last migration (Down)
	fmt.Println("Rolling back last migration...")
	err = migrationManager.Rollback(1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Rollback completed successfully")

	// Example 6: Check status after rollback
	statuses, err = migrationManager.Status()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Migration statuses after rollback (%d):n", len(statuses))
	for _, status := range statuses {
		fmt.Printf("  - %s: Applied=%vn", status.Name, status.Applied)
	}

	// Example 7: Run specific number of migrations
	fmt.Println("Running 1 migration...")
	err = migrationManager.Migrate(1)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("1 migration completed successfully")

	// Example 8: Get current migration
	current, err := migrationManager.Current()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Current migration: %sn", current)

	// Example 9: Reset all migrations
	fmt.Println("Resetting all migrations...")
	err = migrationManager.Reset()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("All migrations reset successfully")

	// Example 10: Refresh migrations (reset and re-run)
	fmt.Println("Refreshing migrations...")
	err = migrationManager.Refresh()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Migrations refreshed successfully")

	// Example 11: Get all registered migrations
	migrations := migrationManager.GetMigrations()
	fmt.Printf("Registered migrations: %dn", len(migrations))
	for name := range migrations {
		fmt.Printf("  - %sn", name)
	}
}
