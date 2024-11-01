package HDXT

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MySQLSetup sets up the MySQL database
func MySQLSetup(tableName string) (*gorm.DB, error) {
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
	if err := db.AutoMigrate(&MitraCipherText{}, &AuhmeCipherText{}); err != nil {
		return nil, err
	}

	return db, nil
}

// WriteMitraCipherList writes a list of MitraCipherText to the database
func WriteMitraCipherList(db *gorm.DB, mitraCipherList []MitraCipherText) error {
	return db.Create(&mitraCipherList).Error
}

// WriteAuhmeCipherList inserts a list of AuhmeCipherText records into the provided Gorm database connection.
func WriteAuhmeCipherList(db *gorm.DB, auhmeCipherList []AuhmeCipherText) error {
	return db.Create(&auhmeCipherList).Error
}
