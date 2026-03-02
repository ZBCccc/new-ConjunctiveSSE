package main

import (
	"ConjunctiveSSE/pkg/Database"
	sdssecqClient "ConjunctiveSSE/pkg/SDSSE-CQ/Client"
	"ConjunctiveSSE/pkg/utils"
	"context"
	"flag"
	"fmt"
	"log"
	"math"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	util "github.com/ZBCccc/Aura/Util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	serverAddr := flag.String("server", "localhost:50051", "gRPC server address (host:port)")
	dbName := flag.String("db", "Crime_USENIX_REV", "MongoDB database name")
	mongoURI := flag.String("mongo", "mongodb://localhost:27017", "MongoDB connection URI")
	phase := flag.String("phase", "cs", "Phase to run: c=ciphertext generation, s=search")
	group := flag.String("group", "keywords_2.txt", "Search group file name")
	flag.Parse()

	fmt.Println("*********************************************")
	fmt.Println("Test_on:", *dbName)
	fmt.Println("Phase:", *phase)
	fmt.Println("Server:", *serverAddr)
	fmt.Println("*********************************************")

	// Initialize MongoDB
	plaintextDB, err := Database.MongoDBSetup(*dbName, *mongoURI)
	if err != nil {
		log.Fatal("Error initializing MongoDB:", err)
	}
	defer plaintextDB.Client().Disconnect(context.Background())

	// Create gRPC client
	c, err := sdssecqClient.NewSDSSEcqClient(*serverAddr)
	if err != nil {
		log.Fatal("Error creating gRPC client:", err)
	}
	defer c.Close()

	// Run ciphertext generation phase
	if strings.Contains(*phase, "c") {
		err := CiphertextGenPhase(c, plaintextDB, *dbName)
		if err != nil {
			log.Fatal("Error in CiphertextGenPhase:", err)
		}
	}

	// Run search phase
	if strings.Contains(*phase, "s") {
		SearchPhase(c, *dbName, *group)
	}
}

func CiphertextGenPhase(c *sdssecqClient.SDSSEcqClient, plaintextDB *mongo.Database, dbName string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	collection := plaintextDB.Collection("keyword_ids")
	count, err := collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("count documents failed: %w", err)
	}

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
			err := c.Update(util.Insert, keyword, id)
			if err != nil {
				log.Printf("Error updating keyword %s, id %s: %v", keyword, id, err)
			}
		}
		encryptTime := time.Since(start)

		encryptTimeList = append(encryptTimeList, encryptTime)
		keywordList = append(keywordList, keyword)
		volumeList = append(volumeList, len(ids))
	}

	// Save results
	saveTime := time.Now()
	resultPath := filepath.Join("result", "Update", "SDSSE-CQ", dbName, fmt.Sprintf("%s.csv", saveTime.Format("2006-01-02_15-04-05")))
	resultHeader := []string{"keyword", "volume", "addTime"}
	resultData := make([][]string, len(keywordList))
	for i, keyword := range keywordList {
		resultData[i] = []string{keyword, strconv.Itoa(volumeList[i]), strconv.Itoa(int(encryptTimeList[i].Microseconds()))}
	}
	err = utils.WriteResultToCSV(resultPath, resultHeader, resultData)
	if err != nil {
		log.Println(err)
		return err
	}

	fmt.Println("CiphertextGenPhase completed")
	return nil
}

func SearchPhase(c *sdssecqClient.SDSSEcqClient, tableName, fileName string) {
	fileName = "./cmd/SDSSE-CQ/configs/" + fileName
	keywordsList := utils.QueryKeywordsFromFile(fileName)

	resultList := make([][]string, 0, len(keywordsList)+1)
	clientTimeList := make([]time.Duration, 0, len(keywordsList)+1)
	resultLengthList := make([]int, 0, len(keywordsList)+1)
	w1CounterList := make([]int, 0, len(keywordsList)+1)
	w2CounterList := make([]int, 0, len(keywordsList)+1)

	ct := c.GetCT()

	for _, keywords := range keywordsList {
		// Find w1's count
		counter := math.MaxInt
		for _, w := range keywords {
			num := ct[w]
			if num < counter {
				counter = num
			}
		}
		w1CounterList = append(w1CounterList, ct[keywords[0]]+1)
		w2CounterList = append(w2CounterList, ct[keywords[1]]+1)

		result, clientTime, err := c.Search(keywords)
		if err != nil {
			log.Printf("Error searching %v: %v", keywords, err)
			continue
		}

		resultList = append(resultList, result)
		resultLengthList = append(resultLengthList, len(result))
		clientTimeList = append(clientTimeList, clientTime)
	}

	// Save results
	resultPath := filepath.Join("result", "Search", "SDSSE-CQ", tableName, "w1_keywords_2")
	resultHeader := []string{"keyword", "clientTime", "resultLength", "w1", "w2"}
	resultData := make([][]string, len(resultList))
	for i, keywords := range keywordsList {
		resultData[i] = []string{strings.Join(keywords, "#"), strconv.Itoa(int(clientTimeList[i].Microseconds())), strconv.Itoa(resultLengthList[i]), strconv.Itoa(w1CounterList[i]), strconv.Itoa(w2CounterList[i])}
	}

	err := utils.WriteResultToCSV(resultPath, resultHeader, resultData)
	if err != nil {
		log.Println(err)
	}

	fmt.Println("SearchPhase completed")
}
