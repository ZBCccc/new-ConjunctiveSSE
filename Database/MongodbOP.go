package Database

import (
	"ConjunctiveSSE/util"
	"context"
	"log"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func MongoDBSetup(dbName string) (*mongo.Database, error) {
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	// 获取数据库和集合的句柄
	PlaintextDB := client.Database(dbName)

	return PlaintextDB, nil
}

func GenQuerydataFromDB(dbName, tableName string, numPairs int) error {
	PlaintextDB, err := MongoDBSetup(dbName)
	if err != nil {
		log.Fatal(err)
		return err
	}
	defer PlaintextDB.Client().Disconnect(context.TODO())

	collection := PlaintextDB.Collection(tableName)

	// 创建一个游标，设置不超时并每次获取1000条记录
	ctx := context.TODO()
	opts := options.Find().SetNoCursorTimeout(true).SetBatchSize(1000)
	cur, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		log.Fatal(err)
	}

	// 关闭游标
	defer cur.Close(ctx)

	// 读取游标中的所有记录
	var keywordIds []bson.M
	if err = cur.All(ctx, &keywordIds); err != nil {
		log.Fatal(err)
	}

	var keywordsList []string
	for _, keywordId := range keywordIds {
		keyword := keywordId["k"].(string)
		keywordsList = append(keywordsList, keyword)
	}

	// 从keywordsList中随机选择2个关键词，共形成numPairs对
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	keywordsPair := make([][]string, numPairs)
	for i := 0; i < numPairs; i++ {
		// 创建一个新的切片来避免重复
		shuffledKeywords := make([]string, len(keywordsList))
		copy(shuffledKeywords, keywordsList)
		
		// 随机选择两个不同的关键词
		r.Shuffle(len(shuffledKeywords), func(i, j int) {
			shuffledKeywords[i], shuffledKeywords[j] = shuffledKeywords[j], shuffledKeywords[i]
		})
		keywordsPair[i] = shuffledKeywords[:2]
	}
	util.WriteResultToFile("keywords_2.txt", keywordsPair)

	// 从keywordsList中随机选择6个关键词，共形成numPairs对
	keywordsSix := make([][]string, numPairs)
	for i := 0; i < numPairs; i++ {
		// 创建一个新的切片来避免重复
		shuffledKeywords := make([]string, len(keywordsList))
		copy(shuffledKeywords, keywordsList)
		
		// 随机选择两个不同的关键词
		r.Shuffle(len(shuffledKeywords), func(i, j int) {
			shuffledKeywords[i], shuffledKeywords[j] = shuffledKeywords[j], shuffledKeywords[i]
		})
		keywordsSix[i] = shuffledKeywords[:6]
	}
	util.WriteResultToFile("keywords_6.txt", keywordsSix)

	return nil
}