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
		db.Close()
		log.Fatal(err)
		return nil, err
	}

	// 测试数据库连接
	err = db.Ping()
	if err != nil {
		log.Fatal("Cannot connect to database:", err)
		return nil, err
	}

	// 创建数据表tableName
	createTableSQL := fmt.Sprintf(`
	CREATE TABLE IF NOT EXISTS %s (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(50) NOT NULL,
		email VARCHAR(100) NOT NULL UNIQUE,
		age INT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`, tableName)

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
