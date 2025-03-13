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
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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
)

func (hdxt *HDXT) Init(dbName string, randomKey bool) error {
	// 初始化私钥
	var err error
	if randomKey {
		// 生成4个16字节长度的随机私钥
		keyLen := 16
		hdxt.Mitra.Key = make([]byte, keyLen)
		if _, err := rand.Read(hdxt.Mitra.Key); err != nil {
			log.Fatal("Error generating random key:", err)
		}
		for i := 0; i < 3; i++ {
			key := make([]byte, keyLen)
			if _, err := rand.Read(key); err != nil {
				log.Fatal("Error generating random keys:", err)
			}
			hdxt.Auhme.Keys[i] = key
		}
	} else {
		// 读取私钥
		// hdxt.Mitra.Key, hdxt.Auhme.Keys, err = utils.HdxtReadKeys("./cmd/HDXT/configs/keys.txt")
		// if err != nil {
		// 	log.Fatal("Error reading keys:", err)
		// }
		hdxt.Mitra.Key = []byte("0123456789123456")
		hdxt.Auhme.Keys[0] = []byte("0123456789123456")
		hdxt.Auhme.Keys[1] = []byte("0123456789123456")
		hdxt.Auhme.Keys[2] = []byte("0123456789123456")
	}

	// 连接MongoDB
	hdxt.PlaintextDB, err = Database.MongoDBSetup(dbName)
	if err != nil {
		log.Fatal("Error initializing PlaintextDB:", err)
	}

	// 获取keyword数量
	universeKeywords, err = Database.GetUniqueKeywords(hdxt.PlaintextDB)
	if err != nil {
		log.Fatal("Error getting universeKeywords:", err)
	}
	universeKeywords = utils.RemoveDuplicates(universeKeywords)
	universeKeywordsNums = len(universeKeywords)

	// 获取id数量
	universeIDs, err = Database.GetUniqueIDs(hdxt.PlaintextDB)
	if err != nil {
		log.Fatal("Error getting universeIDs:", err)
	}
	universeIDs = utils.RemoveDuplicates(universeIDs)

	// 初始化FileCnt
	hdxt.Mitra.FileCnt = make(map[string]int, universeKeywordsNums)

	// 初始化Auhme
	hdxt.Auhme.Deltas = &Delta{cnt: 0, t: make(map[string]int), delta: 0, s: make([]string, 0)}

	hdxt.MitraCipherList = make(map[string]string)
	hdxt.AuhmeCipherList = make(map[string]string)

	setupTimeList = make([]time.Duration, 0, 1000000)
	idList = make([]string, 0, 1000000)
	volumeList = make([]int, 0, 1000000)

	return nil
}

// 替代一次性加载全部数据
func loadDataInBatches(collection *mongo.Collection, sizeIdKeywords int64, hdxt *HDXT) {
    batchSize := 1000
    var skip int64 = 0
    
    for {
        // 查询一批数据
        opts := options.Find().SetLimit(int64(batchSize)).SetSkip(skip)
        cursor, err := collection.Find(context.Background(), bson.M{}, opts)
        if err != nil {
            log.Fatal(err)
        }
        
        // 处理这批数据
        var idKeywords []bson.M
		if err = cursor.All(context.Background(), &idKeywords); err != nil {
			log.Fatal("Error getting keywordIds:", err)
		}
        
        // 如果没有更多数据，退出循环
        if len(idKeywords) == 0 {
			saveData(idList, volumeList, setupTimeList)

            break
        }
        
        // 分情况处理数据，当小于 1/2 时，进行setup，否则进行update
        if skip < sizeIdKeywords/2 {
            setupBatch(idKeywords, hdxt)
        } else {
            updateBatch(idKeywords, hdxt)
        }
        
        // 更新skip以获取下一批
        skip += int64(batchSize)
        
        // 可选：手动触发GC
        runtime.GC()
    }
}

func updateBatch(idKeywords []bson.M, hdxt *HDXT) {
	for _, idKeyword := range idKeywords {
		valSet, ok := idKeyword["keywords"].(primitive.A)
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
		
		id := idKeyword["id"].(string)
		encryptTime, tokList, err := hdxt.Encrypt(id, keywords, Add)
		if err != nil {
			log.Fatal("Error in Encrypt:", err)
		}

		// server update
		for _, tok := range tokList {
			auhmeApplyUpd(hdxt, tok)
		}

		// save to []
		setupTimeList = append(setupTimeList, encryptTime)
		idList = append(idList, id)
		volumeList = append(volumeList, len(keywords))
	}
}

func saveData(idList []string, volumeList []int, setupTimeList []time.Duration) {
	// save to file
	saveTime := time.Now()
	resultpath := filepath.Join("result", "Setup", "HDXT", fmt.Sprintf("%s.csv", saveTime.Format("2006-01-02_15-04-05")))
	resultHeader := []string{"id", "volume", "addTime"}
	resultData := make([][]string, len(idList))
	for i, id := range idList {
		resultData[i] = []string{id, strconv.Itoa(volumeList[i]), strconv.Itoa(int(setupTimeList[i].Microseconds()))}
	}
	err := utils.WriteResultToCSV(resultpath, resultHeader, resultData)
	if err != nil {
		log.Fatal("Error writing result to file:", err)
	}
}

