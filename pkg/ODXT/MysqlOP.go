package ODXT

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"sync"
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

	// 删除表
	err = DropTable(db, tableName)
	if err != nil {
		log.Fatal(err)
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
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	insertSQL := fmt.Sprintf("INSERT INTO %s (address, value, alpha) VALUES (?, ?, ?)", tableName)
	stmt, err := tx.Prepare(insertSQL)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, payload := range uploadList {
		if payload.Address == "" || payload.Val == "" || payload.Alpha == "" {
			return fmt.Errorf("invalid payload data: %v", payload)
		}
		_, err = stmt.Exec(payload.Address, payload.Val, payload.Alpha)
		if err != nil {
			return fmt.Errorf("error inserting data: %v", err)
		}
	}

	return tx.Commit()
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
	var wg sync.WaitGroup
	result := make([]SearchPayload, len(address))
	errs := make([]error, len(address)) // 用于存储错误

	for index, addr := range address {
		wg.Add(1)
		go func(i int, addr string) {
			defer wg.Done()
			var value, alpha string
			err := db.QueryRow(querySQL, addr).Scan(&value, &alpha)
			if err != nil {
				errs[i] = err
				return
			}
			result[i] = SearchPayload{Value: value, Alpha: alpha}
		}(index, addr)
	}

	wg.Wait()

	// 检查是否有错误
	for _, err := range errs {
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
	}

	return result, nil
}

// DropTable 删除指定的表
func DropTable(db *sql.DB, tableName string) error {
	query := fmt.Sprintf("DROP TABLE IF EXISTS %s", tableName)
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("删除表 %s 时出错: %v", tableName, err)
	}
	fmt.Printf("表 %s 已成功删除\n", tableName)
	return nil
}
