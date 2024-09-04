package ODXT

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func MySQLSetup() error {
	// Connect to the MySQL database
	db, err := sql.Open("mysql", "root:123456@tcp(localhost:3306)/ODXT")
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer db.Close()

	return nil
}
