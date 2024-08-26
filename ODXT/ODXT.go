package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"
)

type operation int

const (
	del operation = iota // 0
	add                  // 1
)

type EDB struct {
	TSet map[string]Record
	XSet map[string]int
}

type Record struct {
	Value string
	Alpha string
}

type Server struct {
	EDB EDB
}

type Client struct {
	Keys      [4][]byte
	UpdateCnt map[string]int
}

func main() {
	var client Client
	var server Server

	// ODXT.Setup
	err := Setup(&client, &server)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(client)
	fmt.Println(server)

}

func Setup(client *Client, server *Server) (err error) {
	// 生成4个32字节长度的随机私钥
	keyLen := 32
	numKeys := 4

	for i := 0; i < numKeys; i++ {
		key := make([]byte, keyLen)
		_, err := rand.Read(key)
		if err != nil {
			fmt.Printf("Error generating key %d: %v\n", i, err)
			return err
		}
		client.Keys[i] = key
	}

	// 初始化 UpdateCnt
	client.UpdateCnt = make(map[string]int)

	// 初始化 EDB 结构体
	edb := EDB{
		TSet: make(map[string]Record),
		XSet: make(map[string]int),
	}

	// 初始化 Server 结构体
	server.EDB = edb

	return nil
}

func Update(client *Client, id string, w string, op operation) {
	kt, kx, ky, kz := client.Keys[0], client.Keys[1], client.Keys[2], client.Keys[3]
	_, ok := client.UpdateCnt[w]
	if !ok {
		client.UpdateCnt[w] = 0
	}
	client.UpdateCnt[w] += 1

	// 计算HMAC-SHA256 PRF值
	mac := hmac.New(sha256.New, kt)
	mac.Write([]byte(w))
	mac.Write(big.NewInt(int64(client.UpdateCnt[w])).Bytes())
	mac.Write(big.NewInt(int64(0)).Bytes())
	address := mac.Sum(nil)

	mac = hmac.New(sha256.New, kt)
	mac.Write([]byte(w))
	mac.Write(big.NewInt(int64(client.UpdateCnt[w])).Bytes())
	mac.Write(big.NewInt(int64(1)).Bytes())
	val := mac.Sum(nil)
	for i := 0; i < 65; i++ {
		val[i] = val[i] ^ id[i]
	}
	if op == add {
		val[65] = val[65] ^ 0xFF
	} else {
		val[65] = val[65] ^ 0x00
	}

	mac = hmac.New(sha256.New, ky)
	mac.Write([]byte(id))
	mac.Write(big.NewInt(int64(op)).Bytes())
	alpha := mac.Sum(nil)

}
