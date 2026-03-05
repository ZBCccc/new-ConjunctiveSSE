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
	"go.mongodb.org/mongo-driver/mongo"
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

	hdxt := c.GetHDXT()
	plaintextDB := hdxt.PlaintextDB
	defer plaintextDB.Client().Disconnect(context.Background())

	// Setup Phase
	log.Println("Setup Phase Start")
	setupTime := time.Now()
	SetupPhase(c, plaintextDB, *dbName)
	log.Printf("Setup phase completed in %v\n", time.Since(setupTime))

	// Update Phase
	log.Println("Update Phase Start")
	updateTime := time.Now()
	UpdatePhase(c, plaintextDB, *dbName)
	log.Printf("Update phase completed in %v\n", time.Since(updateTime))

	// Search Phase
	log.Println("Search Phase Start")
	searchTime := time.Now()
	SearchPhase(c, *dbName)
	log.Printf("Search phase completed in %v\n", time.Since(searchTime))
}

func SetupPhase(c *client.HDXTClient, plaintextDB *mongo.Database, dbName string) {
	hdxt := c.GetHDXT()

	collection := plaintextDB.Collection("id_keywords")
	ctx := context.TODO()

	opts := options.Find().SetNoCursorTimeout(true).SetBatchSize(3000)
	cur, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		log.Fatal("Error getting collection:", err)
	}
	defer cur.Close(ctx)

	var idKeywords []bson.M
	if err = cur.All(ctx, &idKeywords); err != nil {
		log.Fatal("Error getting keywordIds:", err)
	}

	idKeywordsSetup := idKeywords[:len(idKeywords)/2]
	idList := make([]string, 0, len(idKeywordsSetup))
	volumeList := make([]int, 0, len(idKeywordsSetup))
	setupTimeList := make([]time.Duration, 0, len(idKeywordsSetup))

	for _, idKeyword := range idKeywordsSetup {
		keywords := extractKeywords(idKeyword)
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

	// Send to server
	c.Setup(hdxt.MitraCipherList, hdxt.AuhmeCipherList)

	// Save results
	saveResults("Setup", "HDXT", dbName, idList, volumeList, setupTimeList)
}

func UpdatePhase(c *client.HDXTClient, plaintextDB *mongo.Database, dbName string) {
	collection := plaintextDB.Collection("id_keywords")
	ctx := context.TODO()

	opts := options.Find().SetNoCursorTimeout(true).SetBatchSize(3000)
	cur, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		log.Fatal("Error getting collection:", err)
	}
	defer cur.Close(ctx)

	var idKeywords []bson.M
	if err = cur.All(ctx, &idKeywords); err != nil {
		log.Fatal("Error getting keywordIds:", err)
	}

	idKeywordsUpdate := idKeywords[len(idKeywords)/2:]
	idList := make([]string, 0, len(idKeywordsUpdate))
	volumeList := make([]int, 0, len(idKeywordsUpdate))
	updateTimeList := make([]time.Duration, 0, len(idKeywordsUpdate))

	for _, idKeyword := range idKeywordsUpdate {
		keywords := extractKeywords(idKeyword)
		id := idKeyword["id"].(string)

		encryptTime := time.Now()
		err = c.Update(id, keywords, HDXT.Add)
		if err != nil {
			log.Fatal("Error in Update:", err)
		}

		updateTimeList = append(updateTimeList, time.Since(encryptTime))
		idList = append(idList, id)
		volumeList = append(volumeList, len(keywords))
	}

	// Save results
	saveResults("Update", "HDXT", dbName, idList, volumeList, updateTimeList)
}

func SearchPhase(c *client.HDXTClient, tableName string) {
	fileName := "./cmd/HDXT/configs/keywords_2.txt"
	keywordsList := utils.QueryKeywordsFromFile(fileName)

	clientSearchTime := make([]time.Duration, 0, len(keywordsList))
	resultLengthList := make([]int, 0, len(keywordsList))

	for _, keywords := range keywordsList {
		searchTime := time.Now()
		sIdList, err := c.Search(keywords)
		if err != nil {
			log.Fatal("Error in Search:", err)
		}

		clientSearchTime = append(clientSearchTime, time.Since(searchTime))
		resultLengthList = append(resultLengthList, len(sIdList))
	}

	// Save results
	resultpath := filepath.Join("result", "Search", "HDXT", tableName, fmt.Sprintf("%s.csv", time.Now().Format("2006-01-02_15-04-05")))
	resultHeader := []string{"keyword", "clientTime", "resultLength"}
	resultData := make([][]string, len(keywordsList))

	for i, keywords := range keywordsList {
		resultData[i] = []string{strings.Join(keywords, "#"), strconv.Itoa(int(clientSearchTime[i].Microseconds())), strconv.Itoa(resultLengthList[i])}
	}

	err := utils.WriteResultToCSV(resultpath, resultHeader, resultData)
	if err != nil {
		log.Fatal(err)
	}
}

func extractKeywords(idKeyword bson.M) []string {
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
	return utils.RemoveDuplicates(keywords)
}

func saveResults(phase, scheme, dbName string, idList []string, volumeList []int, timeList []time.Duration) {
	saveTime := time.Now()
	resultpath := filepath.Join("result", phase, scheme, dbName, fmt.Sprintf("%s.csv", saveTime.Format("2006-01-02_15-04-05")))
	resultHeader := []string{"id", "volume", "addTime"}
	resultData := make([][]string, len(idList))

	for i, id := range idList {
		resultData[i] = []string{id, strconv.Itoa(volumeList[i]), strconv.Itoa(int(timeList[i].Microseconds()))}
	}

	err := utils.WriteResultToCSV(resultpath, resultHeader, resultData)
	if err != nil {
		log.Fatal("Error writing result to file:", err)
	}
}
