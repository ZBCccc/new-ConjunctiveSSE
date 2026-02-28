package main

import (
	"ConjunctiveSSE/pkg/ODXT"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Db       string `json:"db"`
	Phase    string `json:"phase"`
	Group    string `json:"group"`
	DelRate  int    `json:"del_rate"`
	MongoURI string `json:"mongo_uri"`
}

func main() {
	var config Config
	// Read config file
	file, err := os.Open("./cmd/ODXT/configs/config.json")
	if err != nil {
		fmt.Println("Error opening config file:", err)
		return
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		fmt.Println("Error decoding config file:", err)
		return
	}

	// Use parameters from config file
	fmt.Println("*********************************************")
	fmt.Println("Test_on: ", config.Db, "del_rate:", config.DelRate)
	fmt.Println("Start test_group:", config.Group, "phase:", config.Phase)
	fmt.Println("Start initial db...")

	// Run tests
	err = TestODXT(config)
	if err != nil {
		fmt.Println("TestODXT error:", err)
	}
}

func TestODXT(cfg Config) error {
	var odxt ODXT.ODXT
	err = odxt.DBSetup(cfg.Db, false, cfg.MongoURI)
	if err != nil {
		fmt.Println("DBSetup error", err)
		return err
	}

	if strings.Contains(cfg.Phase, "c") {
		t1 := time.Now()
		err = odxt.CiphertextGenPhase(cfg.Db)
		if err != nil {
			fmt.Println("CiphertextGenPhase error:", err)
			return err
		}
		t2 := time.Since(t1)
		fmt.Println("CiphertextGenPhase time:", t2)
	}
	if strings.Contains(cfg.Phase, "s") {
		t1 := time.Now()
		odxt.SearchPhase(cfg.Db, cfg.Group)
		t2 := time.Since(t1)
		fmt.Println("SearchPhase time:", t2)
	}

	return nil
}
