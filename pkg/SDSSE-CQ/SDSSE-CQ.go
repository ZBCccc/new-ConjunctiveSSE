package sdssecq

import (
	"ConjunctiveSSE/pkg/Database"
	sdssecqClient "ConjunctiveSSE/pkg/SDSSE-CQ/Client"
	"ConjunctiveSSE/pkg/utils"
	"context"
	"fmt"
	"log"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	util "github.com/ZBCccc/Aura/Util"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	// PlaintextDB is the MongoDB database for storing plaintext data.
	PlaintextDB *mongo.Database
	client      *sdssecqClient.Client
)

func Init(dbName string, mongoURI string) error {
	// Initialize Client
	client = sdssecqClient.NewClient()

	// Connect to MongoDB
	var err error
	PlaintextDB, err = Database.MongoDBSetup(dbName, mongoURI)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func CiphertextGenPhase(dbName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	defer PlaintextDB.Client().Disconnect(ctx)

	collection := PlaintextDB.Collection("keyword_ids")
	// Get total count first to pre-allocate space
	count, err := collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("count documents failed: %w", err)
	}

	// Pre-allocate using actual count
	encryptTimeList := make([]time.Duration, 0, count)
	keywordList := make([]string, 0, count)
	volumeList := make([]int, 0, count)

	opts := options.Find().SetNoCursorTimeout(true).SetBatchSize(3000)
	cur, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return fmt.Errorf("find collection failed: %w", err)
	}
	defer cur.Close(ctx)

	var keywordIds []bson.M
	if err = cur.All(ctx, &keywordIds); err != nil {
		return fmt.Errorf("read cursor failed: %w", err)
	}

	// Read all records
	for _, keywordId := range keywordIds {
		valSet, ok := keywordId["val_set"].(primitive.A)
		if !ok {
			log.Println("val_set is not of type primitive.A")
			continue
		}
		ids := make([]string, 0, len(valSet))
		for _, v := range valSet {
			if str, ok := v.(string); ok {
				ids = append(ids, str)
			} else {
				log.Println("val_set contains non-string value")
				continue
			}
		}
		ids = utils.RemoveDuplicates(ids)
		keyword := keywordId["k"].(string)
		start := time.Now()
		for _, id := range ids {
			client.Update(util.Insert, keyword, id)
		}
		encryptTime := time.Since(start)

		encryptTimeList = append(encryptTimeList, encryptTime)
		keywordList = append(keywordList, keyword)
		volumeList = append(volumeList, len(ids))
	}
	saveTime := time.Now()

	// Set the path and name of the result file
	resultPath := filepath.Join("result", "Update", "SDSSE-CQ", dbName, fmt.Sprintf("%s.csv", saveTime.Format("2006-01-02_15-04-05")))

	// Define the result header
	resultHeader := []string{"keyword", "volume", "addTime"}

	// Organize result data into tabular form
	resultData := make([][]string, len(keywordList))
	for i, keyword := range keywordList {
		resultData[i] = []string{keyword, strconv.Itoa(volumeList[i]), strconv.Itoa(int(encryptTimeList[i].Microseconds()))}
	}

	// Write results to file
	err = utils.WriteResultToCSV(resultPath, resultHeader, resultData)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func SearchPhase(tableName, fileName string) {
	fileName = "./cmd/SDSSE-CQ/configs/" + fileName
	keywordsList := utils.QueryKeywordsFromFile(fileName)

	// Initialize result list
	resultList := make([][]string, 0, len(keywordsList)+1)
	clientTimeList := make([]time.Duration, 0, len(keywordsList)+1)
	serverTimeList := make([]time.Duration, 0, len(keywordsList)+1)
	serverAuraTimeList := make([]time.Duration, 0, len(keywordsList)+1)
	resultLengthList := make([]int, 0, len(keywordsList)+1)
	totalTimeList := make([]time.Duration, 0, len(keywordsList)+1)
	w1CounterList := make([]int, 0, len(keywordsList)+1)
	w2CounterList := make([]int, 0, len(keywordsList)+1)
	// payloadSizeList := make([]int, 0, len(keywordsList)+1)

	// Search loop
	for _, keywords := range keywordsList {
		totalStart := time.Now()
		// find w1's lens
		// Select the keyword with the lowest query frequency
		counter := math.MaxInt
		for _, w := range keywords {
			num := client.CT[w]
			if num < counter {
				counter = num
			}
		}
		w1CounterList = append(w1CounterList, client.CT[keywords[0]]+1)
		w2CounterList = append(w2CounterList, client.CT[keywords[1]]+1)
		result, clientTime, serverTime, serverAuraTime := client.Search(keywords)
		totalTime := time.Since(totalStart)

		// Add results to the result list
		resultList = append(resultList, result)
		resultLengthList = append(resultLengthList, len(result))
		totalTimeList = append(totalTimeList, totalTime)
		clientTimeList = append(clientTimeList, clientTime)
		serverTimeList = append(serverTimeList, serverTime)
		serverAuraTimeList = append(serverAuraTimeList, serverAuraTime)
	}

	// Set the path and name of the result file
	resultPath := filepath.Join("result", "Search", "SDSSE-CQ", tableName, "w1_keywords_2")

	// Define the result header
	resultHeader := []string{"keyword", "clientTime", "serverTime", "serverAuraTime", "totalTime", "resultLength", "w1", "w2"}

	// Organize result data into tabular form
	resultData := make([][]string, len(resultList))
	for i, keywords := range keywordsList {
		resultData[i] = []string{strings.Join(keywords, "#"), strconv.Itoa(int(clientTimeList[i].Microseconds())), strconv.Itoa(int(serverTimeList[i].Microseconds())), strconv.Itoa(int(serverAuraTimeList[i].Microseconds())), strconv.Itoa(int(totalTimeList[i].Microseconds())), strconv.Itoa(resultLengthList[i]), strconv.Itoa(w1CounterList[i]), strconv.Itoa(w2CounterList[i])}
	}

	err := utils.WriteResultToCSV(resultPath, resultHeader, resultData)
	if err != nil {
		log.Println(err)
	}
}
