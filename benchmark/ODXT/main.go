package main

import (
	"ConjunctiveSSE/ODXT"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

// 定义一个类型
type Config struct {
	Db      string `json:"db"`
	Phase   string `json:"phase"`
	Group   string `json:"group"`
	DelRate    int    `json:"del_rate"`
	DBSetupFromFiles bool `json:"db_setup_from_files"`
	XSetPath   string `json:"xset_path"`
	UpdateCntPath string `json:"update_cnt_path"`
}

func main() {
	var config Config
	// 读取配置文件
	file, err := os.Open("./benchmark/ODXT/config.json")
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

	// 使用配置文件中的参数
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
	if cfg.DBSetupFromFiles {
		err := odxt.DBSetupFromFiles(cfg.Db, cfg.XSetPath, cfg.UpdateCntPath)
		if err != nil {
			fmt.Println("DBSetup error", err)
			return err
		}
	} else {
		err := odxt.DBSetup(cfg.Db, false)
		if err != nil {
			fmt.Println("DBSetup error", err)
			return err
		}
	}
	if strings.Contains(cfg.Phase, "c") {
		t1 := time.Now()
		odxt.CiphertextGenPhase(cfg.Db)
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
