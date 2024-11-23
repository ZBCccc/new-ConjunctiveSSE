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
	PlaintextDB *mongo.Database
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

type SEOp struct {
	J    int
	Sval []byte
	Cnt  int
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
		odxt.Keys = ReadKeys("./cmd/ODXT/configs/keys.txt")
	}

	// 初始化 UpdateCnt
	odxt.UpdateCnt = make(map[string]int)

	// 初始化 g 和 p
	g = big.NewInt(65537)
	p, _ = new(big.Int).SetString("69445180235231407255137142482031499329548634082242122837872648805446522657159", 10)

	// 初始化 XSet 和 TSet
	odxt.XSet = make(map[string]int)
	odxt.TSet = make(map[string]*tsetValue)

	// 连接MongoDB
	var err error
	PlaintextDB, err = Database.MongoDBSetup(dbName)
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func (odxt *ODXT) CiphertextGenPhase(dbName string) {
	// 获取MongoDB数据库
	defer PlaintextDB.Client().Disconnect(context.Background())

	// 初始化
	encryptTimeList := make([]time.Duration, 0, 1000000)
	keywordList := make([]string, 0, 1000000)
	volumeList := make([]int, 0, 1000000)

	// 从MongoDB数据库中获取名为"id_keywords"的集合
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
	var keywordIds []bson.M
	if err = cur.All(ctx, &keywordIds); err != nil {
		log.Fatal(err)
	}

	// 读取所有记录
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

		encryptTime, err := odxt.Encrypt(keyword, ids, utils.Add)
		if err != nil {
			log.Fatal(err)
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
		log.Fatal(err)
	}
}

func (odxt *ODXT) DeletionPhaseWithSearch(del_rate int) {

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

	// 循环搜索
	for _, keywords := range keywordsList {
		trapdoorTime, serverTime, sEOpList := odxt.Search(keywords, tableName)

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
		clientSearchTime = append(clientSearchTime, clientTimeTotal)
		serverTimeList = append(serverTimeList, serverTimeTotal)
		resultLengthList = append(resultLengthList, len(sIdList))
	}

	// 设置结果文件的路径和名称
	resultpath := filepath.Join("result", "Search", "ODXT", tableName, fmt.Sprintf("%s.csv", time.Now().Format("2006-01-02_15-04-05")))

	// 定义结果表头
	resultHeader := []string{"keyword", "clientSearchTime", "serverTime", "resultLength"}

	// 将结果数据整理成表格形式
	resultData := make([][]string, len(resultList))
	for i, keywords := range keywordsList {
		resultData[i] = []string{strings.Join(keywords, "#"), clientSearchTime[i].String(), serverTimeList[i].String(), strconv.Itoa(resultLengthList[i])}
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
