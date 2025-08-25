# QueryCraft

QueryCraft is a powerful and flexible Go library for building and executing SQL queries. It provides a fluent interface for constructing complex queries while maintaining type safety and performance.

## Features

- Fluent API for building SELECT, INSERT, UPDATE, DELETE, and UPSERT queries
- Schema management (CREATE, ALTER, DROP tables)
- Bulk operations for high-performance data manipulation
- Database transactions
- Raw SQL query support
- Database migrations
- MySQL support with extensible dialect system
- SQL query logging with file output
- PrintSQL() method for debugging queries
- Easy-to-use options-based configuration for logging

## Installation

```bash
go get github.com/antibomberman/querycraft@v0.0.7
```

## Quick Start

```go
import (
    "database/sql"
    "github.com/antibomberman/querycraft"
    _ "github.com/go-sql-driver/mysql"
)

// Connect to database
db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/dbname")

// Create QueryCraft with options for logging to files only in JSON format
qc, err := querycraft.New("mysql", db, querycraft.Options{
    LogEnabled:        true,
    LogLevel:          querycraft.LogLevelInfo,
    LogFormat:         querycraft.LogFormatJSON, // Use JSON format
    LogSaveToFile:     true,
    LogPrintToConsole: false, // Don't print to console
    LogDir:            "./storage/logs/sql/",
    LogAutoCleanDays:  7,
})

// Insert a record
user := User{Name: "John", Email: "john@example.com"}
id, err := qc.Insert("users").Values(user).ExecReturnID()

// Select records with PrintSQL for debugging (this will print to console)
var users []User
err = qc.Select("*").From("users").Where("age", ">", 18).PrintSQL().All(&users)

// Update records
_, err = qc.Update("users").Set("name", "Jane").WhereEq("id", id).Exec()

// Delete records
_, err = qc.Delete("users").WhereEq("id", id).Exec()
```

## Examples

Check out the [examples](example/) directory for comprehensive examples of all QueryCraft features:

- [Full Example](example/full_example.go) - A comprehensive example showing all QueryCraft features
- [Insert Operations](example/insert_example.go) - Demonstrates various INSERT operations
- [Update Operations](example/update_example.go) - Shows UPDATE operations including increment/decrement
- [Upsert Operations](example/upsert_example.go) - Examples of UPSERT operations
- [Delete Operations](example/delete_example.go) - DELETE operations with various conditions
- [Schema Operations](example/schema_example.go) - Schema operations (create, alter, drop tables)
- [Select Operations](example/select_example.go) - SELECT operations with conditions, joins, aggregations
- [Logger Example](example/logger_example.go) - Examples of SQL query logging and debugging with old API
- [New API Example](example/new_api_example.go) - Examples of SQL query logging and debugging with new options-based API
- [Bulk Operations](example/bulk_example.go) - Bulk operations for high-performance data manipulation
- [Transactions](example/transaction_example.go) - Database transactions
- [Logger Example](example/logger_example.go) - Examples of SQL query logging and debugging
- [Raw Queries](example/raw_example.go) - Raw SQL queries
- [Migrations](example/migration_example.go) - Database migrations
- [Pagination Example](example/pagination_example.go) - Examples of pagination with offset and keyset pagination

## Documentation

For detailed documentation, please refer to the code and examples.

## License

MIT