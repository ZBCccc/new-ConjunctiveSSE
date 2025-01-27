package Database

import (
	"ConjunctiveSSE/pkg/utils"
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
		log.Println(err)
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
		log.Println(err)
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
	utils.WriteResultToFile("keywords_2.txt", keywordsPair)

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
	utils.WriteResultToFile("keywords_6.txt", keywordsSix)

	return nil
}

func GetUniqueValSets(PlaintextDB *mongo.Database) ([]string, error) {
	// 获取集合句柄
	collection := PlaintextDB.Collection("id_keywords")
	ctx := context.TODO()

	// 使用聚合管道提取并去重val_set
	pipeline := mongo.Pipeline{
		{{Key: "$unwind", Value: "$val_st"}},                             // 展开val_set数组
		{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$val_st"}}}}, // 按val_set的值进行分组，实现去重
	}
	// 执行聚合查询
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// 读取查询结果
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// 将结果转换为字符串切片
	var uniqueVals []string
	for _, result := range results {
		if val, ok := result["_id"].(string); ok {
			uniqueVals = append(uniqueVals, val)
		}
	}

	return uniqueVals, nil
}

func GetUniqueIDs(PlaintextDB *mongo.Database) ([]string, error) {
	// 获取集合句柄
	collection := PlaintextDB.Collection("id_keywords")
	ctx := context.TODO()

	// 使用聚合管道提取并去重id
	pipeline := mongo.Pipeline{
		{{Key: "$unwind", Value: "$id"}},                             // 展开id数组
		{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$id"}}}}, // 按id的值进行分组，实现去重
	}
	// 执行聚合查询
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// 读取查询结果
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// 将结果转换为字符串切片
	var uniqueVals []string
	for _, result := range results {
		if val, ok := result["_id"].(string); ok {
			uniqueVals = append(uniqueVals, val)
		}
	}

	return uniqueVals, nil
}
