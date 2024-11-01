package HDXT

import (
	"ConjunctiveSSE/pkg/Database"
	"ConjunctiveSSE/pkg/utils"
	"context"
	"crypto/rand"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
	"log"
	"path/filepath"
	"slices"
	"strconv"
	"time"
)

const (
	UploadListMaxLength = 200000
)

type Mitra struct {
	Key     []byte
	FileCnt map[string]int
	TMap    *gorm.DB
}

type Auhme struct {
	Keys  [3][]byte
	Cnt   int
	S     []string
	T     map[string]int
	Delta int
}

type HDXT struct {
	PlaintextDB *mongo.Database
	Mitra
	Auhme
}

type MitraCipherText struct {
	Address string
	Value   string
	gorm.Model
}

type AuhmeCipherText struct {
	Label string
	Enc   string
	gorm.Model
}

type UpdatePayload struct {
	Address string
	Val     string
}

var (
	universeKeywordsNums int
	universeKeywords     []string
)

func (hdxt *HDXT) Init(dbName string, randomKey bool) error {
	// 初始化私钥
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
				log.Println("Error generating random key:", err)
				return err
			}
			hdxt.Auhme.Keys[i] = key
		}
	} else {
		// 读取私钥
		var err error
		hdxt.Mitra.Key, hdxt.Auhme.Keys, err = utils.HdxtReadKeys("./cmd/HDXT/keys.txt")
		if err != nil {
			log.Println("Error reading keys:", err)
			return err
		}
	}

	// 初始化FileCnt
	hdxt.Mitra.FileCnt = make(map[string]int)

	// 初始化EDB，根据选择的数据集建立对应的Mitra和Auhme表
	db, err := MySQLSetup(dbName)
	if err != nil {
		log.Println("Error initializing TMap:", err)
		panic(err)
	}
	hdxt.TMap = db

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
	universeKeywordsNums = len(universeKeywords)

	// 初始化Auhme
	hdxt.Auhme.Cnt = 0
	hdxt.Auhme.S = make([]string, 0)
	hdxt.Auhme.T = make(map[string]int)
	hdxt.Auhme.Delta = len(hdxt.Auhme.T)

	return nil
}

type volume struct {
	mitraVolume int
	auhmeVolume int
}

func (hdxt *HDXT) SetupPhase() error {
	db := hdxt.TMap

	// 获取MongoDB数据库
	plaintextDB := hdxt.PlaintextDB
	defer plaintextDB.Client().Disconnect(context.Background())

	// 初始化
	mitraCipherList := make([]MitraCipherText, 0, UploadListMaxLength+1)
	auhmeCipherList := make([]AuhmeCipherText, 0, UploadListMaxLength+1)
	encryptTimeList := make([]time.Duration, 0, 1000000)
	idList := make([]string, 0, 1000000)
	volumeList := make([]volume, 0, 1000000)
	clientStorageUpdateBytes := make([]int, 0, 1000000)

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
	for _, idKeyword := range idKeywords {
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

		encryptTime, mitraCipher, auhmeCipher, err := hdxt.Setup(id, keywords, 1)
		if err != nil {
			log.Println("Error in Setup:", err)
			return err
		}

		mitraCipherList = append(mitraCipherList, mitraCipher...)
		auhmeCipherList = append(auhmeCipherList, auhmeCipher...)
		encryptTimeList = append(encryptTimeList, encryptTime)
		idList = append(idList, id)
		volumeList = append(volumeList, volume{mitraVolume: len(mitraCipher), auhmeVolume: len(auhmeCipher)})
		clientStorageUpdateBytes = append(clientStorageUpdateBytes, CalculateUpdatePayloadSize(mitraCipher))

		// 如果上传列表的长度达到最大限制， 则将其写入数据库
		if len(mitraCipherList) >= UploadListMaxLength {
			// 写入文件
			err = WriteMitraCipherList(db, mitraCipherList)
			if err != nil {
				log.Println("Error writing MitraCipherList:", err)
				return err
			}

			// 清空上传列表
			mitraCipherList = make([]MitraCipherText, 0, UploadListMaxLength+1)
		}
		if len(auhmeCipherList) >= UploadListMaxLength {
			// 写入文件
			err = WriteAuhmeCipherList(db, auhmeCipherList)
			if err != nil {
				log.Println("Error writing AuhmeCipherList:", err)
				return err
			}

			// 清空上传列表
			auhmeCipherList = make([]AuhmeCipherText, 0, UploadListMaxLength+1)
		}
	}

	// Update Phase
	for _, idKeyword := range idKeywords {
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
		encryptTime, mitraCipher, auhmeCipher, err := hdxt.Encrypt(id, keywords, 1)
		if err != nil {
			log.Println("Error in Encrypt:", err)
			return err
		}

		mitraCipherList = append(mitraCipherList, mitraCipher...)
		auhmeCipherList = append(auhmeCipherList, auhmeCipher...)
		encryptTimeList = append(encryptTimeList, encryptTime)
		idList = append(idList, id)
		volumeList = append(volumeList, volume{mitraVolume: len(mitraCipher), auhmeVolume: len(auhmeCipher)})
		clientStorageUpdateBytes = append(clientStorageUpdateBytes, CalculateUpdatePayloadSize(mitraCipher))

		// 如果上传列表的长度达到最大限制， 则将其写入数据库
		if len(mitraCipherList) >= UploadListMaxLength {
			// 写入文件
			err = WriteMitraCipherList(db, mitraCipherList)
			if err != nil {
				log.Println("Error writing MitraCipherList:", err)
				return err
			}

			// 清空上传列表
			mitraCipherList = make([]MitraCipherText, 0, UploadListMaxLength+1)
		}
		if len(auhmeCipherList) >= UploadListMaxLength {
			// 写入文件
			err = WriteAuhmeCipherList(db, auhmeCipherList)
			if err != nil {
				log.Println("Error writing AuhmeCipherList:", err)
				return err
			}

			// 清空上传列表
			auhmeCipherList = make([]AuhmeCipherText, 0, UploadListMaxLength+1)
		}
	}
	// 如果上传列表不为空， 则将其写入数据库
	if len(mitraCipherList) > 0 {
		// 写入文件
		err = WriteMitraCipherList(db, mitraCipherList)
		if err != nil {
			log.Println("Error writing MitraCipherList:", err)
			return err
		}
	}
	if len(auhmeCipherList) > 0 {
		// 写入文件
		err = WriteAuhmeCipherList(db, auhmeCipherList)
		if err != nil {
			log.Println("Error writing AuhmeCipherList:", err)
			return err
		}
	}

	saveTime := time.Now()

	// 保存 odxt.UpdateCnt 到文件
	err = utils.SaveUpdateCntToFile(hdxt.FileCnt, filepath.Join("result", "Update", "HDXT", fmt.Sprintf("%s_UpdateCnt.json", saveTime.Format("2006-01-02_15-04-05"))))
	if err != nil {
		log.Println("Error saving UpdateCnt to file:", err)
		return err
	}

	// 设置结果文件的路径和名称
	resultpath := filepath.Join("result", "Update", "HDXT", fmt.Sprintf("%s.csv", saveTime.Format("2006-01-02_15-04-05")))

	// 定义结果表头
	resultHeader := []string{"keyword", "volume", "addTime", "storageUpdateBytes"}

	// 将结果数据整理成表格形式
	resultData := make([][]string, len(idList))
	for i, id := range idList {
		resultData[i] = []string{id, strconv.Itoa(volumeList[i].mitraVolume + volumeList[i].auhmeVolume), encryptTimeList[i].String(), strconv.Itoa(clientStorageUpdateBytes[i])}
	}

	// 将结果写入文件
	err = utils.WriteResultToCSV(resultpath, resultHeader, resultData)
	if err != nil {
		log.Println("Error writing result to file:", err)
		return err
	}

	return nil
}

