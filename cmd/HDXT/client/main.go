package main

import (
	"ConjunctiveSSE/pkg/HDXT"
	"ConjunctiveSSE/pkg/HDXT/client"
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
    c, err := client.NewHDXTClient("localhost:50051")
    if err != nil {
        log.Fatalf("failed to create client: %v", err)
    }
    
    // 执行实验
	// 获取MongoDB数据库
	hdxt := c.GetHDXT()
	plaintextDB := hdxt.PlaintextDB
	defer plaintextDB.Client().Disconnect(context.Background())

	// 初始化
	setupTimeList := make([]time.Duration, 0, 1000000)

	// 从MongoDB数据库中获取名为"id_keywords"的集合
	collection := plaintextDB.Collection("id_keywords")

	// 创建一个游标，设置不超时并每次获取3000条记录
	ctx := context.TODO()
	opts := options.Find().SetNoCursorTimeout(true).SetBatchSize(3000)
	cur, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		log.Fatal("Error getting collection:", err)
	}

	// 关闭游标
	defer cur.Close(ctx)

	// 读取游标中的所有记录
	var idKeywords []bson.M
	if err = cur.All(ctx, &idKeywords); err != nil {
		log.Fatal("Error getting keywordIds:", err)
	}
    // 1. Setup Phase
	log.Println("Setup Phase Start")
	setupTime := time.Now()
	idKeywordsSetup := idKeywords[:len(idKeywords)/2]
	idList := make([]string, 0, len(idKeywords)/2)
	volumeList := make([]int, 0, len(idKeywords)/2)
	for _, idKeyword := range idKeywordsSetup {
		valSet, ok := idKeyword["val_st"].(primitive.A)
		if !ok {
			log.Fatal("val_set is not of type primitive.A")
		}
		var keywords []string
		for _, v := range valSet {
			if str, ok := v.(string); ok {
				keywords = append(keywords, str)
			} else {
				log.Fatal("val_set contains non-string value")
			}
		}
		keywords = utils.RemoveDuplicates(keywords) // 对keywords去重
		id := idKeyword["id"].(string)
		encryptTime := time.Now()
		_, err = hdxt.Setup(id, keywords, HDXT.Add)
		if err != nil {
			log.Fatal("Error in Setup:", err)
		}

		setupTimeList = append(setupTimeList, time.Since(encryptTime))
		idList = append(idList, id)
		volumeList = append(volumeList, len(keywords))
	}
	// 发送 Setup 请求
	log.Println("Sending Setup Request, cost time:", time.Since(setupTime))
	c.Setup(hdxt.MitraCipherList, hdxt.AuhmeCipherList)
	// save to file
	saveTime := time.Now()
	resultpath := filepath.Join("result", "Setup", "HDXT", fmt.Sprintf("%s.csv", saveTime.Format("2006-01-02_15-04-05")))
	resultHeader := []string{"id", "volume", "addTime"}
	resultData := make([][]string, len(idList))
	for i, id := range idList {
		resultData[i] = []string{id, strconv.Itoa(volumeList[i]), strconv.Itoa(int(setupTimeList[i].Microseconds()))}
	}
	err = utils.WriteResultToCSV(resultpath, resultHeader, resultData)
	if err != nil {
		log.Fatal("Error writing result to file:", err)
	}

	// 2. Update Phase
	log.Println("Update Phase Start")
	idList = make([]string, 0, len(idKeywords)/2)
	idKeywordsUpdate := idKeywords[len(idKeywords)/2:]
	volumeList = make([]int, 0, len(idKeywords)/2)
	for _, idKeyword := range idKeywordsUpdate {
		valSet, ok := idKeyword["val_st"].(primitive.A)
		if !ok {
			log.Fatal("val_set is not of type primitive.A")
		}
		var keywords []string
		for _, v := range valSet {
			if str, ok := v.(string); ok {
				keywords = append(keywords, str)
			} else {
				log.Fatal("val_set contains non-string value")
			}
		}
		keywords = utils.RemoveDuplicates(keywords) // 对keyword去重
		id := idKeyword["id"].(string)
		encryptTime := time.Now()
		err = c.Update(id, keywords, HDXT.Add)
		if err != nil {
			log.Fatal("Error in Encrypt:", err)
		}

		// save to []
		setupTimeList = append(setupTimeList, time.Since(encryptTime))
		idList = append(idList, id)
		volumeList = append(volumeList, len(keywords))
	}
	saveTime = time.Now()

	// save to file
	resultpath = filepath.Join("result", "Update", "HDXT", fmt.Sprintf("%s.csv", saveTime.Format("2006-01-02_15-04-05")))

	// 定义结果表头
	resultHeader = []string{"id", "volume", "addTime"}

	// 将结果数据整理成表格形式
	resultData = make([][]string, len(idList))
	for i, id := range idList {
		resultData[i] = []string{id, strconv.Itoa(volumeList[i]), strconv.Itoa(int(setupTimeList[i].Microseconds()))}
	}

	// 将结果写入文件
	err = utils.WriteResultToCSV(resultpath, resultHeader, resultData)
	if err != nil {
		log.Fatal("Error writing result to file:", err)
	}

	// 3. Search Phase
	log.Println("Search Phase Start")
	SearchPhase(c, "Crime_USENIX_REV", "keywords_2.txt")
}

func SearchPhase(c *client.HDXTClient, tableName, fileName string) {
	fileName = "./cmd/HDXT/configs/" + fileName
	keywordsList := utils.QueryKeywordsFromFile(fileName)
	// 初始化结果列表
	clientSearchTime := make([]time.Duration, 0, len(keywordsList)+1)
	resultLengthList := make([]int, 0, len(keywordsList)+1)

	// 循环搜索
	for _, keywords := range keywordsList {
		searchTime := time.Now()
		sIdList, err := c.Search(keywords)
		if err != nil {
			log.Fatal("Error in Search:", err)
		}

		// 将结果添加到结果列表
		clientSearchTime = append(clientSearchTime, time.Since(searchTime))
		resultLengthList = append(resultLengthList, len(sIdList))
	}
	log.Println("Search Phase End")

	// 设置结果文件的路径和名称
	resultpath := filepath.Join("result", "Search", "HDXT", tableName, fmt.Sprintf("%s.csv", time.Now().Format("2006-01-02_15-04-05")))

	// 定义结果表头
	resultHeader := []string{"keyword", "clientTime", "resultLength"}

	// 将结果数据整理成表格形式
	resultData := make([][]string, len(keywordsList))
	for i, keywords := range keywordsList {
		resultData[i] = []string{strings.Join(keywords, "#"), strconv.Itoa(int(clientSearchTime[i].Microseconds())), strconv.Itoa(resultLengthList[i])}
	}

	// 将结果写入文件
	err := utils.WriteResultToCSV(resultpath, resultHeader, resultData)
	if err != nil {
		log.Fatal(err)
	}
}