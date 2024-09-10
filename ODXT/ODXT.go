package ODXT

import (
	"ConjunctiveSSE/util"
	"bufio"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"math/big"
	mrand "math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bits-and-blooms/bloom/v3"
	mapset "github.com/deckarep/golang-set/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/mongo"
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
	MySQLDB     *sql.DB
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
		odxt.Keys = ReadKeys("keys.txt")
	}

	// 初始化 UpdateCnt
	odxt.UpdateCnt = make(map[string]int)

	// 初始化 g 和 p
	odxt.g = big.NewInt(65537)
	odxt.p, _ = new(big.Int).SetString("69445180235231407255137142482031499329548634082242122837872648805446522657159", 10)

	// 初始化 XSet
	odxt.XSet = bloom.NewWithEstimates(1000000, 0.01) // 可以存储100万个元素,错误率为1%

	// 连接MongoDB
	var err error
	odxt.PlaintextDB, err = MongoDBSetup(dbName)
	if err != nil {
		log.Fatal(err)
		return err
	}

	// 连接MySQL数据库
	odxt.MySQLDB, err = MySQLSetup(dbName)
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}

func (odxt *ODXT) CiphertextGenPhase(dbName string) {
	// 获取MongoDB数据库
	plaintextDB := odxt.PlaintextDB
	defer plaintextDB.Client().Disconnect(context.Background())

	// 初始化
	uploadList := make([]UpdatePayload, 0)
	encryptTimeList := make([]time.Duration, 0)
	keywordList := make([]string, 0)
	volumeList := make([]int, 0)
	clientStorageUpdateBytes := make([]int, 0)

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
		ids = removeDuplicates(ids)
		keyword := keywordId["k"].(string)

		encryptTime, keywordCipher, err := odxt.Encrypt(keyword, ids, 1)
		if err != nil {
			log.Fatal(err)
		}

		uploadList = append(uploadList, keywordCipher...)
		encryptTimeList = append(encryptTimeList, encryptTime)
		keywordList = append(keywordList, keyword)
		volumeList = append(volumeList, len(keywordCipher))
		clientStorageUpdateBytes = append(clientStorageUpdateBytes, CalculateUpdatePayloadSize(keywordCipher))


		// 如果上传列表的长度达到最大限制， 则将其写入数据库
		if len(uploadList) >= UploadListMaxLength {
			// 写入文件
			err = WriteUploadList(odxt.MySQLDB, uploadList, dbName)
			if err != nil {
				log.Fatal(err)
			}

			// 清空上传列表
			uploadList = make([]UpdatePayload, 0)
		}
	}

	// 如果上传列表不为空， 则将其写入数据库
	if len(uploadList) > 0 {
		// 写入文件
		err = WriteUploadList(odxt.MySQLDB, uploadList, dbName)
		if err != nil {
			log.Fatal(err)
		}
	}

	saveTime := time.Now()
	// 保存 XSet 到文件
	err = SaveBloomFilterToFile(odxt.XSet, filepath.Join("result", "Update", "ODXT", fmt.Sprintf("%s_%s_XSet.bin", dbName, saveTime.Format("2006-01-02_15-04-05"))))
	if err != nil {
		log.Fatal(err)
	}

	// 设置结果文件的路径和名称
	resultpath := filepath.Join("result", "Addition", "ODXT", fmt.Sprintf("%s_%s.csv", dbName, saveTime.Format("2006-01-02_15-04-05")))

	// 定义结果表头
	resultHeader := []string{"keyword", "volume", "addTime", "storageUpdateBytes"}

	// 将结果数据整理成表格形式
	resultData := make([][]string, len(keywordList))
	for i, keyword := range keywordList {
		resultData[i] = []string{keyword, strconv.Itoa(volumeList[i]), encryptTimeList[i].String(), strconv.Itoa(clientStorageUpdateBytes[i])}
	}

	// 将结果写入文件
	err = util.WriteResult(resultpath, resultHeader, resultData)
	if err != nil {
		log.Fatal(err)
	}
}


