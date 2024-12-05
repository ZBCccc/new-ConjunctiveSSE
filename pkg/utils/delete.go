package utils

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
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

func GenDeletePairs(filePath string, num int) []deletePair {
	// 1.读取.txt文件，文件的格式为string:int，其中key是文件id，value是该文件id所包含的关键词的数量
	idCounts, err := texFileRead(filePath)
	if err != nil {
		log.Fatal(err)
	}

	// 2.边累加边记录id，直到累加和达到num
	var deleteIDs []string
	sum := 0
	for _, val := range idCounts {
		sum += val.Count
		deleteIDs = append(deleteIDs, val.Id)
		if sum >= num {
			break
		}
	}

	// 3.得到id，从mongodb数据库中读取id和对应的keywords，保存在deletePair中
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			log.Fatal(err)
		}
	}()

	db := client.Database("Wiki_USENIX")
	collection := db.Collection("id_keywords")

	deletePairs := make([]deletePair, len(deleteIDs))
	for i := 0; i < len(deleteIDs); i++ {
		fmt.Println("deleteID:", deleteIDs[i])
		cur, err := collection.Find(context.TODO(), bson.D{{"id", deleteIDs[i]}})
		if err != nil {
			log.Fatal(err)
		}
		defer cur.Close(context.TODO())

		var result []bson.M
		if err = cur.All(context.TODO(), &result); err != nil {
			log.Fatal(err)
		}
		for _, v := range result {
			// 假设 val_set 是一个数组（切片）
			if valSet, ok := v["val_set"].([]interface{}); ok {
				// 遍历 val_set 数组，提取其中的字符串元素
				deletePairs[i].Keywords = make([]string, 0, len(valSet))
				for _, item := range valSet {
					if keyword, ok := item.(string); ok {
						deletePairs[i].Keywords = append(deletePairs[i].Keywords, keyword)
					}
				}
			}
		}

	}

	// 4.返回deletePair

	return deletePairs
}
