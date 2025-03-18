package main

import (
	"ConjunctiveSSE/pkg/HDXT"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"
)

// Config 定义一个类型
type Config struct {
	Db               string `json:"db"`
	Phase            string `json:"phase"`
	Group            string `json:"group"`
	DelRate          int    `json:"del_rate"`
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
	
	if err := hdxt.Init(cfg.Db, false); err != nil {
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
		hdxt.SearchPhase(cfg.Db, cfg.Group)
		t2 := time.Since(t1)
		log.Println("SearchPhase time:", t2)
	}

	return nil
}