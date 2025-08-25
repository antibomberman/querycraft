package main

import (
	"database/sql"
	"fmt"
	"github.com/antibomberman/querycraft"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"time"
)

type User struct {
	ID        int64          `db:"id"`
	Name      sql.NullString `db:"name"`
	Email     string         `db:"email"`
	Age       int            `db:"age"`
	CreatedAt time.Time      `db:"created_at"`
	UpdatedAt time.Time      `db:"updated_at"`
}

var DB *sql.DB
var QC querycraft.QueryCraft

func Connect() {
	var err error
	DB, err = sql.Open("mysql", "root:rootpassword@tcp(127.0.0.1:3336)/test_db?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	QC, err = querycraft.New("mysql", DB)
	if err != nil {
		log.Fatal(err)
	}

}
func main() {
	Connect()
	defer DB.Close()
	/*
		QC.Schema().ClearTable("items")

		// Create users table
		QC.Schema().CreateTable("items", func(builder querycraft.TableBuilder) {
			builder.ID()
			builder.String("name", 100).Nullable()
			builder.Text("desc")
			builder.Integer("number").Default(0)

			builder.Timestamp("created_at").NotNull()
			builder.Timestamp("updated_at").NotNull()
		})

		_, err := QC.Insert("items").Columns("name", "desc", "number", "created_at", "updated_at").ValuesMap(map[string]any{
			"name":       "test",
			"desc":       "desc test",
			"number":     12321,
			"created_at": time.Now(),
			"updated_at": time.Now(),
		}).Exec()
		if err != nil {
			panic(err)
		}
	*/

	var numbers []int

	err := QC.Select("SUM(number)").From("items").All(&numbers)
	if err != nil {
		panic(err)
	}
	fmt.Println(numbers)

}
