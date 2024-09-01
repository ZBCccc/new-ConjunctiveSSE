package main

import (
	"ConjunctiveSSE/ODXT"
	"ConjunctiveSSE/util"
	"fmt"
	"log"
	"sync"
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

	// 示例：发送更新请求
	err = client.Update("example_id", "example_value", util.Add)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println("Update request sent successfully")

	err = client.Update("example_id", "w2", util.Add)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Println("第二次更新成功")

	// 示例：发送搜索请求
	err = client.Search([]string{"example_value", "w2"})
	if err != nil {
		log.Println("Error sending search request:", err)
		log.Fatal(err)
	}
	fmt.Println("Search request sent successfully")
}

func main() {
	var wg sync.WaitGroup
	wg.Add(2)

	go startServer(&wg)
	go startClient(&wg)

	wg.Wait()
}
