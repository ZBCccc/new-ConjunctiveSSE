package ODXT

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLSetup sets up the MySQL database
func MySQLSetup(tableName string) (*sql.DB, error) {
	// Connect to the MySQL database
	db, err := sql.Open("mysql", "root:123456@tcp(localhost:3306)/ODXT")
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// 测试数据库连接
	err = db.Ping()
	if err != nil {
		log.Fatal("Cannot connect to database:", err)
		return nil, err
	}

	// 创建数据表tableName;如果表不存在则创建，如果表存在则不创建
	// 表的结构为：id, address, value, alpha, created_at
	// id 为自增主键
	// address 为地址
	// value 为值
	// alpha 为alpha
	// created_at 为创建时间
	createTableSQL := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id INT AUTO_INCREMENT PRIMARY KEY,
		address VARCHAR(255) NOT NULL,
		value VARCHAR(255) NOT NULL,
		alpha VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`, tableName)

	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	return db, nil
}

// WriteUploadList writes the upload list to the MySQL database
func WriteUploadList(db *sql.DB, uploadList []UpdatePayload, tableName string) error {
	// 准备插入语句，将结构体切片逐个写入数据库，将结构体中的Address, Value, Alpha写入数据库
	insertSQL := "INSERT INTO " + tableName + " (address, value, alpha) VALUES (?, ?, ?)"
	stmt, err := db.Prepare(insertSQL)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer stmt.Close()

	for _, payload := range uploadList {
		_, err = stmt.Exec(payload.Address, payload.Val, payload.Alpha)
		if err != nil {
			log.Fatal(err)
			return err
		}
	}
	return nil
}

type SearchPayload struct {
	Value string
	Alpha string
}

// SearchStoken searches the stokenList in the MySQL database
func SearchStoken(db *sql.DB, address []string, tableName string) ([]SearchPayload, error) {
	// 准备查询语句，查询数据库中的value和alpha
	querySQL := "SELECT value, alpha FROM " + tableName + " WHERE address = ?"

	// 查询数据库
	var value, alpha string
	result := make([]SearchPayload, len(address))
	for index, addr := range address {
		err := db.QueryRow(querySQL, addr).Scan(&value, &alpha)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		result[index] = SearchPayload{Value: value, Alpha: alpha}
	}

	return result, nil
}
