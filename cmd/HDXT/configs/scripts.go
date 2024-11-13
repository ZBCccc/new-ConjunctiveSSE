package main

import (
	"ConjunctiveSSE/pkg/Database"
	"ConjunctiveSSE/pkg/utils"
	"context"
	"log"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	PlaintextDB, err := Database.MongoDBSetup("Crime_USENIX_REV")
	if err != nil {
		log.Println(err)
		return
	}
	defer PlaintextDB.Client().Disconnect(context.TODO())

	collection := PlaintextDB.Collection("id_keywords")

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
	var idKeywords []bson.M
	if err = cur.All(ctx, &idKeywords); err != nil {
		log.Fatal(err)
	}

	var keywordsList []string
	for _, idKeyword := range idKeywords {
		valSet, ok := idKeyword["val_set"].(primitive.A)
		if !ok {
			log.Println("val_set is not of type primitive.A")
			return 
		}
		var keywords []string
		for _, v := range valSet {
			if str, ok := v.(string); ok {
				keywords = append(keywords, str)
			} else {
				log.Println("val_set contains non-string value")
				return
			}
		}
		keywords = utils.RemoveDuplicates(keywords)
		keywordsList = append(keywordsList, keywords...)
	}

	keywordsList = utils.RemoveDuplicates(keywordsList)
	// 从keywordsList中随机选择2个关键词，共形成numPairs对
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	keywordsPair := make([][]string, 10000)
	for i := 0; i < 10000; i++ {
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
	keywordsSix := make([][]string, 10000)
	for i := 0; i < 10000; i++ {
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
}