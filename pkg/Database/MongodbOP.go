package Database

import (
	"ConjunctiveSSE/pkg/utils"
	"context"
	"log"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func MongoDBSetup(dbName string, mongoURI ...string) (*mongo.Database, error) {
	uri := "mongodb://localhost:27017"
	if len(mongoURI) > 0 && mongoURI[0] != "" {
		uri = mongoURI[0]
	}
	// Set client options
	clientOptions := options.Client().ApplyURI(uri)

	// Connect to MongoDB
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	// Get the database and collection handles
	PlaintextDB := client.Database(dbName)

	return PlaintextDB, nil
}

func GenQuerydataFromDB(dbName, tableName string, numPairs int) error {
	PlaintextDB, err := MongoDBSetup(dbName)
	if err != nil {
		log.Println(err)
		return err
	}
	defer PlaintextDB.Client().Disconnect(context.TODO())

	collection := PlaintextDB.Collection(tableName)

	// Create a cursor with no timeout and a batch size of 1000
	ctx := context.TODO()
	opts := options.Find().SetNoCursorTimeout(true).SetBatchSize(1000)
	cur, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return err
	}

	defer cur.Close(ctx)

	var keywordIds []bson.M
	if err = cur.All(ctx, &keywordIds); err != nil {
		return err
	}

	var keywordsList []string
	for _, keywordId := range keywordIds {
		keyword := keywordId["k"].(string)
		keywordsList = append(keywordsList, keyword)
	}

	// Randomly select 2 keywords from keywordsList to form numPairs pairs
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	keywordsPair := make([][]string, numPairs)
	for i := 0; i < numPairs; i++ {
		// Create a new slice to avoid duplicates
		shuffledKeywords := make([]string, len(keywordsList))
		copy(shuffledKeywords, keywordsList)

		// Randomly select two different keywords
		r.Shuffle(len(shuffledKeywords), func(i, j int) {
			shuffledKeywords[i], shuffledKeywords[j] = shuffledKeywords[j], shuffledKeywords[i]
		})
		keywordsPair[i] = shuffledKeywords[:2]
	}
	utils.WriteResultToFile("keywords_2.txt", keywordsPair)

	// Randomly select 6 keywords from keywordsList to form numPairs pairs
	keywordsSix := make([][]string, numPairs)
	for i := 0; i < numPairs; i++ {
		// Create a new slice to avoid duplicates
		shuffledKeywords := make([]string, len(keywordsList))
		copy(shuffledKeywords, keywordsList)

		// Randomly select two different keywords
		r.Shuffle(len(shuffledKeywords), func(i, j int) {
			shuffledKeywords[i], shuffledKeywords[j] = shuffledKeywords[j], shuffledKeywords[i]
		})
		keywordsSix[i] = shuffledKeywords[:6]
	}
	utils.WriteResultToFile("keywords_6.txt", keywordsSix)

	return nil
}

func GetUniqueKeywords(PlaintextDB *mongo.Database) ([]string, error) {
	// Get the collection handle
	collection := PlaintextDB.Collection("id_keywords")
	ctx := context.TODO()

	// Use an aggregation pipeline to extract and deduplicate val_set
	pipeline := mongo.Pipeline{
		{{Key: "$unwind", Value: "$val_st"}},                             // Unwind the val_set array
		{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$val_st"}}}}, // Group by val_set value to deduplicate
	}
	// Execute the aggregation query
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Read query results
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Convert results to a string slice
	var uniqueVals []string
	for _, result := range results {
		if val, ok := result["_id"].(string); ok {
			uniqueVals = append(uniqueVals, val)
		}
	}

	return uniqueVals, nil
}

func GetUniqueIDs(PlaintextDB *mongo.Database) ([]string, error) {
	// Get the collection handle
	collection := PlaintextDB.Collection("id_keywords")
	ctx := context.TODO()

	// Use an aggregation pipeline to extract and deduplicate id
	pipeline := mongo.Pipeline{
		{{Key: "$unwind", Value: "$id"}},                             // Unwind the id array
		{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$id"}}}}, // Group by id value to deduplicate
	}
	// Execute the aggregation query
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Read query results
	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	// Convert results to a string slice
	var uniqueVals []string
	for _, result := range results {
		if val, ok := result["_id"].(string); ok {
			uniqueVals = append(uniqueVals, val)
		}
	}

	return uniqueVals, nil
}
