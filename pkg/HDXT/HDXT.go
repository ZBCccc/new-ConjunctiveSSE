package HDXT

import (
	"ConjunctiveSSE/pkg/Database"
	"ConjunctiveSSE/pkg/utils"
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"math"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	UploadListMaxLength = 200000
)

type Mitra struct {
	Key     []byte
	FileCnt map[string]int
}

type Auhme struct {
	Keys   [3][]byte
	Deltas *Delta
}

type HDXT struct {
	PlaintextDB     *mongo.Database
	MitraCipherList map[string]string
	AuhmeCipherList map[string]string
	Mitra
	Auhme
}

type MitraCipherText struct {
	Address string
	Value   string
}

type AuhmeCipherText struct {
	Label string
	Enc   string
}

var (
	universeKeywordsNums int
	universeKeywords     []string
	universeIDs          []string
	universeIDsNums      int
)

func (hdxt *HDXT) Init(dbName string, randomKey bool) error {
	// 初始化私钥
	var err error
	if randomKey {
		// 生成4个16字节长度的随机私钥
		keyLen := 16
		hdxt.Mitra.Key = make([]byte, keyLen)
		if _, err := rand.Read(hdxt.Mitra.Key); err != nil {
			log.Println("Error generating random key:", err)
			return err
		}
		for i := 0; i < 3; i++ {
			key := make([]byte, keyLen)
			if _, err := rand.Read(key); err != nil {
				log.Println("Error generating random keys:", err)
				return err
			}
			hdxt.Auhme.Keys[i] = key
		}
	} else {
		// 读取私钥	
		hdxt.Mitra.Key, hdxt.Auhme.Keys, err = utils.HdxtReadKeys("./cmd/HDXT/configs/keys.txt")
		if err != nil {
			log.Println("Error reading keys:", err)
			return err
		}
	}

	// 连接MongoDB
	hdxt.PlaintextDB, err = Database.MongoDBSetup(dbName)
	if err != nil {
		log.Println("Error initializing PlaintextDB:", err)
		return err
	}

	// 获取keyword数量
	universeKeywords, err = Database.GetUniqueValSets(hdxt.PlaintextDB)
	if err != nil {
		log.Println("Error getting universeKeywords:", err)
		return err
	}
	universeKeywords = utils.RemoveDuplicates(universeKeywords)
	universeKeywordsNums = len(universeKeywords)

	// 获取id数量
	universeIDs, err = Database.GetUniqueKs(hdxt.PlaintextDB)
	if err != nil {
		log.Println("Error getting universeIDs:", err)
		return err
	}
	universeIDs = utils.RemoveDuplicates(universeIDs)
	universeIDsNums = len(universeIDs)

	// 初始化FileCnt
	hdxt.Mitra.FileCnt = make(map[string]int, universeKeywordsNums)

	// 初始化Auhme
	hdxt.Auhme.Deltas = &Delta{cnt: 0, t: make(map[string]int), delta: 0, s: make([]string, 0)}

	hdxt.MitraCipherList = make(map[string]string)
	hdxt.AuhmeCipherList = make(map[string]string)

	return nil
}

// type volume struct {
// 	mitraVolume int
// 	auhmeVolume int
// }

