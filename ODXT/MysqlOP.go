package ODXT

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

func MySQLSetup(dbName string) (*sql.DB, error) {
	// Connect to the MySQL database
	db, err := sql.Open("mysql", "root:123456@tcp(localhost:3306)/")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer db.Close()

	// 如果存在同名数据库，则删除
	_, err = db.Exec("DROP DATABASE IF EXISTS " + dbName)
	if err != nil {
		return nil, err
	}

	// 创建数据库
	_, err = db.Exec("CREATE DATABASE " + dbName)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func MySQLInsert(db *sql.DB, tableName string, data []byte) error {
	stmt, err := db.Prepare("INSERT INTO " + tableName + " (data) VALUES (?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(data)
	if err != nil {
		return err
	}

	return nil
}
