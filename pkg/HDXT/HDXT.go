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

type CipherText struct {
	Address string
	Value   string
	gorm.Model
}

type AuhmeCipherText struct {
	Label string
	Enc   string
}

type UpdatePayload struct {
	Address string
	Val     string
}

var (
	universeIDNums int
	universeIDs    []string
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

	// 初始化EDB
	db, err := MySQLSetup(dbName)
	if err != nil {
		log.Println("Error initializing TMap:", err)
		panic(err)
	}
	hdxt.TMap = db

	// 连接MongoDB
	hdxt.PlaintextDB, err = Database.MongoDBSetup(dbName)
	if err != nil {
		log.Fatal(err)
		return err
	}

	// 获取id数量
	universeIDs, err = Database.GetUniqueValSets(hdxt.PlaintextDB)
	if err != nil {
		log.Fatal(err)
		return err
	}
	universeIDNums = len(universeIDs)

	// 初始化Auhme
	hdxt.Auhme.Cnt = 0
	hdxt.Auhme.S = make([]string, 0)
	hdxt.Auhme.T = make(map[string]int)
	hdxt.Auhme.Delta = 0

	return nil
}

func (hdxt *HDXT) SetupPhase() error {
	db := hdxt.TMap

	// 获取MongoDB数据库
	plaintextDB := hdxt.PlaintextDB
	defer plaintextDB.Client().Disconnect(context.Background())

	// 初始化
	uploadList := make([]CipherText, 0, UploadListMaxLength+1)
	auhmeCipherList := make([]AuhmeCipherText, 0, UploadListMaxLength+1)
	encryptTimeList := make([]time.Duration, 0, 1000000)
	keywordList := make([]string, 0, 1000000)
	volumeList := make([]int, 0, 1000000)
	clientStorageUpdateBytes := make([]int, 0, 1000000)

	// 从MongoDB数据库中获取名为"id_keywords"的集合
	collection := plaintextDB.Collection("id_keywords")

	// 创建一个游标，设置不超时并每次获取1000条记录
	ctx := context.TODO()
	opts := options.Find().SetNoCursorTimeout(true).SetBatchSize(1000)
	cur, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		log.Fatal(err)
		return err
	}

	// 关闭游标
	defer cur.Close(ctx)

	// 读取游标中的所有记录
	var keywordIds []bson.M
	if err = cur.All(ctx, &keywordIds); err != nil {
		log.Fatal(err)
		return err
	}

	// 读取所有记录
	for _, keywordId := range keywordIds {
		valSet, ok := keywordId["val_set"].(primitive.A)
		if !ok {
			log.Fatal("val_set is not of type primitive.A")
			return fmt.Errorf("val_set is not of type primitive.A")
		}
		var ids []string
		for _, v := range valSet {
			if str, ok := v.(string); ok {
				ids = append(ids, str)
			} else {
				log.Fatal("val_set contains non-string value")
				return fmt.Errorf("val_set contains non-string value")
			}
		}
		ids = utils.RemoveDuplicates(ids)
		keyword := keywordId["k"].(string)

		encryptTime, keywordCipher, auhmeCipher, err := hdxt.Setup(keyword, ids, 1)
		if err != nil {
			log.Fatal(err)
			return err
		}

		uploadList = append(uploadList, keywordCipher...)
		auhmeCipherList = append(auhmeCipherList, auhmeCipher...)
		encryptTimeList = append(encryptTimeList, encryptTime)
		keywordList = append(keywordList, keyword)
		volumeList = append(volumeList, len(keywordCipher))
		clientStorageUpdateBytes = append(clientStorageUpdateBytes, CalculateUpdatePayloadSize(keywordCipher))

		// 如果上传列表的长度达到最大限制， 则将其写入数据库
		if len(uploadList) >= UploadListMaxLength {
			// 写入文件
			err = WriteUploadList(db, uploadList)
			if err != nil {
				log.Fatal(err)
				return err
			}

			// 清空上传列表
			uploadList = make([]CipherText, 0, UploadListMaxLength+1)
		}
		if len(auhmeCipherList) >= UploadListMaxLength {
			// 写入文件
			err = WriteAuhmeCipherList(db, auhmeCipherList)
			if err != nil {
				log.Fatal(err)
				return err
			}

			// 清空上传列表
			auhmeCipherList = make([]AuhmeCipherText, 0, UploadListMaxLength+1)
		}
	}

	// 如果上传列表不为空， 则将其写入数据库
	if len(uploadList) > 0 {
		// 写入文件
		err = WriteUploadList(db, uploadList)
		if err != nil {
			log.Fatal(err)
			return err
		}
	}
	if len(auhmeCipherList) > 0 {
		// 写入文件
		err = WriteAuhmeCipherList(db, auhmeCipherList)
		if err != nil {
			log.Fatal(err)
			return err
		}
	}

	saveTime := time.Now()

	// 保存 odxt.UpdateCnt 到文件
	err = utils.SaveUpdateCntToFile(hdxt.FileCnt, filepath.Join("result", "Update", "HDXT", fmt.Sprintf("%s_UpdateCnt.json", saveTime.Format("2006-01-02_15-04-05"))))
	if err != nil {
		log.Fatal(err)
	}

	// 设置结果文件的路径和名称
	resultpath := filepath.Join("result", "Update", "HDXT", fmt.Sprintf("%s.csv", saveTime.Format("2006-01-02_15-04-05")))

	// 定义结果表头
	resultHeader := []string{"keyword", "volume", "addTime", "storageUpdateBytes"}

	// 将结果数据整理成表格形式
	resultData := make([][]string, len(keywordList))
	for i, keyword := range keywordList {
		resultData[i] = []string{keyword, strconv.Itoa(volumeList[i]), encryptTimeList[i].String(), strconv.Itoa(clientStorageUpdateBytes[i])}
	}

	// 将结果写入文件
	err = utils.WriteResultToCSV(resultpath, resultHeader, resultData)
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func (hdxt *HDXT) Setup(keyword string, ids []string, operation int) (time.Duration, []CipherText, []AuhmeCipherText, error) {
	var encryptedTime time.Duration
	keywordsCipher := make([]CipherText, 0, len(ids))
	auhmeCipher := make([]AuhmeCipherText, universeIDNums)

	_, ok := hdxt.FileCnt[keyword]
	if !ok {
		hdxt.FileCnt[keyword] = 0
	}

	for i, id := range universeIDs {
		if slices.Contains(ids, id) {
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
			keywordsCipher = append(keywordsCipher, CipherText{Address: address, Value: val})
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

	return encryptedTime, keywordsCipher, auhmeCipher, nil
}

// CalculateUpdatePayloadSize 计算[]UpdatePayload的字节大小
func CalculateUpdatePayloadSize(payloads []CipherText) int {
	size := 0
	for _, payload := range payloads {
		size += len(payload.Address) + len(payload.Value)
	}
	return size
}
