package HDXT

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MySQLSetup sets up the MySQL database
func MySQLSetup(dbName string) (*gorm.DB, error) {
	// Connect to the MySQL database
	dsn := "root:123456@tcp(127.0.0.1:3306)/HDXT?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	dbConn, err := db.DB() // 获取 *sql.DB 实例
	if err != nil {
		return nil, err
	}

	// Ping the database
	err = dbConn.Ping()
	if err != nil {
		return nil, err
	}

	// Initialize the model, figure out the table name
	tableName := dbName
	db = db.Table(tableName)
	if err := db.AutoMigrate(&CipherText{}); err != nil {
		return nil, err
	}

	return db, nil
}

func WriteUploadList(db *gorm.DB, uploadList []CipherText) error {
	return db.Create(&uploadList).Error
}
