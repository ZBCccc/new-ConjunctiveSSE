package sdssecqClient

import (
	"ConjunctiveSSE/pkg/utils"
	"encoding/base64"
	"encoding/binary"
	"log"
	"math"
	"math/big"
	"time"

	sseclient "github.com/ZBCccc/Aura/Core/SSEClient"
	util "github.com/ZBCccc/Aura/Util"
)

var (
	// PlaintextDB *mongo.Database
	p, g *big.Int
)

// Client is the client of SDSSE-CQ.
type Client struct {
	TSet          *sseclient.SSEClient
	XSet          *sseclient.SSEClient
	CT            map[string]int
	k, kx, ki, kz []byte
	iv            []byte
}

func init() {
	// Initialize g and p
	g = big.NewInt(65537)
	p, _ = new(big.Int).SetString("69445180235231407255137142482031499329548634082242122837872648805446522657159", 10)

	// Connect to MongoDB
	// var err error
	// PlaintextDB, err = Database.MongoDBSetup("SDSSE-CQ")
	// if err != nil {
	// 	log.Fatal(err)
	// }
}

// NewClient creates a new Client.
func NewClient() *Client {
	client := &Client{
		TSet: sseclient.NewSSEClient(),
		XSet: sseclient.NewSSEClient(),
		CT:   make(map[string]int),
		k:    []byte("0123456789123456"),
		kx:   []byte("0123456789123456"),
		ki:   []byte("0123456789123456"),
		kz:   []byte("0123456789123456"),
		iv:   []byte("0123456789123456"),
	}

	// 检查 TSet 和 XSet 是否被正确初始化
	if client.TSet == nil || client.XSet == nil {
		log.Fatal("Failed to initialize SSEClient")
	}

	return client
}

func (c *Client) Update(op util.Operation, keyword string, id string) {
	// update CT
	if _, exists := c.CT[keyword]; !exists {
		c.CT[keyword] = -1
	}
	c.CT[keyword]++

	// Compute e, xind, z, y, xTag
	kw, _ := utils.PrfF(c.k, []byte(keyword))
	e, err := util.AesEncrypt([]byte(id), kw, c.iv)
	if err != nil {
		log.Fatal("Failed to AesEncrypt", err)
	}

	xind, _ := utils.PrfFp(c.ki, []byte(id), p, g)
	z, _ := utils.PrfFp(c.kz, append([]byte(keyword), big.NewInt(int64(c.CT[keyword])).Bytes()...), p, g)
	pMinus1 := new(big.Int).Sub(p, big.NewInt(1))
	zInv := new(big.Int).ModInverse(z, pMinus1)
	y := new(big.Int).Mul(xind, zInv)
	xTagHead, _ := utils.PrfFp(c.kx, []byte(keyword), p, g)
	xTagHead = new(big.Int).Mul(xTagHead, xind)
	xTag := new(big.Int).Exp(g, xTagHead, p)

	// Serialize data
	serializedData := serializeData(e, y, c.CT[keyword])

	// Run Aura.Update
	// log.Println("op =", op, "keyword =", keyword, "serializedData =", base64.StdEncoding.EncodeToString(serializedData))
	c.TSet.Update(op, keyword, base64.StdEncoding.EncodeToString(serializedData))
	c.XSet.Update(op, keyword, base64.StdEncoding.EncodeToString(xTag.Bytes()))
}

func (c *Client) Search(keywords []string) ([]string, time.Duration, time.Duration) {
	clientStart := time.Now()
	// find the least count of keywords
	minCount := math.MaxInt
	w1 := keywords[0]
	for _, keyword := range keywords {
		if count, exists := c.CT[keyword]; exists {
			if count < minCount {
				minCount = count
				w1 = keyword
			}
		} else {
			return nil, 0, 0
		}
	}

	// Initialize xtokenList
	xtokenList := make([][]*big.Int, minCount+1)
	for i := range xtokenList {
		xtokenList[i] = make([]*big.Int, len(keywords)-1)
	}

	qt := utils.RemoveElement(keywords, w1)
	for i := 0; i <= minCount; i++ {
		for j, wj := range qt {
			xtoken1, _ := utils.PrfFp(c.kx, []byte(wj), p, g)
			xtoken2, _ := utils.PrfFp(c.kz, append([]byte(w1), big.NewInt(int64(i)).Bytes()...), p, g)
			xtoken := new(big.Int).Exp(g, new(big.Int).Mul(xtoken1, xtoken2), p)
			xtokenList[i][j] = xtoken
		}
	}
	clientTime := time.Since(clientStart)

	serverStart := time.Now()
	// Run Aura.Search
	ResT := c.TSet.Search(w1)
	if ResT == nil {
		return nil, 0, 0
	}
	XSet := make(map[string]bool)
	for _, wj := range qt {
		ResX := c.XSet.Search(wj)
		if ResX == nil {
			return nil, 0, 0
		}
		for _, x := range ResX {
			XSet[x] = true
		}
	}

	// Server side
	Res := make([]string, 0, len(ResT))
	for _, v := range ResT {
		decoded, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			log.Fatal("Failed to decode string", err)
		}
		e, y, c := deserializeData(decoded)
		flag := true
		for _, xtoken := range xtokenList[c] {
			if _, exists := XSet[base64.StdEncoding.EncodeToString(new(big.Int).Exp(xtoken, y, p).Bytes())]; !exists {
				flag = false
				break
			}
		}
		if flag {
			Res = append(Res, string(e))
		}
	}
	serverTime := time.Since(serverStart)

	clientStart = time.Now()
	// Client side: decrypt
	kw1, _ := utils.PrfF(c.k, []byte(w1))
	ResInd := make([]string, 0, len(Res))
	for _, e := range Res {
		ind, err := util.AesDecrypt([]byte(e), kw1, c.iv)
		if err != nil {
			log.Fatal("Failed to AesDecrypt", err)
		}
		ResInd = append(ResInd, string(ind))
	}
	clientTime += time.Since(clientStart)

	return ResInd, clientTime, serverTime
}

// 新增辅助函数
func serializeData(e []byte, y *big.Int, counter int) []byte {
	// 为长度信息预留空间
	eLen := len(e)
	yBytes := y.Bytes()
	yLen := len(yBytes)

	result := make([]byte, 8+8+eLen+yLen+8) // 8字节分别存储eLen和yLen，8字节存储counter

	// 写入长度信息
	binary.BigEndian.PutUint64(result[0:8], uint64(eLen))
	binary.BigEndian.PutUint64(result[8:16], uint64(yLen))

	// 写入数据
	copy(result[16:16+eLen], e)
	copy(result[16+eLen:16+eLen+yLen], yBytes)
	binary.BigEndian.PutUint64(result[16+eLen+yLen:], uint64(counter))

	return result
}

func deserializeData(data []byte) (e []byte, y *big.Int, counter int) {
	eLen := binary.BigEndian.Uint64(data[0:8])
	yLen := binary.BigEndian.Uint64(data[8:16])

	e = data[16 : 16+eLen]
	y = new(big.Int).SetBytes(data[16+eLen : 16+eLen+yLen])
	counter = int(binary.BigEndian.Uint64(data[16+eLen+yLen:]))

	return
}
