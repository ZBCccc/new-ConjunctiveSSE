package ODXT

import (
	"ConjunctiveSSE/pkg/Database"
	"ConjunctiveSSE/pkg/utils"
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Nik-U/pbc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/mongo"
)

type TsetValue struct {
	Val   string
	Alpha *pbc.Element
}

var (
	PlaintextDB *mongo.Database
)

type ODXT struct {
	Keys      [4][]byte
	UpdateCnt map[string]int
	TSet      map[string]*TsetValue
	XSet      map[string]int
}

func (odxt *ODXT) DBSetup(dbName string, randomKey bool) error {
	if randomKey {
		// 生成4个32字节长度的随机私钥
		keyLen := 32
		for i := 0; i < 4; i++ {
			key := make([]byte, keyLen)
			if _, err := rand.Read(key); err != nil {
				log.Println("Error generating random key:", err)
				return err
			}
			odxt.Keys[i] = key
		}
	} else {
		// 读取私钥
		odxt.Keys[0] = []byte("0123456789123456")
		odxt.Keys[1] = []byte("0123456789123456")
		odxt.Keys[2] = []byte("0123456789123456")
		odxt.Keys[3] = []byte("0123456789123456")
	}

	// 初始化 UpdateCnt
	odxt.UpdateCnt = make(map[string]int, 1000000)

	// 初始化 XSet 和 TSet
	odxt.XSet = make(map[string]int, 1000000)
	odxt.TSet = make(map[string]*TsetValue, 1000000)

	// 连接MongoDB
	var err error
	PlaintextDB, err = Database.MongoDBSetup(dbName)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (odxt *ODXT) CiphertextGenPhase(dbName string) error {
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
		keyword := keywordId["k"].(string)

		encryptTime, err := odxt.Encrypt(keyword, ids, utils.Add)
		if err != nil {
			log.Println(err)
			continue
		}

		encryptTimeList = append(encryptTimeList, encryptTime)
		keywordList = append(keywordList, keyword)
		volumeList = append(volumeList, len(ids))
	}
	saveTime := time.Now()
	// save filecnt
	if err := utils.SaveFileCntToFile(odxt.UpdateCnt, "./cmd/ODXT/configs/filecnt.json"); err != nil {
		log.Println("Error saving filecnt to file:", err)
		return err
	}

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
		log.Println(err)
		return err
	}
	return nil
}

func (odxt *ODXT) DeletionPhaseWithSearch(delRate int) {
	fileName := "./cmd/ODXT/configs/del_keywords.txt"
	delNums := int(float64(len(odxt.TSet)) * (float64(delRate) / 100.0))
	keywordsList := utils.GenDeletePairs(fileName, delNums)
	_ = keywordsList
}

func (odxt *ODXT) SearchPhase(tableName, fileName string) {
	fileName = "./cmd/ODXT/configs/" + fileName
	keywordsList := utils.QueryKeywordsFromFile(fileName)

	// 初始化结果列表
	resultList := make([][]string, 0, len(keywordsList)+1)
	clientTimeList := make([]time.Duration, 0, len(keywordsList)+1)
	serverTimeList := make([]time.Duration, 0, len(keywordsList)+1)
	resultLengthList := make([]int, 0, len(keywordsList)+1)
	w1CounterList := make([]int, 0, len(keywordsList)+1)
	w2CounterList := make([]int, 0, len(keywordsList)+1)
	totalTimeList := make([]time.Duration, 0, len(keywordsList)+1)
	payloadSizeList := make([]int, 0, len(keywordsList)+1)

	// 循环搜索
	for _, keywords := range keywordsList {
		totalStart := time.Now()
		// find w1's lens
		// 选择查询频率最低的关键字
		counter := math.MaxInt
		var w1 string
		for _, w := range keywords {
			num := odxt.UpdateCnt[w]
			if num < counter {
				counter = num
				w1 = w
			}
		}

		// 生成陷门
		start := time.Now()
		stokenList, xtokenList := odxt.ClientSearchStep1(w1, keywords)
		trapdoorTime := time.Since(start)

		// 服务器检索
		start = time.Now()
		sEOpList := odxt.Search(stokenList, xtokenList)
		serverTime := time.Since(start)

		// 解密密文获得最终结果
		start = time.Now()
		sIdList := odxt.ClientSearchStep2(w1, keywords, sEOpList)
		decryptTime := time.Since(start)

		clientTime := trapdoorTime + decryptTime
		totalTime := time.Since(totalStart)

		// 将结果添加到结果列表
		w1CounterList = append(w1CounterList, counter)
		w2CounterList = append(w2CounterList, odxt.UpdateCnt[keywords[1]])
		resultList = append(resultList, sIdList)
		clientTimeList = append(clientTimeList, clientTime) // clientTimeList = trapdoorTime + decryptTime
		serverTimeList = append(serverTimeList, serverTime) // serverTimeList = serverTime
		totalTimeList = append(totalTimeList, totalTime)    // totalTimeList = totalTime
		resultLengthList = append(resultLengthList, len(sIdList))
		payloadSizeList = append(payloadSizeList, utils.CalculatePayloadSize(sEOpList))
	}

	// 设置结果文件的路径和名称
	resultpath := filepath.Join("result", "Search", "ODXT", tableName, time.Now().Format("2006-01-02_15-04")+"_w2_keywords_2.csv")

	// 定义结果表头
	resultHeader := []string{"keyword", "clientTime", "serverTime", "totalTime", "resultLength", "payloadSize", "w1", "w2"}

	// 将结果数据整理成表格形式
	resultData := make([][]string, len(resultList))
	for i, keywords := range keywordsList {
		resultData[i] = []string{strings.Join(keywords, "#"), strconv.Itoa(int(clientTimeList[i].Microseconds())), strconv.Itoa(int(serverTimeList[i].Microseconds())), strconv.Itoa(int(totalTimeList[i].Microseconds())), strconv.Itoa(resultLengthList[i]), strconv.Itoa(payloadSizeList[i]), strconv.Itoa(w1CounterList[i]), strconv.Itoa(w2CounterList[i])}
	}

	// 将结果写入文件
	err := utils.WriteResultToCSV(resultpath, resultHeader, resultData)
	if err != nil {
		log.Fatal(err)
	}
}
