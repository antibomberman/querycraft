package main

import (
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func ExampleRaw() {
	var err error

	// Clear existing data
	QC.Schema().ClearTable("users")

	fmt.Println("=== Raw Query Examples ===")

	// Example 1: Simple raw query
	raw := QC.Raw("SELECT COUNT(*) as count FROM users")
	var result struct {
		Count int `db:"count"`
	}
	err = raw.One(&result)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("User count from raw query: %dn", result.Count)

	// Insert some sample data
	QC.Insert("users").
		Columns("name", "email", "age", "created_at", "updated_at").
		Values("John Doe", "john@example.com", 30, time.Now(), time.Now()).
		Exec()

	QC.Insert("users").
		Columns("name", "email", "age", "created_at", "updated_at").
		Values("Jane Smith", "jane@example.com", 25, time.Now(), time.Now()).
		Exec()

	// Example 2: Raw query with parameters
	raw = QC.Raw("SELECT * FROM users WHERE age > ?", 25)
	var users []User
	err = raw.All(&users)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Selected %d users with age > 25 using raw queryn", len(users))

	// Example 3: Raw query with multiple parameters
	raw = QC.Raw("SELECT * FROM users WHERE age BETWEEN ? AND ? AND email LIKE ?", 20, 35, "%example.com%")
	err = raw.All(&users)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Selected %d users with complex conditions using raw queryn", len(users))

	// Example 4: Raw insert
	raw = QC.Raw("INSERT INTO users (name, email, age, created_at, updated_at) VALUES (?, ?, ?, ?, ?)",
		"Raw Insert User",
		"raw@example.com",
		28,
		time.Now(),
		time.Now())
	result2, err := raw.Exec()
	if err != nil {
		log.Fatal(err)
	}
	rowsAffected, _ := result2.RowsAffected()
	fmt.Printf("Inserted user with raw query, rows affected: %dn", rowsAffected)

	// Example 5: Raw update
	raw = QC.Raw("UPDATE users SET name = ?, age = ? WHERE email = ?", "Raw Updated User", 35, "raw@example.com")
	result3, err := raw.Exec()
	if err != nil {
		log.Fatal(err)
	}
	rowsAffected, _ = result3.RowsAffected()
	fmt.Printf("Updated user with raw query, rows affected: %dn", rowsAffected)

	// Example 6: Raw delete
	raw = QC.Raw("DELETE FROM users WHERE email = ?", "raw@example.com")
	result4, err := raw.Exec()
	if err != nil {
		log.Fatal(err)
	}
	rowsAffected, _ = result4.RowsAffected()
	fmt.Printf("Deleted user with raw query, rows affected: %dn", rowsAffected)

	// Example 7: Complex raw query with JOIN
	raw = QC.Raw(`
		SELECT u.name, u.email, COUNT(*) as user_count
		FROM users u
		WHERE u.email LIKE ?
		GROUP BY u.name, u.email
		HAVING COUNT(*) > ?
		ORDER BY user_count DESC`,
		"%example.com%",
		0)
	var results []map[string]any
	err = raw.All(&results)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Complex raw query returned %d rowsn", len(results))

	// Example 8: Raw query to create table
	raw = QC.Raw(`
		CREATE TABLE IF NOT EXISTS raw_test (
			id INT AUTO_INCREMENT PRIMARY KEY,
			data VARCHAR(255) NOT NULL
		)`)
	_, err = raw.Exec()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created table with raw query")

	// Example 9: Raw query to insert into new table
	raw = QC.Raw("INSERT INTO raw_test (data) VALUES (?)", "Test data")
	_, err = raw.Exec()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted data into raw_test table")

	// Example 10: Raw query with transaction
	tx, err := QC.Begin()
	if err != nil {
		log.Fatal(err)
	}

	raw = tx.Raw("INSERT INTO raw_test (data) VALUES (?)", "Transaction data")
	_, err = raw.Exec()
	if err != nil {
		tx.Rollback()
		log.Fatal(err)
	}

	err = tx.Commit()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Raw query executed in transaction")
}
