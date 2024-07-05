// db/db.go
package mysql

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    "github.com/joho/godotenv"
    _ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func Init() {
    godotenv.Load()
    host := os.Getenv("DB_HOST_MYSQL")
    user := os.Getenv("DB_USER_MYSQL")
    password := os.Getenv("DB_PASSWORD_MYSQL")
    database := os.Getenv("DB_DATABASE_MYSQL")

    // Define the data source name (DSN) for root user with no password
    dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, password, host, database)

    // Open a connection to the database
    var err error
    DB, err = sql.Open("mysql", dsn)
    if err != nil {
        log.Fatal(err)
    }

    // Test the connection
    err = DB.Ping()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Successfully connected to the database!")
}