func (odxt *ODXT) Encrypt(keyword string, ids []string, operation int) (time.Duration, []UpdatePayload, error) {
	kt, kx, ky, kz := odxt.Keys[0], odxt.Keys[1], odxt.Keys[2], odxt.Keys[3]
	p, g := odxt.p, odxt.g

	var encryptedTime time.Duration
	keywordsCipher := make([]UpdatePayload, 0)

	_, ok := odxt.UpdateCnt[keyword]
	if !ok {
		odxt.UpdateCnt[keyword] = 0
	}

	for _, id := range ids {
		start := time.Now()

		odxt.UpdateCnt[keyword]++

		// address = PRF(kt, w||wc||0)
		wWc := append([]byte(keyword), big.NewInt(int64(odxt.UpdateCnt[keyword])).Bytes()...)
		address, err := util.PrfF(kt, append(wWc, big.NewInt(int64(0)).Bytes()...))
		if err != nil {
			log.Println(err)
			return time.Second, nil, err
		}

		// val = PRF(kt, w||wc||1) xor (id||op)
		val, _ := util.PrfF(kt, append(wWc, big.NewInt(int64(1)).Bytes()...))
		if err != nil {
			log.Println(err)
			return time.Second, nil, err
		}
		val, _ = util.BytesXORWithOp(val, []byte(id), operation)
		if err != nil {
			log.Println(err)
			return time.Second, nil, err
		}

		// alpha = Fp(ky, id||op) * Fp(kz, w||wc)^-1
		alpha, alpha1, _ := util.ComputeAlpha(ky, kz, []byte(id), operation, wWc, p, g)
		if err != nil {
			log.Println(err)
			return time.Second, nil, err
		}

		// xtag = g^{Fp(Kx, w)*Fp(Ky, id||op)} mod p
		C, _ := util.PrfFp(kx, []byte(keyword), p, g)
		if err != nil {
			log.Println(err)
			return time.Second, nil, err
		}
		A := new(big.Int).Mul(C, alpha1)
		xtag := new(big.Int).Exp(g, A, p)

		encryptedTime += time.Since(start)

		// Encoded the ciphertext
		base64Address := base64.StdEncoding.EncodeToString(address)
		base64Val := base64.StdEncoding.EncodeToString(val)
		base64Alpha := base64.StdEncoding.EncodeToString(alpha.Bytes())

		keywordsCipher = append(keywordsCipher, UpdatePayload{base64Address, base64Val, base64Alpha})
		odxt.XSet.Add(xtag.Bytes())
	}

	return encryptedTime, keywordsCipher, nil
}

func (odxt *ODXT) DeletionPhaseWithSearch(del_rate int) {

}

