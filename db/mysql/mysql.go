// db/db.go
package mysql

import (
    "database/sql"
    "fmt"
    "log"

    _ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func Init() {
    // Define the data source name (DSN) for root user with no password
    dsn := "root:@tcp(127.0.0.1:3306)/crawler"

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
