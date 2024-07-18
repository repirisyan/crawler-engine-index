// db/db.go
package mysql

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"log"
	"os"
)

var DB *sql.DB

func Init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	host := os.Getenv("DB_HOST_MYSQL")
	user := os.Getenv("DB_USER_MYSQL")
	password := os.Getenv("DB_PASSWORD_MYSQL")
	database := os.Getenv("DB_DATABASE_MYSQL")

	// Define the data source name (DSN) for root user with no password
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, password, host, database)

	// Open a connection to the database
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal(err)
	}

	// Test the connection
	err = DB.Ping()
	if err != nil {
		log.Fatal(err)
	}
}
