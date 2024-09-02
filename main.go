package main

import (
	"ConjunctiveSSE/ODXT"
	"ConjunctiveSSE/util"
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"time"
)

func init() {
	util.RegisterTypes()
}

func startServer(wg *sync.WaitGroup) {
	defer wg.Done()

	var server ODXT.Server
	err := server.Setup()
	if err != nil {
		log.Fatal(err)
	}
}

func startClient(wg *sync.WaitGroup) {
	defer wg.Done()

	var client ODXT.Client
	err := client.Setup()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Conn.Close()

	// 发送更新请求
	// start time
	start := time.Now()
	// 读取文件内容
	file, err := os.Open("sse_data_lite")
	if err != nil {
		log.Fatal("无法打开文件:", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// 读取关键词数量
	if !scanner.Scan() {
		log.Fatal("无法读取关键词数量")
	}
	keywordNumber, err := strconv.Atoi(scanner.Text())
	if err != nil {
		log.Fatal("无法解析关键词数量:", err)
	}

	// 循环读取关键词和文件
	for i := 0; i < keywordNumber; i++ {
		// 读取关键词
		if !scanner.Scan() {
			log.Fatal("无法读取关键词")
		}
		keyword := scanner.Text()

		// 读取文件数量
		if !scanner.Scan() {
			log.Fatal("无法读取文件数量")
		}
		num, err := strconv.Atoi(scanner.Text())
		fmt.Println("关键词", keyword, "的数量为：", num)
		if err != nil {
			log.Fatal("无法解析文件数量:", err)
		}

		// 读取文件
		for j := 0; j < num; j++ {
			if !scanner.Scan() {
				log.Fatal("无法读取文件")
			}
			file := scanner.Text()
			// 发送更新请求
			fmt.Printf("更新次数为 %d, 更新关键词 %s, 文件为 %s\n", j, keyword, file)
			err = client.Update(file, keyword, util.Add)
			if err != nil {
				log.Printf("更新关键词 %s 时出错: %v\n", keyword, err)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("读取文件时出错:", err)
	}

	// end time
	elapsed := time.Since(start)
	fmt.Printf("Elapsed time: %s\n", elapsed)

	// 示例：发送搜索请求
	// err = client.Search([]string{"w1", "w2"})
	// if err != nil {
	// 	log.Println("Error sending search request:", err)
	// 	log.Fatal(err)
	// }
}

func main() {
	var wg sync.WaitGroup
	wg.Add(2)

	go startServer(&wg)
	go startClient(&wg)

	wg.Wait()
}
