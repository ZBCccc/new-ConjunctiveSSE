package main

import (
	"ConjunctiveSSE/pkg/HDXT"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"
)

// Config defines the benchmark configuration.
type Config struct {
	Db       string `json:"db"`
	Phase    string `json:"phase"`
	Group    string `json:"group"`
	DelRate  int    `json:"del_rate"`
	MongoURI string `json:"mongo_uri"`
}

func main() {
	var config Config
	// 读取配置文件
	file, err := os.Open("./cmd/HDXT/configs/config.json")
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
	log.Println("*********************************************")
	log.Println("Test_on: ", config.Db, "del_rate:", config.DelRate)
	log.Println("Start test_group:", config.Group, "phase:", config.Phase)
	log.Println("Start initial db...")

	// Run tests
	err = TestHDXT(config)
	if err != nil {
		log.Println("TestHDXT error:", err)
	}
}

func TestHDXT(cfg Config) error {
	var hdxt HDXT.HDXT
	
	if err := hdxt.Init(cfg.Db, false, cfg.MongoURI); err != nil {
		log.Println("DBSetup error", err)
		return err
	}
	if strings.Contains(cfg.Phase, "c") {
		t1 := time.Now()
		hdxt.SetupPhase()
		t2 := time.Since(t1)
		log.Println("SetupPhase time:", t2)
	}
	if strings.Contains(cfg.Phase, "s") {
		t1 := time.Now()
		if err := hdxt.SearchPhase(cfg.Db, cfg.Group); err != nil {
			log.Println("SearchPhase error:", err)
			return err
		}
		t2 := time.Since(t1)
		log.Println("SearchPhase time:", t2)
	}

	return nil
}