package main

import (
	"ConjunctiveSSE/pkg/Database"
	"ConjunctiveSSE/pkg/utils"
	"context"
	"log"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	PlaintextDB, err := Database.MongoDBSetup("Crime_USENIX_REV")
	if err != nil {
		log.Println(err)
		return
	}
	defer PlaintextDB.Client().Disconnect(context.TODO())

	collection := PlaintextDB.Collection("id_keywords")

	// Create a cursor with no timeout and batch size of 1000
	ctx := context.TODO()
	opts := options.Find().SetNoCursorTimeout(true).SetBatchSize(1000)
	cur, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		log.Fatal(err)
	}

	// Close cursor
	defer cur.Close(ctx)

	// Read all records from cursor
	var idKeywords []bson.M
	if err = cur.All(ctx, &idKeywords); err != nil {
		log.Fatal(err)
	}

	var keywordsList []string
	for _, idKeyword := range idKeywords {
		valSet, ok := idKeyword["val_set"].(primitive.A)
		if !ok {
			log.Println("val_set is not of type primitive.A")
			return 
		}
		var keywords []string
		for _, v := range valSet {
			if str, ok := v.(string); ok {
				keywords = append(keywords, str)
			} else {
				log.Println("val_set contains non-string value")
				return
			}
		}
		keywords = utils.RemoveDuplicates(keywords)
		keywordsList = append(keywordsList, keywords...)
	}

	keywordsList = utils.RemoveDuplicates(keywordsList)
	// Randomly select 2 keywords from keywordsList to form numPairs pairs
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	keywordsPair := make([][]string, 10000)
	for i := 0; i < 10000; i++ {
		// Create a new slice to avoid duplicates
		shuffledKeywords := make([]string, len(keywordsList))
		copy(shuffledKeywords, keywordsList)

		// Randomly select two different keywords
		r.Shuffle(len(shuffledKeywords), func(i, j int) {
			shuffledKeywords[i], shuffledKeywords[j] = shuffledKeywords[j], shuffledKeywords[i]
		})
		keywordsPair[i] = shuffledKeywords[:2]
	}
	utils.WriteResultToFile("cmd/HDXT/configs/keywords_2.txt", keywordsPair)

	// Randomly select 6 keywords from keywordsList to form numPairs pairs
	keywordsSix := make([][]string, 10000)
	for i := 0; i < 10000; i++ {
		// Create a new slice to avoid duplicates
		shuffledKeywords := make([]string, len(keywordsList))
		copy(shuffledKeywords, keywordsList)

		// Randomly select two different keywords
		r.Shuffle(len(shuffledKeywords), func(i, j int) {
			shuffledKeywords[i], shuffledKeywords[j] = shuffledKeywords[j], shuffledKeywords[i]
		})
		keywordsSix[i] = shuffledKeywords[:6]
	}
	utils.WriteResultToFile("cmd/HDXT/configs/keywords_6.txt", keywordsSix)
}