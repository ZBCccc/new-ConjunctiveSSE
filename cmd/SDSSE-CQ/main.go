package main

import (
	"ConjunctiveSSE/pkg/SDSSE-CQ"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)


type Config struct {
	Db               string `json:"db"`
	Phase            string `json:"phase"`
	Group            string `json:"group"`
	DelRate          int    `json:"del_rate"`
}

func main() {
	var config Config
	file, err := os.Open("./cmd/SDSSE-CQ/configs/config.json")
	if err != nil {
		log.Fatal("Error opening config file:", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatal("Error decoding config file:", err)
	}

	// 使用配置文件中的参数
	fmt.Println("*********************************************")
	fmt.Println("Test_on: ", config.Db, "del_rate:", config.DelRate)
	fmt.Println("Start test_group:", config.Group, "phase:", config.Phase)
	fmt.Println("Start initial db...")

	// Run tests
	err = TestSDSSE_CQ(config)
	if err != nil {
		fmt.Println("TestSDSSE_CQ error:", err)
	}
}

func TestSDSSE_CQ(cfg Config) error {
	err := sdssecq.Init(cfg.Db)
	if err != nil {
		log.Fatal("Error initializing db:", err)
	}

	if strings.Contains(cfg.Phase, "c") {
		t1 := time.Now()
		err := sdssecq.CiphertextGenPhase(cfg.Db)
		if err != nil {
			fmt.Println("err:", err)
			return err
		}
		t2 := time.Since(t1)
		fmt.Println("UpdatePhase time:", t2)
	}
	if strings.Contains(cfg.Phase, "s") {
		t1 := time.Now()
		sdssecq.SearchPhase(cfg.Db, cfg.Group)
		t2 := time.Since(t1)
		fmt.Println("SearchPhase time:", t2)
	}
	return nil
}
