package ODXT

import (
	"database/sql"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLSetup sets up the MySQL database
func MySQLSetup(dbName string) (*sql.DB, error) {
	// Connect to the MySQL database
	db, err := sql.Open("mysql", "root:123456@tcp(localhost:3306)/")
	if err != nil {
		db.Close()
		log.Fatal(err)
		return nil, err
	}

	// 如果存在同名数据库，则删除
	_, err = db.Exec("DROP DATABASE IF EXISTS " + dbName)
	if err != nil {
		db.Close()
		log.Fatal(err)
		return nil, err
	}

	// 创建数据库
	_, err = db.Exec("CREATE DATABASE " + dbName)
	if err != nil {
		db.Close()
		log.Fatal(err)
		return nil, err
	}

	return db, nil
}

func WriteUploadList(db *sql.DB, uploadList []UpdatePayload, tableName string) error {
	// 开始事务
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// 准备插入语句
	stmt, err := tx.Prepare("INSERT INTO " + tableName + " (data) VALUES (?)")
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	// 将结构体切片逐个写入数据库
	for _, payload := range uploadList {
		_, err := stmt.Exec(payload)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// 提交事务
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
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
