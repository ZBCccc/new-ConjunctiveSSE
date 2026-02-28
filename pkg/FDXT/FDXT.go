package FDXT

import (
	"ConjunctiveSSE/pkg/Database"
	"ConjunctiveSSE/pkg/utils"
	"context"
	"encoding/json"
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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	pbcUtil "ConjunctiveSSE/pkg/utils/pbc"

	"github.com/Nik-U/pbc"
)

type Counter struct {
	srch, updt, Max int
}

type TsetValue struct {
	Val   string
	Alpha *pbc.Element
}

type FDXT struct {
	Keys    [5][]byte
	Count   map[string]*Counter
	CDBXtag map[string]string
	CDBTSet map[string]*TsetValue
	XSet    map[string]int
}

var (
	PlaintextDB *mongo.Database
	err         error
)

func (fdxt *FDXT) Setup(dbName string, mongoURI ...string) error {
	// Fixed keys for reproducible benchmarking as used in the paper.
	fdxt.Keys[0] = []byte("0123456789123456")
	fdxt.Keys[1] = []byte("0123456789123456")
	fdxt.Keys[2] = []byte("0123456789123456")
	fdxt.Keys[3] = []byte("0123456789123456")
	fdxt.Keys[4] = []byte("0123456789123456")
	fdxt.Count = make(map[string]*Counter, 1000000)
	fdxt.CDBXtag = make(map[string]string, 1000000)
	fdxt.CDBTSet = make(map[string]*TsetValue, 1000000)
	fdxt.XSet = make(map[string]int, 1000000)

	// Connect to MongoDB
	PlaintextDB, err = Database.MongoDBSetup(dbName, mongoURI...)
	if err != nil {
		return err
	}
	return nil
}

func (fdxt *FDXT) UpdatePhase() error {
	// Get the MongoDB database
	plaintextDB := PlaintextDB
	defer plaintextDB.Client().Disconnect(context.Background())

	// Get the collection named "keyword_ids" from the MongoDB database
	collection := plaintextDB.Collection("keyword_ids")

	// Create a cursor with no timeout and a batch size of 3000
	ctx := context.TODO()
	opts := options.Find().SetNoCursorTimeout(true).SetBatchSize(3000)
	cur, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return fmt.Errorf("find keyword_ids: %w", err)
	}

	defer cur.Close(ctx)

	var keywordIds []bson.M
	if err = cur.All(ctx, &keywordIds); err != nil {
		return fmt.Errorf("read cursor: %w", err)
	}

	encryptTimeList := make([]time.Duration, 0, 1000000)
	cipherList := make([]int, 0, 1000000)
	for _, keywordId := range keywordIds {
		valSet, ok := keywordId["val_set"].(primitive.A)
		if !ok {
			return fmt.Errorf("val_set field is not an array")
		}
		var ids []string
		for _, v := range valSet {
			if str, ok := v.(string); ok {
				ids = append(ids, str)
			} else {
				return fmt.Errorf("val_set contains non-string value")
			}
		}
		keyword := keywordId["k"].(string)

		encryptTime, err := fdxt.Encrypt(keyword, ids, Add)
		if err != nil {
			log.Println(err)
		}
		encryptTimeList = append(encryptTimeList, encryptTime)
		cipherList = append(cipherList, len(ids))
	}
	// save to file
	// Write fdxt.Count.max to file
	if err := FDXTSaveFileCntToFile(fdxt.Count, "./cmd/FDXT/configs/filecnt.json"); err != nil {
		log.Println("Error saving filecnt to file:", err)
		return err
	}
	saveTime := time.Now()
	resultpath := filepath.Join("result", "Update", "FDXT", fmt.Sprintf("%s.csv", saveTime.Format("2006-01-02_15-04-05")))
	resultHeader := []string{"encryptTime", "cipherLength"}
	resultData := make([][]string, len(cipherList))
	for i, cipher := range cipherList {
		resultData[i] = []string{encryptTimeList[i].String(), strconv.Itoa(cipher)}
	}
	err = utils.WriteResultToCSV(resultpath, resultHeader, resultData)
	if err != nil {
		log.Println("Error writing result to file:", err)
		return err
	}

	return nil
}

