package ODXT

import (
	"ConjunctiveSSE/pkg/Database"
	"ConjunctiveSSE/pkg/utils"
	"bufio"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/mongo"
)

type tsetValue struct {
	val   string
	alpha *big.Int
}

var (
	g           *big.Int
	p           *big.Int
	pMinusOne   *big.Int
	PlaintextDB *mongo.Database
)

const (
	keysPath = "./cmd/ODXT/configs/keys.txt"
)

type ODXT struct {
	Keys      [4][]byte
	UpdateCnt map[string]int
	TSet      map[string]*tsetValue
	XSet      map[string]int
}

type UpdatePayload struct {
	Address string
	Val     string
	Alpha   string
}

func ReadKeys(fileName string) [4][]byte {
	// 读取文件
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// 读取密钥
	var keys [4][]byte
	scanner := bufio.NewScanner(file)

	// 读取4个密钥
	for i := 0; i < 4; i++ {
		line := scanner.Text()
		key, err := base64.StdEncoding.DecodeString(line)
		if err != nil {
			log.Fatal(err)
		}

		keys[i] = key
	}

	return keys
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
		odxt.Keys = ReadKeys(keysPath)
	}

	// 初始化 UpdateCnt
	odxt.UpdateCnt = make(map[string]int)

	// 初始化 g 和 p
	g = big.NewInt(65537)
	p, _ = new(big.Int).SetString("69445180235231407255137142482031499329548634082242122837872648805446522657159", 10)
	pMinusOne = new(big.Int).Sub(p, big.NewInt(1))
	// 初始化 XSet 和 TSet
	odxt.XSet = make(map[string]int)
	odxt.TSet = make(map[string]*tsetValue)

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
		ids = utils.RemoveDuplicates(ids)
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

	// 设置结果文件的路径和名称
	resultpath := filepath.Join("result", "Update", "ODXT", dbName, fmt.Sprintf("%s.csv", saveTime.Format("2006-01-02_15-04-05")))

	// 定义结果表头
	resultHeader := []string{"keyword", "volume", "addTime"}

	// 将结果数据整理成表格形式
	resultData := make([][]string, len(keywordList))
	for i, keyword := range keywordList {
		resultData[i] = []string{keyword, strconv.Itoa(volumeList[i]), encryptTimeList[i].String()}
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
	clientSearchTime := make([]time.Duration, 0, len(keywordsList)+1)
	serverTimeList := make([]time.Duration, 0, len(keywordsList)+1)
	resultLengthList := make([]int, 0, len(keywordsList)+1)
	clientTimeTotal := time.Duration(0)
	serverTimeTotal := time.Duration(0)
	counterList := make([]int, 0, len(keywordsList)+1)

	// 循环搜索
	for _, keywords := range keywordsList {
		// find w1's lens
		// 选择查询频率最低的关键字
		counter := math.MaxInt
		for _, w := range keywords {
			num := odxt.UpdateCnt[w]
			if num < counter {
				counter = num
			}
		}
		counterList = append(counterList, counter)
		trapdoorTime, serverTime, sEOpList := odxt.Search(keywords)

		// 解密密文获得最终结果
		start := time.Now()
		sIdList, err := odxt.Decrypt(keywords, sEOpList)
		if err != nil {
			log.Fatal(err)
		}
		decryptTime := time.Since(start)
		clientTime := trapdoorTime + decryptTime
		clientTimeTotal += clientTime
		serverTimeTotal += serverTime

		// 将结果添加到结果列表
		resultList = append(resultList, sIdList)
		clientSearchTime = append(clientSearchTime, clientTime)
		serverTimeList = append(serverTimeList, serverTime)
		resultLengthList = append(resultLengthList, len(sIdList))
	}

	// 设置结果文件的路径和名称
	resultpath := filepath.Join("result", "Search", "ODXT", tableName, fmt.Sprintf("%s.csv", time.Now().Format("2006-01-02_15-04-05")))

	// 定义结果表头
	resultHeader := []string{"keyword", "clientSearchTime", "serverTime", "resultLength", "counter"}

	// 将结果数据整理成表格形式
	resultData := make([][]string, len(resultList))
	for i, keywords := range keywordsList {
		resultData[i] = []string{strings.Join(keywords, "#"), clientSearchTime[i].String(), serverTimeList[i].String(), strconv.Itoa(resultLengthList[i]), strconv.Itoa(counterList[i])}
	}

	// 将结果写入文件
	err := utils.WriteResultToCSV(resultpath, resultHeader, resultData)
	if err != nil {
		log.Fatal(err)
	}
}

// CalculateUpdatePayloadSize 计算[]UpdatePayload的字节大小
func CalculateUpdatePayloadSize(payloads []UpdatePayload) int {
	size := 0
	for _, payload := range payloads {
		size += len(payload.Address) + len(payload.Val) + len(payload.Alpha)
	}
	return size
}
