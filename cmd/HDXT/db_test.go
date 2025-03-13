package main

import (
	"ConjunctiveSSE/pkg/Database"
	"context"
	"slices"
	"strconv"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestMongoDB(t *testing.T) {
	db, err := Database.MongoDBSetup("Crime_USENIX_REV")
	if err != nil {
		t.Fatal(err)
	}

	collection := db.Collection("id_keywords")

	// 查询collection中的一个数据
	var result bson.M
	err = collection.FindOne(context.TODO(), bson.D{{Key: "k", Value: "F19"}}).Decode(&result)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(result)
}

func TestMongoDBCount(t *testing.T) {
	db, err := Database.MongoDBSetup("Crime_USENIX_REV")
	if err != nil {
		t.Fatal(err)
	}

	uniqueVals, err := Database.GetUniqueKeywords(db)
	if err != nil {
		t.Fatal(err)
	}
	// cover string slice to int slice
	var intSlice []int
	for _, val := range uniqueVals {
		intVal, err := strconv.Atoi(val)
		if err != nil {
			t.Fatal(err)
		}
		intSlice = append(intSlice, intVal)
	}
	slices.Sort(intSlice)
	t.Log(intSlice)
	t.Log(len(intSlice))
}