func (odxt *ODXT) SearchPhase(q []string, tableName string) {
	// 读取待搜索的连接关键词文件，文件格式为：
	// 每一行都是关键词的集合，关键词之间用#隔开
	// 例如：
	// 关键词1#关键词2#关键词3
	// 关键词4#关键词5
	// 关键词6
	// 读取待搜索的关键词文件
	file, err := os.Open("search_keywords.txt")
	if err != nil {
		log.Fatal("无法打开文件:", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var keywordsList [][]string

	for scanner.Scan() {
		line := scanner.Text()
		keywords := strings.Split(line, "#")
		keywordsList = append(keywordsList, keywords)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("读取文件时出错:", err)
	}

	// 打印读取到的关键词列表（可选）
	fmt.Println("待搜索的关键词列表:")
	for i, keywords := range keywordsList {
		fmt.Printf("第%d组关键词: %v\n", i+1, keywords)
	}
	
	// 搜索
	trapdoorTime, serverTime, sEOpList := odxt.Search(q, tableName)

	// 解密密文获得最终结果
	start := time.Now()
	sIdList, err := odxt.Decrypt(q, sEOpList)
	if err != nil {
		log.Fatal(err)
	}
	decryptTime := time.Since(start)

	clientTime := trapdoorTime + decryptTime

	// 设置结果文件的路径和名称
	resultpath := filepath.Join("result", "Search", "ODXT", fmt.Sprintf("%s_%s.csv", tableName, time.Now().Format("2006-01-02_15-04-05")))

	// 定义结果表头
	resultHeader := []string{"keyword", "clientSearchTime", "serverTime", "resultLength"}

	// 将结果数据整理成表格形式
	resultData := make([][]string, len(sIdList))
	for i, sId := range sIdList {
		resultData[i] = []string{sId, strconv.Itoa(volumeList[i]), encryptTimeList[i].String(), strconv.Itoa(clientStorageUpdateBytes[i])}
	}

	// 将结果写入文件
	err = util.WriteResult(resultpath, resultHeader, resultData)
	if err != nil {
		log.Fatal(err)
	}

}

// Search 搜索，生成search token，并查询SQL数据库
func (odxt *ODXT) Search(q []string, tableName string) (time.Duration, time.Duration, []util.SEOp) {
	start := time.Now()
	// 生成陷门
	stokenList, xtokenList := odxt.Trapdoor(q)
	trapdoorTime := time.Since(start)

	start = time.Now()
	// 查询SQL数据库
	tmpResult, err := SearchStoken(odxt.MySQLDB, stokenList, tableName)
	if err != nil {
		log.Fatal(err)
	}

	sEOpList := make([]util.SEOp, len(stokenList))

	// 搜索数据
	var cnt int
	for j, value := range tmpResult {
		cnt = 1

		// 遍历 xtokenList
		for _, xtoken := range xtokenList[j] {
			// 类型转换
			xtokenInt, err := util.Base64ToBigInt(xtoken)
			if err != nil {
				log.Fatal(err)
			}
			alpha, err := util.Base64ToBigInt(value.Alpha)
			if err != nil {
				log.Fatal(err)
			}

			// 判断 xtag 是否匹配
			xtag := new(big.Int).Exp(xtokenInt, alpha, odxt.p)
			if odxt.XSet.Test(xtag.Bytes()) {
				cnt++
			}
		}

		sEOpList[j] = util.SEOp{
			J:    j + 1,
			Sval: value.Value,
			Cnt:  cnt,
		}
	}
	serverTime := time.Since(start)

	return trapdoorTime, serverTime, sEOpList
}

// Trapdoor 生成陷门
func (odxt *ODXT) Trapdoor(q []string) ([]string, [][]string) {
	// 读取密钥
	kt, kx, kz := odxt.Keys[0], odxt.Keys[1], odxt.Keys[3]
	counter, w1, st := 1000000, q[0], odxt.UpdateCnt

	// 选择查询频率最低的关键字
	for _, w := range q {
		num := st[w]
		if num < counter {
			w1 = w
			counter = num
		}
	}

	// 初始化stokenList和xtokenList
	stokenList := make([]string, counter)
	xtokenList := make([][]string, counter)
	for i := range xtokenList {
		xtokenList[i] = make([]string, len(q)-1)
	}
	for j := 0; j < counter; j++ {
		saddr, err := util.PrfF(kt, append(append([]byte(w1), big.NewInt(int64(j+1)).Bytes()...), big.NewInt(int64(0)).Bytes()...))
		if err != nil {
			fmt.Println(err)
		}
		stokenList[j] = base64.StdEncoding.EncodeToString(saddr)

		i := 0
		for _, wi := range q {
			if wi == w1 {
				continue
			}
			xtoken1, _ := util.PrfFp(kx, []byte(wi), odxt.p, odxt.g)
			xtoken2, _ := util.PrfFp(kz, append([]byte(w1), big.NewInt(int64(j+1)).Bytes()...), odxt.p, odxt.g)
			xtoken := new(big.Int).Exp(odxt.g, new(big.Int).Mul(xtoken1, xtoken2), odxt.p)
			xtokenList[j][i] = base64.StdEncoding.EncodeToString(xtoken.Bytes())
			i++
		}

		// 打乱切片中的元素
		mrand.Shuffle(len(xtokenList[j]), func(i, j int) {
			xtokenList[j][i], xtokenList[j][j] = xtokenList[j][j], xtokenList[j][i]
		})
	}

	return stokenList, xtokenList
}

// Decrypt 解密
func (odxt *ODXT) Decrypt(q []string, sEOpList []util.SEOp) ([]string, error) {
	kt := odxt.Keys[0]
	counter, w1, st := 1000000, q[0], odxt.UpdateCnt

	// 选择查询频率最低的关键字
	for _, w := range q {
		num := st[w]
		if num < counter {
			w1 = w
			counter = num
		}
	}

	sIdList := make([]string, 0)
	for _, sEOp := range sEOpList {
		j, sval, cnt := sEOp.J, sEOp.Sval, sEOp.Cnt
		w1Andj := append(append([]byte(w1), big.NewInt(int64(j)).Bytes()...), big.NewInt(int64(1)).Bytes()...)
		tmp, err := util.PrfF(kt, w1Andj)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		id := make([]byte, 31)
		val, err := base64.StdEncoding.DecodeString(sval)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		for i := 0; i < 31; i++ {
			id[i] = tmp[i] ^ val[i]
		}
		var op util.Operation = util.Operation(tmp[31] ^ val[31])
		if op == util.Add && cnt == len(q) {
			sIdList = append(sIdList, base64.StdEncoding.EncodeToString(id))
		} else if op == util.Del && cnt > 0 {
			sIdList = util.RemoveElement(sIdList, base64.StdEncoding.EncodeToString(id))
		}
	}

	return sIdList, nil
}

// removeDuplicates 去除切片中的重复元素
func removeDuplicates(intSlice []string) []string {
	// 创建一个新的string集合
	stringSet := mapset.NewSet[string]()

	// 将切片中的元素添加到集合中
	for _, v := range intSlice {
		stringSet.Add(v)
	}

	// 转换为切片
	return stringSet.ToSlice()
}

// CalculateUpdatePayloadSize 计算[]UpdatePayload的字节大小
func CalculateUpdatePayloadSize(payloads []UpdatePayload) int {
	size := 0
	for _, payload := range payloads {
		size += len(payload.Address) + len(payload.Val) + len(payload.Alpha)
	}
	return size
}

// 保存 Bloom filter 到文件
func SaveBloomFilterToFile(filter *bloom.BloomFilter, filename string) error {
	// 创建文件，如果所在目录不存在，则先创建目录，再创建文件
	dir := filepath.Dir(filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}

	// 创建文件
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// 将 Bloom filter 写入文件
	_, err = filter.WriteTo(file)
	return err
}

// 从文件加载 Bloom filter
func LoadBloomFilterFromFile(filename string) (*bloom.BloomFilter, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    filter := bloom.NewWithEstimates(1000000, 0.01) // 创建一个新的 Bloom filter，使用整数参数
    _, err = filter.ReadFrom(file)
    if err != nil {
        return nil, err
    }

	return filter, nil
}