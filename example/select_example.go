package main

import (
	"fmt"
	"log"
	"time"

	"github.com/antibomberman/querycraft"
	_ "github.com/go-sql-driver/mysql"
)

func ExampleSelect() {
	var err error

	// Clear existing data
	QC.Schema().DropTable("users")

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
	for i := 0; i < 50; i++ {
		_, err := QC.Insert("users").
			Columns("name", "email", "age", "created_at", "updated_at").
			Values(
				fmt.Sprintf("User %d", i),
				fmt.Sprintf("user%d@example.com", i),
				20+(i%50),
				time.Now(),
				time.Now(),
			).
			Exec()
		if err != nil {
			log.Println(err)
		}
	}

	fmt.Println("=== Select Examples ===")

	// Example 1: Select all users
	var users []User
	err = QC.Select("*").From("users").All(&users)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Selected %d users\n", len(users))

	// Example 2: Select single user
	var user User
	err = QC.Select("*").From("users").WhereEq("email", "user10@example.com").One(&user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Selected user: %s (%s)\n", user.Name.String, user.Email)

	// Example 3: Select with complex conditions
	var users2 []User
	err = QC.Select("*").From("users").
		WhereGroup(func(sb querycraft.SelectBuilder) querycraft.SelectBuilder {
			return sb.Where("age", ">", 25).Where("age", "<", 40)
		}).
		OrWhere("email", "LIKE", "%example.com").
		OrderByDesc("created_at").
		Limit(10).
		All(&users2)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Selected %d users with complex conditions\n", len(users2))

	// Example 4: Select with aggregations
	count, err := QC.Select().From("users").Count()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Total users: %d\n", count)

	avgAge, err := QC.Select().From("users").Avg("age")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Average user age: %.2f\n", avgAge)

	maxAge, err := QC.Select().From("users").Max("age")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Maximum user age: %v\n", maxAge)

	// Example 5: Pluck operation
	emails, err := QC.Select().From("users").Pluck("email")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("First 5 user emails: %v\n", emails[:5])

	// Example 6: Select distinct values
	var distinctAges []int
	err = QC.Select("DISTINCT age").From("users").All(&distinctAges)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Distinct ages: %d values\n", len(distinctAges))

	// Example 7: Select with GROUP BY and HAVING
	type AgeGroup struct {
		Age   int `db:"age"`
		Count int `db:"count"`
	}
	var ageGroups []AgeGroup
	err = QC.Select("age", "COUNT(*) as count").
		From("users").
		GroupBy("age").
		Having("COUNT(*) > ?", 1).
		All(&ageGroups)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Age groups with more than 1 user: %d\n", len(ageGroups))

	QC.Schema().DropTable("posts")
	// Example 8: Select with JOIN
	// First create a posts table for join example
	err = QC.Schema().CreateTable("posts", func(builder querycraft.TableBuilder) {
		builder.ID()
		builder.BigInteger("user_id").NotNull()
		builder.String("title", 255).NotNull()
		builder.Text("content").Nullable()
		builder.Timestamp("created_at").NotNull()
	})

	if err != nil {
		fmt.Println("create table posts err", err.Error())
		return
	}

	// Insert sample posts
	_, err = QC.Insert("posts").
		Columns("user_id", "title", "content", "created_at").
		Values(1, "First Post", "Content of first post", time.Now()).
		Exec()
	if err != nil {
		fmt.Println("insert posts err", err.Error())

		return
	}

	QC.Insert("posts").
		Columns("user_id", "title", "content", "created_at").
		Values(2, "Second Post", "Content of second post", time.Now()).
		Exec()

	// Select with join
	var postsWithUsers []map[string]any
	err = QC.Select("posts.title", "users.name").
		From("posts").
		LeftJoin("users", "posts.user_id = users.id").
		All(&postsWithUsers)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Selected %d posts with user info\n", len(postsWithUsers))

	// Example 9: Exists check
	exists, err := QC.Select().From("users").WhereEq("email", "user10@example.com").Exists()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("User with email user10@example.com exists: %v\n", exists)

	// Example 10: Raw select
	var userCount struct {
		Count int `db:"count"`
	}
	err = QC.Select("COUNT(*) as count").From("users").One(&userCount)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("User count from raw query: %d\n", userCount.Count)
}
