package main

import (
	"ConjunctiveSSE/pkg/Database"
	"ConjunctiveSSE/pkg/HDXT"
	"context"
	"slices"
	"strconv"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestDBFind(t *testing.T) {
	db, err := HDXT.MySQLSetup("Crime_USENIX_REV")
	if err != nil {
		t.Fatal(err)
	}

	var cipherTexts []HDXT.CipherText
	db.Where("address = ?", "0x124").Find(&cipherTexts)

	for _, cipherText := range cipherTexts {
		t.Log(cipherText)
	}
}

func TestInsert(t *testing.T) {
	db, err := HDXT.MySQLSetup("Crime_USENIX_REV")
	if err != nil {
		t.Fatal(err)
	}

	var cipherTexts []HDXT.CipherText
	cipherTexts = append(cipherTexts, HDXT.CipherText{
		Address: "0x125",
		Value:   "0x125",
	})
	cipherTexts = append(cipherTexts, HDXT.CipherText{
		Address: "0x126",
		Value:   "0x126",
	})

	err = HDXT.WriteUploadList(db, cipherTexts)
	if err != nil {
		t.Fatal(err)
	}
}

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

	uniqueVals, err := Database.GetUniqueValSets(db)
	if err != nil {
		t.Fatal(err)
	}
	// cover sring slice to int slice
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