var (
	setupTimeList []time.Duration
	idList        []string
	volumeList    []int
)

func setupBatch(batch []bson.M, hdxt *HDXT) {
	for _, idKeyword := range batch {
		valSet, ok := idKeyword["keywords"].(primitive.A)
		if !ok {
			log.Fatal("val_set is not of type primitive.A")
		}
		var keywords []string
		for _, v := range valSet {
			if str, ok := v.(string); ok {
				keywords = append(keywords, str)
			} else {
				log.Fatal("关键词列表中包含非字符串类型的值")
			}
		}
		id := idKeyword["id"].(string)
		encryptTime, err := hdxt.Setup(id, keywords, Add)
		if err != nil {
			log.Fatal("Error in Setup:", err)
		}

		setupTimeList = append(setupTimeList, encryptTime)
		idList = append(idList, id)
		volumeList = append(volumeList, len(keywords))
	}
}

func (hdxt *HDXT) SetupPhase() error {
	// 获取MongoDB数据库
	plaintextDB := hdxt.PlaintextDB
	defer plaintextDB.Client().Disconnect(context.Background())
	
	// 从MongoDB数据库中获取名为"id_keywords"的集合
	collection := plaintextDB.Collection("id_keywords")
	sizeIdKeywords, _ := collection.CountDocuments(context.Background(), bson.M{})

	loadDataInBatches(collection, sizeIdKeywords, hdxt)

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
			ut := utok.Tok
			for k, v := range ut {
				UT[k] = v
			}
		}
		tokx := &UTok{Tok: UT, Op: Add}
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
				enc, err := PrfF(hdxt.Auhme.Keys[0], []byte(keyword+id))
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
	totalTimeList := make([]time.Duration, 0, len(keywordsList)+1)
	volumeList := make([]int, 0, len(keywordsList)+1)

	// 循环搜索
	// keywordsList = keywordsList[:1]
	for _, keywords := range keywordsList {
		clientTimeTotal := time.Duration(0)
		serverTimeTotal := time.Duration(0)
		// 单关键词搜索, mitra part
		// 选择查询频率最低的关键字
		volume := 0
		totalStart := time.Now()
		counter, w1 := math.MaxInt64, keywords[0]
		for _, w := range keywords {
			num := hdxt.FileCnt[w]
			if num < counter {
				w1 = w
				counter = num
			}
		}
		// fmt.Println("w1 size: ", hdxt.FileCnt[w1])
		trapdoorTime, serverTime, w1Ids, err := hdxt.SearchOneKeyword(w1)
		if err != nil {
			log.Fatal(err)
		}
		clientTimeTotal += trapdoorTime
		serverTimeTotal += serverTime

		// auhme part
		// client search step 1
		q := utils.RemoveElement(keywords, w1)
		start := time.Now()
		dkList, err := AuhmeClientSearchStep1(hdxt, w1Ids, q)
		if err != nil {
			log.Fatal(err)
		}
		clientTimeTotal += time.Since(start)
		volume += CalculateDkListSize(dkList)

		// server search step
		start = time.Now()

		posList := auhmeServerSearch(hdxt, dkList)
		serverTimeTotal += time.Since(start)

		// client search step 2
		start = time.Now()
		sIdList := AuhmeClientSearchStep2(w1Ids, posList)
		clientTimeTotal += time.Since(start)
		totalTime := time.Since(totalStart)
		volume += CalculatesIdListSize(sIdList)

		// 将结果添加到结果列表
		resultList = append(resultList, sIdList)
		clientSearchTime = append(clientSearchTime, clientTimeTotal)
		serverTimeList = append(serverTimeList, serverTimeTotal)
		resultLengthList = append(resultLengthList, len(sIdList))
		totalTimeList = append(totalTimeList, totalTime)
		volumeList = append(volumeList, volume)
	}

	// 设置结果文件的路径和名称
	resultpath := filepath.Join("result", "Search", "HDXT", tableName, fmt.Sprintf("%s.csv", time.Now().Format("2006-01-02_15-04-05")))

	// 定义结果表头
	resultHeader := []string{"keyword", "clientTime", "serverTime", "totalTime", "resultLength", "payloadSize"}

	// 将结果数据整理成表格形式
	resultData := make([][]string, len(resultList))
	for i, keywords := range keywordsList {
		resultData[i] = []string{strings.Join(keywords, "#"), strconv.Itoa(int(clientSearchTime[i].Microseconds())), strconv.Itoa(int(serverTimeList[i].Microseconds())), strconv.Itoa(int(totalTimeList[i].Microseconds())), strconv.Itoa(resultLengthList[i]), strconv.Itoa(volumeList[i])}
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
	tList, err := MitraGenTrapdoor(hdxt, keyword)
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
	ids, err := MitraDecrypt(hdxt, keyword, encryptedIds)
	if err != nil {
		log.Println(err)
		return 0, 0, nil, nil
	}
	clientTime += time.Since(start)

	return clientTime, serverTime, ids, nil
}
