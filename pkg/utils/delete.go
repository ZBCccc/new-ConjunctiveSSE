package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type deletePair struct {
	Id       string
	Keywords []string
}

type idCount struct {
	Id    string
	Count int
}

func texFileRead(filePath string) ([]idCount, error) {
	// Read the entire file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Read the txt data
	lines := strings.Split(string(data), "\n")
	idCounts := make([]idCount, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		var idCount idCount
		var idInt int
		_, err = fmt.Sscanf(line, "%d:%d", &idInt, &idCount.Count)
		if err != nil {
			return nil, err
		}
		idCount.Id = fmt.Sprintf("%d", idInt)
		idCounts = append(idCounts, idCount)
	}
	return idCounts, nil
}

func GenDeletePairs(filePath string, num int, mongoURI, dbName string) []deletePair {
	// 1. Read .txt file; format is string:int where key=file id, value=keyword count for that file
	idCounts, err := texFileRead(filePath)
	if err != nil {
		log.Fatal(err)
	}

	// 2. Accumulate IDs until the sum of keyword counts reaches num
	var deleteIDs []string
	sum := 0
	for _, val := range idCounts {
		sum += val.Count
		deleteIDs = append(deleteIDs, val.Id)
		if sum >= num {
			break
		}
	}

	// 3. Fetch keywords for each ID from MongoDB
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Fatal(err)
		}
	}()

	db := client.Database(dbName)
	collection := db.Collection("id_keywords")

	deletePairs := make([]deletePair, len(deleteIDs))
	for i := 0; i < len(deleteIDs); i++ {
		deletePairs[i].Id = deleteIDs[i]
		var result bson.M
		err := collection.FindOne(context.TODO(), bson.D{{Key: "id", Value: deleteIDs[i]}}).Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		// Assume val_set is an array (slice)
		if valSet, ok := result["val_st"].(primitive.A); ok {
			// Traverse val_set array to extract string elements
			deletePairs[i].Keywords = make([]string, 0, len(valSet))
			for _, v := range valSet {
				if str, ok := v.(string); ok {
					deletePairs[i].Keywords = append(deletePairs[i].Keywords, str)
				}
			}
		}
	}

	// 4. Return deletePairs
	return deletePairs
}
