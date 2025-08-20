package main

import (
	"database/sql"
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
	ExampleDelete()

	defer DB.Close()
}
