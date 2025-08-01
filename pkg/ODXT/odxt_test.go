package ODXT

import (
	"ConjunctiveSSE/pkg/utils"
	"context"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGenCiphertext(t *testing.T) {
	// Init ODXT
	dbName := "Crime_USENIX_REV"
	var odxt ODXT
	err := odxt.DBSetup(dbName, false)
	if err != nil {
		t.Fatal("DBSetup error", err)
	}
	// Init mongo
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	defer PlaintextDB.Client().Disconnect(ctx)

	collection := PlaintextDB.Collection("keyword_ids")

	start := time.Now()
	// from mongo
	// 读取第一个 document
	var result bson.M
	err = collection.FindOne(ctx, bson.D{}).Decode(&result)
	if err != nil {
		t.Fatal(err)
	}
	valSet, ok := result["val_set"].(primitive.A)
	if !ok {
		t.Fatal("val_set is not of type primitive.A")
	}
	ids := make([]string, 0, len(valSet))
	for _, v := range valSet {
		if str, ok := v.(string); ok {
			ids = append(ids, str)
		} else {
			t.Fatal("val_set contains non-string value")
		}
	}
	keyword := result["k"].(string)

	_, _ = odxt.Encrypt(keyword, ids, utils.Add)

	t.Log("Time cost:", time.Since(start))
}