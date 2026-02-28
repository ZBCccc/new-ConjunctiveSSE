package main

import (
	"ConjunctiveSSE/pkg/HDXT"
	"ConjunctiveSSE/pkg/HDXT/client"
	"ConjunctiveSSE/pkg/utils"
	"context"
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	serverAddr := flag.String("server", "localhost:50051", "gRPC server address (host:port)")
	dbName := flag.String("db", "Crime_USENIX_REV", "MongoDB database name")
	mongoURI := flag.String("mongo", "mongodb://localhost:27017", "MongoDB connection URI")
	flag.Parse()

	c, err := client.NewHDXTClient(*serverAddr, *dbName, *mongoURI)
    if err != nil {
        log.Fatalf("failed to create client: %v", err)
    }
	defer c.Close()
    
    // Execute experiment
	// Get MongoDB database
	hdxt := c.GetHDXT()
	plaintextDB := hdxt.PlaintextDB
	defer plaintextDB.Client().Disconnect(context.Background())

	// Initialize
	setupTimeList := make([]time.Duration, 0, 1000000)

	// Get collection named 'id_keywords' from MongoDB database
	collection := plaintextDB.Collection("id_keywords")

	// Create a cursor with no timeout and batch size of 3000
	ctx := context.TODO()
	opts := options.Find().SetNoCursorTimeout(true).SetBatchSize(3000)
	cur, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		log.Fatal("Error getting collection:", err)
	}

	// Close cursor
	defer cur.Close(ctx)

	// Read all records from cursor
	var idKeywords []bson.M
	if err = cur.All(ctx, &idKeywords); err != nil {
		log.Fatal("Error getting keywordIds:", err)
	}
    // 1. Setup Phase
	log.Println("Setup Phase Start")
	setupTime := time.Now()
	idKeywordsSetup := idKeywords[:len(idKeywords)/2]
	idList := make([]string, 0, len(idKeywords)/2)
	volumeList := make([]int, 0, len(idKeywords)/2)
	for _, idKeyword := range idKeywordsSetup {
		valSet, ok := idKeyword["val_st"].(primitive.A)
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
		keywords = utils.RemoveDuplicates(keywords) // Deduplicate keywords
		id := idKeyword["id"].(string)
		encryptTime := time.Now()
		_, err = hdxt.Setup(id, keywords, HDXT.Add)
		if err != nil {
			log.Fatal("Error in Setup:", err)
		}

		setupTimeList = append(setupTimeList, time.Since(encryptTime))
		idList = append(idList, id)
		volumeList = append(volumeList, len(keywords))
	}
	// Send Setup request
	log.Println("Sending Setup Request, cost time:", time.Since(setupTime))
	c.Setup(hdxt.MitraCipherList, hdxt.AuhmeCipherList)
	// save to file
	saveTime := time.Now()
	resultpath := filepath.Join("result", "Setup", "HDXT", fmt.Sprintf("%s.csv", saveTime.Format("2006-01-02_15-04-05")))
	resultHeader := []string{"id", "volume", "addTime"}
	resultData := make([][]string, len(idList))
	for i, id := range idList {
		resultData[i] = []string{id, strconv.Itoa(volumeList[i]), strconv.Itoa(int(setupTimeList[i].Microseconds()))}
	}
	err = utils.WriteResultToCSV(resultpath, resultHeader, resultData)
	if err != nil {
		log.Fatal("Error writing result to file:", err)
	}

	// 2. Update Phase
	log.Println("Update Phase Start")
	idList = make([]string, 0, len(idKeywords)/2)
	idKeywordsUpdate := idKeywords[len(idKeywords)/2:]
	volumeList = make([]int, 0, len(idKeywords)/2)
	for _, idKeyword := range idKeywordsUpdate {
		valSet, ok := idKeyword["val_st"].(primitive.A)
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
		keywords = utils.RemoveDuplicates(keywords) // Deduplicate keywords
		id := idKeyword["id"].(string)
		encryptTime := time.Now()
		err = c.Update(id, keywords, HDXT.Add)
		if err != nil {
			log.Fatal("Error in Encrypt:", err)
		}

		// save to []
		setupTimeList = append(setupTimeList, time.Since(encryptTime))
		idList = append(idList, id)
		volumeList = append(volumeList, len(keywords))
	}
	saveTime = time.Now()

	// save to file
	resultpath = filepath.Join("result", "Update", "HDXT", fmt.Sprintf("%s.csv", saveTime.Format("2006-01-02_15-04-05")))

	// Define result header
	resultHeader = []string{"id", "volume", "addTime"}

	// Organize result data into tabular form
	resultData = make([][]string, len(idList))
	for i, id := range idList {
		resultData[i] = []string{id, strconv.Itoa(volumeList[i]), strconv.Itoa(int(setupTimeList[i].Microseconds()))}
	}

	// Write results to file
	err = utils.WriteResultToCSV(resultpath, resultHeader, resultData)
	if err != nil {
		log.Fatal("Error writing result to file:", err)
	}

	// 3. Search Phase
	log.Println("Search Phase Start")
	SearchPhase(c, *dbName, "keywords_2.txt")
}

func SearchPhase(c *client.HDXTClient, tableName, fileName string) {
	fileName = "./cmd/HDXT/configs/" + fileName
	keywordsList := utils.QueryKeywordsFromFile(fileName)
	// Initialize result list
	clientSearchTime := make([]time.Duration, 0, len(keywordsList)+1)
	resultLengthList := make([]int, 0, len(keywordsList)+1)

	// Search loop
	for _, keywords := range keywordsList {
		searchTime := time.Now()
		sIdList, err := c.Search(keywords)
		if err != nil {
			log.Fatal("Error in Search:", err)
		}

		// Add results to result list
		clientSearchTime = append(clientSearchTime, time.Since(searchTime))
		resultLengthList = append(resultLengthList, len(sIdList))
	}
	log.Println("Search Phase End")

	// Set result file path and name
	resultpath := filepath.Join("result", "Search", "HDXT", tableName, fmt.Sprintf("%s.csv", time.Now().Format("2006-01-02_15-04-05")))

	// Define result header
	resultHeader := []string{"keyword", "clientTime", "resultLength"}

	// Organize result data into tabular form
	resultData := make([][]string, len(keywordsList))
	for i, keywords := range keywordsList {
		resultData[i] = []string{strings.Join(keywords, "#"), strconv.Itoa(int(clientSearchTime[i].Microseconds())), strconv.Itoa(resultLengthList[i])}
	}

	// Write results to file
	err := utils.WriteResultToCSV(resultpath, resultHeader, resultData)
	if err != nil {
		log.Fatal(err)
	}
}