func (hdxt *HDXT) Setup(id string, keywords []string, operation int) (time.Duration, []MitraCipherText, []AuhmeCipherText, error) {
	var encryptedTime time.Duration
	mitraCipher := make([]MitraCipherText, 0, len(keywords))
	auhmeCipher := make([]AuhmeCipherText, universeKeywordsNums)

	for i, keyword := range universeKeywords {
		if slices.Contains(keywords, keyword) {
			if _, ok := hdxt.FileCnt[keyword]; !ok {
				hdxt.FileCnt[keyword] = 0
			}
			start := time.Now()

			// OXT Part
			address, val, err := mitraEncrypt(hdxt, keyword, id, operation)
			if err != nil {
				log.Println(err)
				return 0, nil, nil, err
			}

			// Auhme Part
			label, enc, err := auhmeEncrypt(hdxt, keyword, id, 1, 0)
			if err != nil {
				log.Println(err)
				return 0, nil, nil, err
			}

			encryptedTime += time.Since(start)
			auhmeCipher[i] = AuhmeCipherText{Label: label, Enc: enc}
			mitraCipher = append(mitraCipher, MitraCipherText{Address: address, Value: val})
		} else {
			start := time.Now()
			// Auhme Part
			label, enc, err := auhmeEncrypt(hdxt, keyword, id, 0, 0)
			if err != nil {
				log.Println(err)
				return 0, nil, nil, err
			}

			encryptedTime += time.Since(start)
			auhmeCipher[i] = AuhmeCipherText{Label: label, Enc: enc}
		}
	}

	return encryptedTime, mitraCipher, auhmeCipher, nil
}

// CalculateUpdatePayloadSize 计算[]UpdatePayload的字节大小
func CalculateUpdatePayloadSize(payloads []MitraCipherText) int {
	size := 0
	for _, payload := range payloads {
		size += len(payload.Address) + len(payload.Value)
	}
	return size
}

func (hdxt *HDXT) Encrypt(id string, keywords []string, operation int) (time.Duration, []MitraCipherText, []AuhmeCipherText, error) {
	var encryptedTime time.Duration
	mitraCipher := make([]MitraCipherText, 0, len(keywords))
	auhmeCipher := make([]AuhmeCipherText, universeKeywordsNums)

	// op == add
	if operation == 1 {
		for _, keyword := range universeKeywords {
			if slices.Contains(keywords, keyword) {
				if _, ok := hdxt.FileCnt[keyword]; !ok {
					hdxt.FileCnt[keyword] = 0
				}
				// oxt part
				address, val, err := mitraEncrypt(hdxt, keyword, id, operation)
				if err != nil {
					log.Println("Error in Encrypt:", err)
					return 0, nil, nil, err
				}

				// auhme part


				mitraCipher = append(mitraCipher, MitraCipherText{Address: address, Value: val})
			}
		}
	}

	return encryptedTime, mitraCipher, auhmeCipher, nil
}
