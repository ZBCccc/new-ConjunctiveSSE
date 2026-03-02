package main

import (
	"ConjunctiveSSE/pkg/Database"
	"ConjunctiveSSE/pkg/FDXT"
	"ConjunctiveSSE/pkg/FDXT/client"
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
	flag.Parse()

	c, err := client.NewFDXTClient(*serverAddr)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	// Execute experiment
	dbName := "Crime_USENIX_REV"
	updateTime := time.Now()
	UpdatePhase(c, dbName)
	log.Printf("Update phase to %v\n", time.Since(updateTime))

	// Search Phase
	searchTime := time.Now()
	SearchPhase(c, dbName)
	log.Printf("Search phase to %v\n", time.Since(searchTime))
}

func UpdatePhase(c *client.FDXTClient, dbName string) {
	PlaintextDB, err := Database.MongoDBSetup(dbName)
	if err != nil {
		log.Fatal(err)
	}
	defer PlaintextDB.Client().Disconnect(context.Background())

	// Update Phase
	// Get collection named 'keyword_ids' from MongoDB database
	collection := PlaintextDB.Collection("keyword_ids")

	// Create a cursor with no timeout and batch size of 3000
	ctx := context.TODO()
	opts := options.Find().SetNoCursorTimeout(true).SetBatchSize(3000)
	cur, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		log.Fatal(err)
	}

	// Close cursor
	defer cur.Close(ctx)

	// Read all records from cursor
	var keywordIds []bson.M
	if err = cur.All(ctx, &keywordIds); err != nil {
		log.Fatal(err)
	}

	// Read all records
	encryptTimeList := make([]time.Duration, 0, 1000000)
	cipherList := make([]int, 0, 1000000)
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

		encryptTime, err := c.GetFDXT().Encrypt(keyword, ids, FDXT.Add)
		if err != nil {
			log.Println(err)
		}
		encryptTimeList = append(encryptTimeList, encryptTime)
		cipherList = append(cipherList, len(ids))
	}
	// send to server
	log.Println("Start to send to server")
	if err = c.Update(c.GetFDXT().CDBXtag, c.GetFDXT().CDBTSet); err != nil {
		log.Fatal(err)
	}
	// save to file
	saveTime := time.Now()
	resultPath := filepath.Join("result", "Update", "FDXT", fmt.Sprintf("%s.csv", saveTime.Format("2006-01-02_15-04-05")))
	resultHeader := []string{"encryptTime", "cipherLength"}
	resultData := make([][]string, len(cipherList))
	for i, cipher := range cipherList {
		resultData[i] = []string{encryptTimeList[i].String(), strconv.Itoa(cipher)}
	}
	err = utils.WriteResultToCSV(resultPath, resultHeader, resultData)
	if err != nil {
		log.Fatal("Error writing result to file:", err)
	}
}

func SearchPhase(c *client.FDXTClient, dbName string) {
	fileName := "./cmd/ODXT/configs/" + "keywords_2.txt"
	keywordsList := utils.QueryKeywordsFromFile(fileName)

	// Initialize result list
	resultList := make([][]string, 0, len(keywordsList)+1)
	resultLengthList := make([]int, 0, len(keywordsList)+1)
	totalTimeList := make([]time.Duration, 0, len(keywordsList)+1)

	// Search loop
	for _, keywords := range keywordsList {
		totalStart := time.Now()
		sIdList, err := c.Search(keywords)
		if err != nil {
			log.Fatal(err)
		}
		totalTime := time.Since(totalStart)

		// Add results to result list
		resultList = append(resultList, sIdList)
		totalTimeList = append(totalTimeList, totalTime) // totalTimeList = totalTime
		resultLengthList = append(resultLengthList, len(sIdList))
	}

	// Set result file path and name
	resultpath := filepath.Join("result", "Search", "FDXT", dbName, fmt.Sprintf("%s.csv", time.Now().Format("2006-01-02_15-04-05")))

	// Define result header
	resultHeader := []string{"keyword", "totalTime", "resultLength"}

	// Organize result data into tabular form
	resultData := make([][]string, len(resultList))
	for i, keywords := range keywordsList {
		resultData[i] = []string{strings.Join(keywords, "#"), strconv.Itoa(int(totalTimeList[i].Microseconds())), strconv.Itoa(resultLengthList[i])}
	}

	// Write results to file
	err := utils.WriteResultToCSV(resultpath, resultHeader, resultData)
	if err != nil {
		log.Fatal(err)
	}
}
