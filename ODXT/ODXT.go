package ODXT

import (
	"context"
	"crypto/rand"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var PlaintextDB *mongo.Collection

const (
	MaxConnection      = 100
	UploadListMaxLength = 200000
)

func DBSetup(db_name string) error {
	// 生成4个32字节长度的随机私钥
	keys := make([][]byte, 4)
	keyLen := 32

	for i := 0; i < 4; i++ {
		key := make([]byte, keyLen)
		if _, err := rand.Read(key); err != nil {
			log.Println("Error generating random key:", err)
			return err
		}
		keys[i] = key
	}

	// 连接MongoDB
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
		return err
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
		return err
	}

	// 获取数据库和集合的句柄
	PlaintextDB = client.Database("ODXT").Collection(db_name)



	return nil
}

func Ciphertext_Gen_Phase() {

}

func Encrypt(keyword string, ids []string, operation string) error {


	return nil
}

func Deletion_phase_with_search(del_rate int) {
}
