package main

// Adapted from:
// https://stackoverflow.com/questions/41053830/how-to-ping-remote-mysql-using-golang

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func main() {
	var (
		server = flag.String("mysql", "localhost:3306", "mysql server")
		user   = flag.String("user", "root", "mysql user")
		pass   = flag.String("password", "", "mysql password")
	)
	flag.Parse()

	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/", *user, *pass, *server))
	if err != nil {
		os.Exit(1)
	}

	defer db.Close()

	err = db.Ping()

	if err != nil {
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