func (fdxt *FDXT) Encrypt(keyword string, ids []string, op Operation) (time.Duration, error) {
	kw, kt, kx, ky, kz := fdxt.Keys[0], fdxt.Keys[1], fdxt.Keys[2], fdxt.Keys[3], fdxt.Keys[4]
	if _, ok := fdxt.Count[keyword]; !ok {
		fdxt.Count[keyword] = &Counter{srch: 0, updt: 0, Max: 0}
	}
	clientTime := time.Duration(0)
	for _, id := range ids {
		start := time.Now()
		fdxt.Count[keyword].updt++
		fdxt.Count[keyword].Max++
		msgLen := len(keyword) + len(big.NewInt(int64(fdxt.Count[keyword].Max)).Bytes()) + 1
		msg := make([]byte, 0, msgLen)
		msg = append(msg, []byte(keyword)...)
		msg = append(msg, big.NewInt(int64(fdxt.Count[keyword].Max)).Bytes()...)

		addr, _ := utils.PrfF(kw, append(msg, byte(0)))
		val, _ := utils.PrfF(kw, append(msg, byte(1)))
		
		val, err = utils.BytesXORWithOp(val, []byte(id), int(op))
		if err != nil {
			return 0, fmt.Errorf("BytesXORWithOp: %w", err)
		}
		
		msgLen = len(keyword) + len(big.NewInt(int64(fdxt.Count[keyword].srch)).Bytes()) + len(big.NewInt(int64(fdxt.Count[keyword].updt)).Bytes()) + 1
		msg = make([]byte, 0, msgLen)
		msg = append(msg, []byte(keyword)...)
		msg = append(msg, big.NewInt(int64(fdxt.Count[keyword].srch)).Bytes()...)
		msg = append(msg, big.NewInt(int64(fdxt.Count[keyword].updt)).Bytes()...)
		l, _ := utils.PrfF(kt, append(msg, byte(0)))
		t, _ := utils.PrfF(kt, append(msg, byte(1)))
		xtag1, _ := pbcUtil.PrfToZr(kx, []byte(keyword))
		xtag2, _ := pbcUtil.PrfToZr(ky, append([]byte(id), byte(op)))
		xtag := pbcUtil.GToPower2(xtag1, xtag2)
		c := utils.BytesXOR(xtag.Bytes(), t)
		alpha, _, err := utils.ComputeAlpha(ky, kz, []byte(id), int(op), append([]byte(keyword), big.NewInt(int64(fdxt.Count[keyword].Max)).Bytes()...))
		if err != nil {
			return 0, err
		}
		clientTime += time.Since(start)

		// server part
		fdxt.CDBXtag[string(l)] = string(c)
		fdxt.CDBTSet[string(addr)] = &TsetValue{Val: string(val), Alpha: alpha}
	}
	return clientTime, nil
}

func (fdxt *FDXT) SearchPhase(tableName, fileName string) error {
	fileName = "./cmd/FDXT/configs/" + fileName
	keywordsList := utils.QueryKeywordsFromFile(fileName)

	// Initialize result list
	resultList := make([][]string, 0, len(keywordsList)+1)
	clientSearchTime := make([]time.Duration, 0, len(keywordsList)+1)
	serverTimeList := make([]time.Duration, 0, len(keywordsList)+1)
	resultLengthList := make([]int, 0, len(keywordsList)+1)
	totalTimeList := make([]time.Duration, 0, len(keywordsList)+1)
	payloadSizeList := make([]int, 0, len(keywordsList)+1)
	w1CounterList := make([]int, 0, len(keywordsList)+1)
	w2CounterList := make([]int, 0, len(keywordsList)+1)

	for _, keywords := range keywordsList {
		clientTimeTotal := time.Duration(0)
		serverTimeTotal := time.Duration(0)
		totalStart := time.Now()
		w1CounterList = append(w1CounterList, fdxt.Count[keywords[0]].Max)
		w2CounterList = append(w2CounterList, fdxt.Count[keywords[1]].Max)

		// client search step 1
		w1 := keywords[0]
		start := time.Now()
		tkl, stkl, xtkList, err := fdxt.ClientSearchStep1(w1, keywords)
		if err != nil {
			log.Println(err)
			return err
		}
		clientTimeTotal += time.Since(start)

		// server search step
		start = time.Now()
		resList := fdxt.ServerSearch(len(keywords), tkl, stkl, xtkList)
		serverTimeTotal += time.Since(start)
		

		// client search step 2
		start = time.Now()
		sIdList := fdxt.ClientSearchStep2(w1, keywords, resList)
		clientTimeTotal += time.Since(start)
		
		// Add results to the result list
		payloadSize := CalculateResListSize(resList)
		totalTimeList = append(totalTimeList, time.Since(totalStart))
		payloadSizeList = append(payloadSizeList, payloadSize)
		resultList = append(resultList, sIdList)
		clientSearchTime = append(clientSearchTime, clientTimeTotal)
		serverTimeList = append(serverTimeList, serverTimeTotal)
		resultLengthList = append(resultLengthList, len(sIdList))
	}

	// Set the path and name of the result file
	resultpath := filepath.Join("result", "Search", "FDXT", tableName, time.Now().Format("2006-01-02_15-04-05")+"w2_keywords_2.csv")

	// Define the result header
	resultHeader := []string{"keyword", "clientTime", "serverTime", "totalTime", "resultLength", "payloadSize", "w1", "w2"}

	// Organize result data into tabular form
	resultData := make([][]string, len(resultList))
	for i, keywords := range keywordsList {
		resultData[i] = []string{strings.Join(keywords, "#"), strconv.Itoa(int(clientSearchTime[i].Microseconds())), strconv.Itoa(int(serverTimeList[i].Microseconds())), strconv.Itoa(int(totalTimeList[i].Microseconds())), strconv.Itoa(resultLengthList[i]), strconv.Itoa(payloadSizeList[i]), strconv.Itoa(w1CounterList[i]), strconv.Itoa(w2CounterList[i])}
	}

	log.Println("Write Time:", time.Now().Format("2006-01-02_15-04-05"))
	err := utils.WriteResultToCSV(resultpath, resultHeader, resultData)
	if err != nil {
		return fmt.Errorf("write results CSV: %w", err)
	}
	return nil
}

func CalculateResListSize(resList []*RES) int {
	size := 0
	for _, res := range resList {
		size += len(res.Val)
		size += 4
	}
	return size
}

func FDXTSaveFileCntToFile(fileCnt map[string]*Counter, filename string) error {
	// Create the file; if the directory does not exist, create it first
	dir := filepath.Dir(filename)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}

	// Create the file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write only the max field of fileCnt to the JSON file
	maxValues := make(map[string]int)
	for key, counter := range fileCnt {
		maxValues[key] = counter.Max
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(maxValues)
}