func (hdxt *HDXT) SetupPhase() error {
	// 获取MongoDB数据库
	plaintextDB := hdxt.PlaintextDB
	defer plaintextDB.Client().Disconnect(context.Background())

	// 初始化
	setupTimeList := make([]time.Duration, 0, 1000000)
	// tokenList := make([]*UTok, 0, 1000000)
	

	// 从MongoDB数据库中获取名为"id_keywords"的集合
	collection := plaintextDB.Collection("id_keywords")

	// 创建一个游标，设置不超时并每次获取1000条记录
	ctx := context.TODO()
	opts := options.Find().SetNoCursorTimeout(true).SetBatchSize(1000)
	cur, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		log.Println("Error getting collection:", err)
		return err
	}

	// 关闭游标
	defer cur.Close(ctx)

	// 读取游标中的所有记录
	var idKeywords []bson.M
	if err = cur.All(ctx, &idKeywords); err != nil {
		log.Println("Error getting keywordIds:", err)
		return err
	}

	// Setup Phase
	idKeywordsSetup := idKeywords[:len(idKeywords)/2]
	idList := make([]string, 0, len(idKeywords)/2)
	for _, idKeyword := range idKeywordsSetup {
		valSet, ok := idKeyword["val_set"].(primitive.A)
		if !ok {
			log.Println("val_set is not of type primitive.A")
			return err
		}
		var keywords []string
		for _, v := range valSet {
			if str, ok := v.(string); ok {
				keywords = append(keywords, str)
			} else {
				log.Println("val_set contains non-string value")
				return err
			}
		}
		keywords = utils.RemoveDuplicates(keywords) // 对keywords去重
		id := idKeyword["k"].(string)
		encryptTime, err := hdxt.Setup(id, keywords, Add)
		if err != nil {
			log.Println("Error in Setup:", err)
			return err
		}

		setupTimeList = append(setupTimeList, encryptTime)
		idList = append(idList, id)
	}
	// save to file
	saveTime := time.Now()
	resultpath := filepath.Join("result", "Setup", "HDXT", fmt.Sprintf("%s.csv", saveTime.Format("2006-01-02_15-04-05")))
	resultHeader := []string{"id", "setupTime"}
	resultData := make([][]string, len(idList))
	for i, id := range idList {
		resultData[i] = []string{id, setupTimeList[i].String()}
	}
	err = utils.WriteResultToCSV(resultpath, resultHeader, resultData)
	if err != nil {
		log.Println("Error writing result to file:", err)
		return err
	}

	// Update Phase
	idList = make([]string, 0, len(idKeywords)/2)
	idKeywordsUpdate := idKeywords[len(idKeywords)/2:]
	for _, idKeyword := range idKeywordsUpdate {
		valSet, ok := idKeyword["val_set"].(primitive.A)
		if !ok {
			log.Println("val_set is not of type primitive.A")
			return err
		}
		var keywords []string
		for _, v := range valSet {
			if str, ok := v.(string); ok {
				keywords = append(keywords, str)
			} else {
				log.Println("val_set contains non-string value")
				return err
			}
		}
		keywords = utils.RemoveDuplicates(keywords) // 对keyword去重
		id := idKeyword["k"].(string)
		encryptTime, tokList, err := hdxt.Encrypt(id, keywords, Add)
		if err != nil {
			log.Println("Error in Encrypt:", err)
			return err
		}

		// server update
		for _, tok := range tokList {
			auhmeApplyUpd(hdxt, tok)
		}

		// save to []
		setupTimeList = append(setupTimeList, encryptTime)
		// tokenList = append(tokenList, tokList...)
		idList = append(idList, id)
	}
	saveTime = time.Now()

	// save filecnt
	if err := utils.SaveFileCntToFile(hdxt.FileCnt, "./cmd/HDXT/configs/filecnt.json"); err != nil {
		log.Println("Error saving filecnt to file:", err)
		return err
	}

	// save to file
	resultpath = filepath.Join("result", "Update", "HDXT", fmt.Sprintf("%s.csv", saveTime.Format("2006-01-02_15-04-05")))

	// 定义结果表头
	resultHeader = []string{"id", "addTime"}

	// 将结果数据整理成表格形式
	resultData = make([][]string, len(idList))
	for i, id := range idList {
		resultData[i] = []string{id, setupTimeList[i].String()}
	}

	// 将结果写入文件
	err = utils.WriteResultToCSV(resultpath, resultHeader, resultData)
	if err != nil {
		log.Println("Error writing result to file:", err)
		return err
	}

	return nil
}

func (hdxt *HDXT) Setup(id string, keywords []string, operation Operation) (time.Duration, error) {
	var encryptedTime time.Duration

	for _, keyword := range universeKeywords {
		if slices.Contains(keywords, keyword) {
			if _, ok := hdxt.FileCnt[keyword]; !ok {
				hdxt.FileCnt[keyword] = 0
			}
			start := time.Now()
			
			// OXT Part
			address, val, err := mitraEncrypt(hdxt, keyword, id, int(operation))
			if err != nil {
				log.Println(err)
				return 0, err
			}

			// Auhme Part
			label, enc, err := auhmeEncrypt(hdxt, keyword, id, 1)
			if err != nil {
				log.Println(err)
				return 0, err
			}

			encryptedTime += time.Since(start)
			hdxt.AuhmeCipherList[label] = enc
			hdxt.MitraCipherList[address] = val
		} else {
			start := time.Now()
			// Auhme Part
			label, enc, err := auhmeEncrypt(hdxt, keyword, id, 0)
			if err != nil {
				log.Println(err)
				return 0, err
			}
			encryptedTime += time.Since(start)
			hdxt.AuhmeCipherList[label] = enc
		}
	}
	return encryptedTime, nil
}

func (hdxt *HDXT) Encrypt(id string, keywords []string, operation Operation) (time.Duration, []*UTok, error) {
	tokList := make([]*UTok, 0)
	UT := make(map[string]string)
	var (
		utok *UTok
		err  error
	)
	start := time.Now()
	// op == add
	if operation == Add {
		for _, keyword := range universeKeywords {
			if slices.Contains(keywords, keyword) {
				if _, ok := hdxt.FileCnt[keyword]; !ok {
					hdxt.FileCnt[keyword] = 0
				}
				// oxt part
				address, val, err := mitraEncrypt(hdxt, keyword, id, int(operation))
				if err != nil {
					log.Println("Error in Encrypt:", err)
					return 0, nil, err
				}
				hdxt.MitraCipherList[address] = val

				// auhme part
				utok, err = auhmeGenUpd(hdxt, Add, keyword+"#"+id, 1)
				if err != nil {
					log.Println("Error in auhmeGenUpd:", err)
					return 0, nil, err
				}
			} else {
				// auhme part
				utok, err = auhmeGenUpd(hdxt, Add, keyword+"#"+id, 0)
				if err != nil {
					log.Println("Error in auhmeGenUpd:", err)
					return 0, nil, err
				}
			}
			ut := utok.tok
			for k, v := range ut {
				UT[k] = v
			}
		}
		tokx := &UTok{tok: UT, op: Add}
		tokList = append(tokList, tokx)
	} else {
		// op == edit
		for _, keyword := range keywords {
			tok, err := hdxt.EditPair(id, keyword, operation)
			if err != nil {
				log.Println("Error in Encrypt:", err)
				return 0, nil, err
			}
			if tok != nil {
				tokList = append(tokList, tok)
			}
		}
	}
	encryptedTime := time.Since(start)
	return encryptedTime, tokList, nil
}

