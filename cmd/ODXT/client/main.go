package main

import (
	"ConjunctiveSSE/pkg/Database"
	"ConjunctiveSSE/pkg/ODXT/client"
	"ConjunctiveSSE/pkg/utils"
	"context"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	c, err := client.NewODXTClient("localhost:50051")
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	// 执行实验
	dbName := "Crime_USENIX_REV"
	updateTime := time.Now()
	UpdatePhase(c, dbName)
	log.Printf("Update phase to %v\n", time.Since(updateTime))

	// Search Phase
	searchTime := time.Now()
	SearchPhase(c, dbName)
	log.Printf("Search phase to %v\n", time.Since(searchTime))
}

func UpdatePhase(c *client.ODXTClient, dbName string) {
	PlaintextDB, err := Database.MongoDBSetup(dbName)
	if err != nil {
		log.Fatal(err)
	}
	defer PlaintextDB.Client().Disconnect(context.Background())

	// Update Phase
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	collection := PlaintextDB.Collection("keyword_ids")
	count, err := collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	encryptTimeList := make([]time.Duration, 0, count)
	keywordList := make([]string, 0, count)
	volumeList := make([]int, 0, count)

	opts := options.Find().SetNoCursorTimeout(true).SetBatchSize(3000)
	cur, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		log.Fatal(err)
	}
	defer cur.Close(ctx)

	var keywordIds []bson.M
	if err = cur.All(ctx, &keywordIds); err != nil {
		log.Fatal(err)
	}

	for _, keywordId := range keywordIds {
		valSet, ok := keywordId["val_set"].(primitive.A)
		if !ok {
			log.Println("val_set is not of type primitive.A")
			continue
		}
		ids := make([]string, 0, len(valSet))
		for _, v := range valSet {
			if str, ok := v.(string); ok {
				ids = append(ids, str)
			} else {
				log.Println("val_set contains non-string value")
				continue
			}
		}
		ids = utils.RemoveDuplicates(ids)
		keyword := keywordId["k"].(string)
		encryptTime := time.Now()
		//err = c.Update(keyword, ids, utils.Add)
		_, err := c.GetODXT().Encrypt(keyword, ids, utils.Add)
		if err != nil {
			log.Fatal(err)
		}
		encryptTimeList = append(encryptTimeList, time.Since(encryptTime))
		keywordList = append(keywordList, keyword)
		volumeList = append(volumeList, len(ids))
	}
	// send to server
	log.Println("Start to send to server")
	if err = c.Update(c.GetODXT().TSet, c.GetODXT().XSet); err != nil {
		log.Fatal(err)
	}
	saveTime := time.Now()
	// 设置结果文件的路径和名称
	resultpath := filepath.Join("result", "Update", "ODXT", dbName, fmt.Sprintf("%s.csv", saveTime.Format("2006-01-02_15-04-05")))

	// 定义结果表头
	resultHeader := []string{"keyword", "volume", "addTime"}

	// 将结果数据整理成表格形式
	resultData := make([][]string, len(keywordList))
	for i, keyword := range keywordList {
		resultData[i] = []string{keyword, strconv.Itoa(volumeList[i]), strconv.Itoa(int(encryptTimeList[i].Microseconds()))}
	}

	// 将结果写入文件
	err = utils.WriteResultToCSV(resultpath, resultHeader, resultData)
	if err != nil {
		log.Fatal(err)
	}
}

func SearchPhase(c *client.ODXTClient, dbName string) {
	fileName := "keywords_2.txt"
	fileName = "./cmd/ODXT/configs/" + fileName
	keywordsList := utils.QueryKeywordsFromFile(fileName)

	// 初始化结果列表
	resultList := make([][]string, 0, len(keywordsList)+1)
	resultLengthList := make([]int, 0, len(keywordsList)+1)
	totalTimeList := make([]time.Duration, 0, len(keywordsList)+1)

	// 循环搜索
	for _, keywords := range keywordsList {
		totalStart := time.Now()
		sIdList, err := c.Search(keywords)
		if err != nil {
			log.Fatal(err)
		}
		totalTime := time.Since(totalStart)

		// 将结果添加到结果列表
		resultList = append(resultList, sIdList)
		totalTimeList = append(totalTimeList, totalTime) // totalTimeList = totalTime
		resultLengthList = append(resultLengthList, len(sIdList))
	}

	// 设置结果文件的路径和名称
	resultpath := filepath.Join("result", "Search", "ODXT", dbName, fmt.Sprintf("%s.csv", time.Now().Format("2006-01-02_15-04-05")))

	// 定义结果表头
	resultHeader := []string{"keyword", "totalTime", "resultLength"}

	// 将结果数据整理成表格形式
	resultData := make([][]string, len(resultList))
	for i, keywords := range keywordsList {
		resultData[i] = []string{strings.Join(keywords, "#"), strconv.Itoa(int(totalTimeList[i].Microseconds())), strconv.Itoa(resultLengthList[i])}
	}

	// 将结果写入文件
	err := utils.WriteResultToCSV(resultpath, resultHeader, resultData)
	if err != nil {
		log.Fatal(err)
	}
}
