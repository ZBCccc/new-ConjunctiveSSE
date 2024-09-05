package ODXT

import (
	"ConjunctiveSSE/util"
	"context"
	"crypto/rand"
	"github.com/bits-and-blooms/bloom/v3"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"math/big"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//var PlaintextDB *mongo.Collection

const (
	MaxConnection       = 100
	UploadListMaxLength = 200000
)

type ODXT struct {
	Keys        [4][]byte
	UpdateCnt   map[string]int
	g           *big.Int
	p           *big.Int
	XSet        *bloom.BloomFilter
	PlaintextDB *mongo.Database
}

type UpdatePayload struct {
	Address []byte
	Val     []byte
	Alpha   *big.Int
}

func (odxt *ODXT) DBSetup(db_name string) error {
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

	// 初始化 UpdateCnt
	odxt.UpdateCnt = make(map[string]int)

	// 初始化 g 和 p
	odxt.g = big.NewInt(65537)
	odxt.p, _ = new(big.Int).SetString("69445180235231407255137142482031499329548634082242122837872648805446522657159", 10)

	// 初始化 XSet
	odxt.XSet = bloom.NewWithEstimates(1000000, 0.01) // 可以存储100万个元素,错误率为1%

	// 连接MongoDB
	// Set client options
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
		return err
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
		return err
	}

	// 获取数据库和集合的句柄
	odxt.PlaintextDB = client.Database(db_name)

	return nil
}

func (odxt *ODXT) CiphertextGenPhase() {
	// 获取MongoDB数据库
	plaintextDB := odxt.PlaintextDB

	uploadList := make([]UpdatePayload, UploadListMaxLength)
	encryptTimeList := make([]time.Duration, 0)

	// 从MongoDB数据库中获取名为"id_keywords"的集合
	collection := plaintextDB.Collection("id_keywords")

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
		ids := keywordId["id"].([]string)
		keyword := keywordId["keyword"].(string)
		encryptTime, keywordCipher, err := odxt.Encrypt(keyword, ids, 1)
		if err != nil {
			log.Fatal(err)
		}
		uploadList = append(uploadList, keywordCipher...)
		encryptTimeList = append(encryptTimeList, encryptTime)

		// 如果上传列表的长度达到最大限制， 则将其写入数据库
		if len(uploadList) >= UploadListMaxLength {
			// 写入文件
			err = WriteUploadList(uploadList, "uploadList.txt")
			if err != nil {
				log.Fatal(err)
			}

			// 清空上传列表
			uploadList = make([]UpdatePayload, UploadListMaxLength)
		}
	}

	// 如果上传列表不为空， 则将其写入数据库
	if len(uploadList) > 0 {
		// 写入文件
		err = WriteUploadList(uploadList, "uploadList.txt")
		if err != nil {
			log.Fatal(err)
		}
	}

	// 写入加密时间
	err = WriteEncryptTime(encryptTimeList, "encryptTime.txt")
	if err != nil {
		log.Fatal(err)
	}

	return

}

func WriteUploadList(uploadList []UpdatePayload, filename string) error {
	// 写入文件，将[]UpdatePayload写入文件
	// 打开文件
	




	return nil
}

func (odxt *ODXT) Encrypt(keyword string, ids []string, operation int) (time.Duration, []UpdatePayload, error) {
	kt, kx, ky, kz := odxt.Keys[0], odxt.Keys[1], odxt.Keys[2], odxt.Keys[3]
	p, g := odxt.p, odxt.g

	var encryptedTime time.Duration
	keywordsCipher := make([]UpdatePayload, len(ids))
	start := time.Now()
	for _, id := range ids {
		_, ok := odxt.UpdateCnt[keyword]
		if !ok {
			odxt.UpdateCnt[keyword] = 0
		}
		odxt.UpdateCnt[keyword]++

		// address = PRF(kt, w||wc||0)
		wWc := append([]byte(keyword), big.NewInt(int64(odxt.UpdateCnt[keyword])).Bytes()...)
		address, err := util.PrfF(kt, append(wWc, big.NewInt(int64(0)).Bytes()...))
		if err != nil {
			log.Println(err)
			return time.Second, nil, err
		}

		// val = PRF(kt, w||wc||1) xor (id||op)
		val, err := util.PrfF(kt, append(wWc, big.NewInt(int64(1)).Bytes()...))
		if err != nil {
			log.Println(err)
			return time.Second, nil, err
		}
		val, err = util.BytesXORWithOp(val, []byte(id), operation)
		if err != nil {
			log.Println(err)
			return time.Second, nil, err
		}

		// alpha = Fp(ky, id||op) * Fp(kz, w||wc)^-1
		alpha, alpha1, err := util.ComputeAlpha(ky, kz, []byte(id), operation, wWc, p, g)
		if err != nil {
			log.Println(err)
			return time.Second, nil, err
		}

		// xtag = g^{Fp(Kx, w)*Fp(Ky, id||op)} mod p
		C, err := util.PrfFp(kx, []byte(keyword), p, g)
		if err != nil {
			log.Println(err)
			return time.Second, nil, err
		}
		A := new(big.Int).Mul(C, alpha1)
		xtag := new(big.Int).Exp(g, A, p)

		encryptedTime += time.Since(start)
		keywordsCipher = append(keywordsCipher, UpdatePayload{address, val, alpha})
		odxt.XSet.Add(xtag.Bytes())
	}

	return encryptedTime, keywordsCipher, nil
}

func DeletionPhaseWithSearch(del_rate int) {
}
