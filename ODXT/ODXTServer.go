package ODXT

import (
	"ConjunctiveSSE/util"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"sync"

	"github.com/bits-and-blooms/bloom/v3"
)

type EDB struct {
	TSet map[string]Record
	XSet *bloom.BloomFilter
}

type Record struct {
	Value []byte
	Alpha *big.Int
}

type Server struct {
	EDB EDB
	p   *big.Int
}

func init() {
	var registerOnce sync.Once
	registerOnce.Do(func() {
		gob.Register(util.UpdatePayload{})
		gob.Register(util.SearchPayload{})
		gob.Register(util.Response{})
	})

}

// Setup 初始化 Server 并启动服务器
func (server *Server) Setup() {

	// 初始化 Server 结构体
	server.EDB = EDB{
		TSet: make(map[string]Record),
		XSet: bloom.NewWithEstimates(1000000, 0.01), // 可以存储100万个元素,错误率为1%
	}
	server.p, _ = new(big.Int).SetString("69445180235231407255137142482031499329548634082242122837872648805446522657159", 10)

	// 启动服务器
	ln, err := net.Listen("tcp", ":8308")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	log.Println("Server is listening on port 8308...")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go server.handleConnection(conn)
	}
}

// Update 更新数据
func (server *Server) Update(data util.UpdatePayload) {
	// 将数据包中的数据存储到 EDB 中

	server.EDB.TSet[string(data.Address)] = Record{
		Value: data.Val,
		Alpha: data.Alpha,
	}
	server.EDB.XSet.Add(data.Xtag.Bytes())
	fmt.Println("Data updated successfully")
}

// Search 搜索数据
func (server *Server) Search(stokenList [][]byte, xtokenList [][]*big.Int) []util.SEOp {
	// 搜索数据
	sEOpList := make([]util.SEOp, len(stokenList))
	// 遍历 stokenList
	cnt := 1
	for j, stoken := range stokenList {
		cnt = 1
		// 获取 Record
		record := server.EDB.TSet[string(stoken)]

		// 遍历 xtokenList
		for _, xtoken := range xtokenList[j] {
			// 判断 xtag 是否匹配
			xtag := new(big.Int).Exp(xtoken, record.Alpha, server.p)
			if server.EDB.XSet.Test(xtag.Bytes()) {
				cnt++
			}
		}
		sEOpList[j] = util.SEOp{
			J:    j,
			Sval: record.Value,
			Cnt:  cnt,
		}
	}

	return sEOpList
}

// handleConnection 处理客户端连接
func (server *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	decoder := gob.NewDecoder(conn)
	encoder := gob.NewEncoder(conn)

	for {
		var req util.Request
		err := decoder.Decode(&req)
		if err != nil {
			if err == io.EOF {
				log.Println("Client disconnected")
			} else {
				log.Println("Error decoding request:", err)
			}
			return
		}

		switch req.Type {
		case util.Update:
			payload := req.Payload.(util.UpdatePayload)
			server.Update(payload)
		case util.Search:
			payload := req.Payload.(util.SearchPayload)
			fmt.Println("Received search request")
			fmt.Println(payload)
			resp := server.Search(payload.StokenList, payload.XtokenList)
			fmt.Println(resp)
			err := encoder.Encode(util.Response{SEOpList: resp})
			if err != nil {
				log.Println("Error encoding response:", err)
			}
		}
	}
}
