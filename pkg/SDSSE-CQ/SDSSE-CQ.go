package sdssecq

import (
	"ConjunctiveSSE/pkg/Database"
	sdssecqClient "ConjunctiveSSE/pkg/SDSSE-CQ/Client"
	"ConjunctiveSSE/pkg/utils"
	"context"
	"fmt"
	util "github.com/ZBCccc/Aura/Util"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	// PlaintextDB is the MongoDB database for storing plaintext data.
	PlaintextDB *mongo.Database
	client      *sdssecqClient.Client
)

func Init(dbName string) error {
	// 初始化Client
	client = sdssecqClient.NewClient()

	// 连接MongoDB
	var err error
	PlaintextDB, err = Database.MongoDBSetup(dbName)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func CiphertextGenPhase(dbName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	defer PlaintextDB.Client().Disconnect(ctx)

	collection := PlaintextDB.Collection("keyword_ids")
	// 先获取总数以预分配空间
	count, err := collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("count documents failed: %w", err)
	}

	// 使用实际数量预分配
	encryptTimeList := make([]time.Duration, 0, count)
	keywordList := make([]string, 0, count)
	volumeList := make([]int, 0, count)

	opts := options.Find().SetNoCursorTimeout(true).SetBatchSize(3000)
	cur, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return fmt.Errorf("find collection failed: %w", err)
	}
	defer cur.Close(ctx)

	var keywordIds []bson.M
	if err = cur.All(ctx, &keywordIds); err != nil {
		return fmt.Errorf("read cursor failed: %w", err)
	}

	// 读取所有记录
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
		start := time.Now()
		for _, id := range ids {
			client.Update(util.Insert, keyword, id)
		}
		encryptTime := time.Since(start)

		encryptTimeList = append(encryptTimeList, encryptTime)
		keywordList = append(keywordList, keyword)
		volumeList = append(volumeList, len(ids))
	}
	saveTime := time.Now()

	// 设置结果文件的路径和名称
	resultPath := filepath.Join("result", "Update", "SDSSE-CQ", dbName, fmt.Sprintf("%s.csv", saveTime.Format("2006-01-02_15-04-05")))

	// 定义结果表头
	resultHeader := []string{"keyword", "volume", "addTime"}

	// 将结果数据整理成表格形式
	resultData := make([][]string, len(keywordList))
	for i, keyword := range keywordList {
		resultData[i] = []string{keyword, strconv.Itoa(volumeList[i]), strconv.Itoa(int(encryptTimeList[i].Microseconds()))}
	}

	// 将结果写入文件
	err = utils.WriteResultToCSV(resultPath, resultHeader, resultData)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func SearchPhase(tableName, fileName string) {
	fileName = "./cmd/SDSSE-CQ/configs/" + fileName
	keywordsList := utils.QueryKeywordsFromFile(fileName)

	// 初始化结果列表
	resultList := make([][]string, 0, len(keywordsList)+1)
	clientTimeList := make([]time.Duration, 0, len(keywordsList)+1)
	serverTimeList := make([]time.Duration, 0, len(keywordsList)+1)
	resultLengthList := make([]int, 0, len(keywordsList)+1)
	totalTimeList := make([]time.Duration, 0, len(keywordsList)+1)
	w1CounterList := make([]int, 0, len(keywordsList)+1)
	w2CounterList := make([]int, 0, len(keywordsList)+1)
	// payloadSizeList := make([]int, 0, len(keywordsList)+1)

	// 循环搜索
	for _, keywords := range keywordsList {
		totalStart := time.Now()
		// find w1's lens
		// 选择查询频率最低的关键字
		counter := math.MaxInt
		for _, w := range keywords {
			num := client.CT[w]
			if num < counter {
				counter = num
			}
		}
		w1CounterList = append(w1CounterList, client.CT[keywords[0]] + 1)
		w2CounterList = append(w2CounterList, client.CT[keywords[1]] + 1)
		result, clientTime, serverTime := client.Search(keywords)
		totalTime := time.Since(totalStart)

		// 将结果添加到结果列表
		resultList = append(resultList, result)
		resultLengthList = append(resultLengthList, len(result))
		totalTimeList = append(totalTimeList, totalTime)
		clientTimeList = append(clientTimeList, clientTime)
		serverTimeList = append(serverTimeList, serverTime)
	}

	// 设置结果文件的路径和名称
	resultPath := filepath.Join("result", "Search", "SDSSE-CQ", tableName, "w1_keywords_2")

	// 定义结果表头
	resultHeader := []string{"keyword", "clientTime", "serverTime", "totalTime", "resultLength", "w1", "w2"}

	// 将结果数据整理成表格形式
	resultData := make([][]string, len(resultList))
	for i, keywords := range keywordsList {
		resultData[i] = []string{strings.Join(keywords, "#"), strconv.Itoa(int(clientTimeList[i].Microseconds())), strconv.Itoa(int(serverTimeList[i].Microseconds())), strconv.Itoa(int(totalTimeList[i].Microseconds())), strconv.Itoa(resultLengthList[i]), strconv.Itoa(w1CounterList[i]), strconv.Itoa(w2CounterList[i])}
	}

	// 将结果写入文件
	err := utils.WriteResultToCSV(resultPath, resultHeader, resultData)
	if err != nil {
		log.Fatal(err)
	}
}
