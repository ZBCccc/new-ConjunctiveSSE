package main

import (
	"ConjunctiveSSE/pkg/Database"
	"ConjunctiveSSE/pkg/FDXT"
	"ConjunctiveSSE/pkg/FDXT/client"
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
	c, err := client.NewFDXTClient("localhost:50053")
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

func UpdatePhase(c *client.FDXTClient, dbName string) {
	PlaintextDB, err := Database.MongoDBSetup(dbName)
	if err != nil {
		log.Fatal(err)
	}
	defer PlaintextDB.Client().Disconnect(context.Background())

	// Update Phase
	// 从MongoDB数据库中获取名为"keyword_ids"的集合
	collection := PlaintextDB.Collection("keyword_ids")

	// 创建一个游标，设置不超时并每次获取3000条记录
	ctx := context.TODO()
	opts := options.Find().SetNoCursorTimeout(true).SetBatchSize(3000)
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

	// 读取所有记录
	encryptTimeList := make([]time.Duration, 0, 1000000)
	cipherList := make([]int, 0, 1000000)
	for _, keywordId := range keywordIds {
		valSet, ok := keywordId["val_set"].(primitive.A)
		if !ok {
			log.Fatal("val_set is not of type primitive.A")
		}
		var ids []string
		for _, v := range valSet {
			if str, ok := v.(string); ok {
				ids = append(ids, str)
			} else {
				log.Fatal("val_set contains non-string value")
			}
		}
		ids = utils.RemoveDuplicates(ids)
		keyword := keywordId["k"].(string)

		encryptTime, err := c.GetFDXT().Encrypt(keyword, ids, FDXT.Add)
		if err != nil {
			log.Println(err)
		}
		encryptTimeList = append(encryptTimeList, encryptTime)
		cipherList = append(cipherList, len(ids))
	}
	// send to server
	log.Println("Start to send to server")
	if err = c.Update(c.GetFDXT().CDBXtag, c.GetFDXT().CDBTSet); err != nil {
		log.Fatal(err)
	}
	// save to file
	saveTime := time.Now()
	resultpath := filepath.Join("result", "Update", "FDXT", fmt.Sprintf("%s.csv", saveTime.Format("2006-01-02_15-04-05")))
	resultHeader := []string{"encryptTime", "cipherLength"}
	resultData := make([][]string, len(cipherList))
	for i, cipher := range cipherList {
		resultData[i] = []string{encryptTimeList[i].String(), strconv.Itoa(cipher)}
	}
	err = utils.WriteResultToCSV(resultpath, resultHeader, resultData)
	if err != nil {
		log.Fatal("Error writing result to file:", err)
	}
}

func SearchPhase(c *client.FDXTClient, dbName string) {
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
	resultpath := filepath.Join("result", "Search", "FDXT", dbName, fmt.Sprintf("%s.csv", time.Now().Format("2006-01-02_15-04-05")))

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