func (hdxt *HDXT) EditPair(id, keyword string, operation Operation) (*UTok, error) {
	t, delta, s := hdxt.Auhme.Deltas.t, hdxt.Auhme.Deltas.delta, hdxt.Auhme.Deltas.s
	if len(t)+1 >= delta {
		s = make([]string, 0)
		for _, keyword := range universeKeywords {
			for _, id := range universeIDs {
				enc, err := utils.FAesni(hdxt.Auhme.Keys[0], []byte(keyword+id), 1)
				if err != nil {
					log.Println("Error in EditPair:", err)
					return nil, err
				}
				s = append(s, base64.StdEncoding.EncodeToString(enc))
			}
		}
	}
	hdxt.Auhme.Deltas.s = s
	if operation == EditPlus {
		utok, err := auhmeGenUpd(hdxt, Edit, keyword+id, 1)
		if err != nil {
			log.Println("Error in EditPair:", err)
			return nil, err
		}
		return utok, nil
	} else {
		utok, err := auhmeGenUpd(hdxt, Edit, keyword+id, 0)
		if err != nil {
			log.Println("Error in EditPair:", err)
			return nil, err
		}
		return utok, nil
	}
}

func (hdxt *HDXT) SearchPhase(tableName, fileName string) {
	fileName = "./cmd/HDXT/configs/" + fileName
	keywordsList := utils.QueryKeywordsFromFile(fileName)

	// 初始化结果列表
	resultList := make([][]string, 0, len(keywordsList)+1)
	clientSearchTime := make([]time.Duration, 0, len(keywordsList)+1)
	serverTimeList := make([]time.Duration, 0, len(keywordsList)+1)
	resultLengthList := make([]int, 0, len(keywordsList)+1)

	clientTimeTotal := time.Duration(0)
	serverTimeTotal := time.Duration(0)

	// 循环搜索
	// keywordsList = keywordsList[:1]
	for _, keywords := range keywordsList {
		// 单关键词搜索, mitra part
		// 选择查询频率最低的关键字
		counter, w1 := math.MaxInt64, keywords[0]
		for _, w := range keywords {
			num := hdxt.FileCnt[w]
			if num < counter {
				w1 = w
				counter = num
			}
		}
		trapdoorTime, serverTime, w1Ids, err := hdxt.SearchOneKeyword(w1)
		if err != nil {
			log.Fatal(err)
		}
		clientTimeTotal += trapdoorTime
		serverTimeTotal += serverTime

		// auhme part
		// clien search step 1
		q := utils.RemoveElement(keywords, w1)
		start := time.Now()
		dkList, err := auhmeClientSearchStep1(hdxt, w1Ids, q)
		if err != nil {
			log.Fatal(err)
		}
		clientTimeTotal += time.Since(start)
		// fmt.Println("dkList's length:", len(dkList))

		// server search step
		start = time.Now()
		posList := auhmeServerSearch(hdxt, dkList)
		serverTimeTotal += time.Since(start)
		// fmt.Println("posList's length:", len(posList))

		// client search step 2
		start = time.Now()
		sIdList := auhmeClientSearchStep2(w1Ids, posList)
		clientTimeTotal += time.Since(start)

		// 将结果添加到结果列表
		resultList = append(resultList, sIdList)
		clientSearchTime = append(clientSearchTime, clientTimeTotal)
		serverTimeList = append(serverTimeList, serverTimeTotal)
		resultLengthList = append(resultLengthList, len(sIdList))
	}

	// 设置结果文件的路径和名称
	resultpath := filepath.Join("result", "Search", "HDXT", tableName, fmt.Sprintf("%s.csv", time.Now().Format("2006-01-02_15-04-05")))

	// 定义结果表头
	resultHeader := []string{"keyword", "clientTime", "serverTime", "resultLength"}

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

func (hdxt *HDXT) SearchOneKeyword(keyword string) (time.Duration, time.Duration, []string, error) {
	// 生成陷门
	start := time.Now()
	tList, err := mitraGenTrapdoor(hdxt, keyword)
	if err != nil {
		log.Println(err)
		return 0, 0, nil, nil
	}
	clientTime := time.Since(start)
	// server search
	start = time.Now()
	encryptedIds := mitraServerSearch(hdxt, tList)
	serverTime := time.Since(start)

	// client decrypt and return result
	start = time.Now()
	ids, err := mitraDecrypt(hdxt, keyword, encryptedIds)
	if err != nil {
		log.Println(err)
		return 0, 0, nil, nil
	}
	clientTime += time.Since(start)

	return clientTime, serverTime, ids, nil
}
