package FDHXT

import (
	"ConjunctiveSSE/pkg/Database"
	"ConjunctiveSSE/pkg/utils"
	"bufio"
	"context"
	"encoding/base64"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type counter struct {
	srch, updt, max int
}

type FDXT struct {
	Keys    [5][]byte
	Count   map[string]*counter
	CDBXtag map[string]string
	CDBTSet map[string]string
	XSet    *bloom.BloomFilter
}

var (
	PlaintextDB *mongo.Database
	err         error
)

func ReadKeys(fileName string) [5][]byte {
	// 读取文件
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// 读取密钥
	var keys [5][]byte
	scanner := bufio.NewScanner(file)

	// 读取5个密钥
	for i := 0; i < 5; i++ {
		line := scanner.Text()
		key, err := base64.StdEncoding.DecodeString(line)
		if err != nil {
			log.Fatal(err)
		}

		keys[i] = key
	}

	return keys
}

func (fdxt *FDXT) Setup(dbName string) error {
	fdxt.Keys = ReadKeys("./cmd/FDHXT/configs/keys.txt")
	fdxt.Count = make(map[string]*counter)
	fdxt.CDBXtag = make(map[string]string)
	fdxt.CDBTSet = make(map[string]string)
	fdxt.XSet = bloom.NewWithEstimates(1000000, 0.01)

	// 初始化mongodb
	PlaintextDB, err = Database.MongoDBSetup(dbName)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (fdxt *FDXT) UpdatePhase(dbName string) error {
	// 获取MongoDB数据库
	plaintextDB := PlaintextDB
	defer plaintextDB.Client().Disconnect(context.Background())

	// 从MongoDB数据库中获取名为"id_keywords"的集合
	collection := plaintextDB.Collection("id_keywords")

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

		encryptTime, err := fdxt.Encrypt(keyword, ids, Add)
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}

func (fdxt *FDXT) Encrypt(keyword string, ids []string, op Operation) (time.Duration, error) {
	kw, kt, kx, ky, kz := fdxt.Keys[0], fdxt.Keys[1], fdxt.Keys[2], fdxt.Keys[3], fdxt.Keys[4]
	if _, ok := fdxt.Count[keyword]; !ok {
		fdxt.Count[keyword] = &counter{srch: 0, updt: 0, max: 0}
	}
	fdxt.Count[keyword].updt++
	fdxt.Count[keyword].max++
	addr, err := utils.PrfF(kw, append(append([]byte(keyword), big.NewInt(int64(fdxt.Count[keyword].max)).Bytes()...), byte(op)))
	if err != nil {
		return 0, err
	}

}
