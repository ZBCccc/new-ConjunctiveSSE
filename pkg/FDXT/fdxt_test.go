package FDXT

import (
	"ConjunctiveSSE/pkg/utils"
	"context"
	"math/big"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGenCiphertext(t *testing.T) {
	// Init fdxt
	dbName := "Crime_USENIX_REV"
	var fdxt FDXT
	fdxt.Setup(dbName)
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

	_, _ = fdxt.Encrypt(keyword, ids, Add)
	
	t.Log("Time cost:", time.Since(start))
}

func TestSearch(t *testing.T) {
	updt := 16644
	w := "F1438"
	srch := 1
	kt := []byte("0123456789123456")
	tklList := make([]*TKL, 0, updt)
	start := time.Now()
	for i := updt; i >= 1; i-- {
		// 预计算总长度
		msgLen := len(w) + len(big.NewInt(int64(srch)).Bytes()) + len(big.NewInt(int64(i)).Bytes()) + 1
		msg := make([]byte, 0, msgLen)

		// 一次性构建消息
		msg = append(msg, []byte(w)...)
		msg = append(msg, big.NewInt(int64(srch)).Bytes()...)
		msg = append(msg, big.NewInt(int64(i)).Bytes()...)

		l, _ := utils.PrfF(kt, append(msg, byte(0)))
		
		t, _ := utils.PrfF(kt, append(msg, byte(1)))
		
		tklList = append(tklList, &TKL{L: string(l), T: string(t)})
	}
	t.Log("Time cost:", time.Since(start))
}