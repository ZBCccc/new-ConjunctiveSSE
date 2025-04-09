package main

import (
	"ConjunctiveSSE/pkg/utils"
	pbcUtil "ConjunctiveSSE/pkg/utils/pbc"
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type Xtag struct {
	ID   int64 // 通常自增主键是 int64
	Label string
	CipherText string 
}

type Tset struct {
	ID   int64 // 通常自增主键是 int64
	Address string
	Value string
	Alpha string
}

func main() {
	// 1. 数据库连接信息 (DSN)
	// 格式: "user:password@tcp(host:port)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	dsn := "root:123456@tcp(127.0.0.1:3306)/update_test?charset=utf8mb4&parseTime=True&loc=Local"

	// 2. 打开数据库连接
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("准备数据库连接失败: %v", err)
	}
	defer db.Close()

	// 3. 检查数据库连接是否真的可用 (可选但推荐)
	err = db.Ping()
	if err != nil {
		log.Fatalf("无法连接到数据库: %v", err)
	}
	log.Println("数据库连接成功!")

	// 5. 准备 SQL INSERT 语句 (使用 ? 作为占位符)
	// 列名要与你的表结构匹配
	insertXtagSQL := "INSERT INTO xtag(label, cipher_text) VALUES (?, ?)"
	insertTsetSQL := "INSERT INTO tset(address, value, alpha) VALUES (?, ?, ?)"
	// 6. 创建预处理语句
	stmtXtag, err := db.Prepare(insertXtagSQL)
	if err != nil {
		log.Printf("准备 SQL 语句失败: %v\n", err)
		return // 或者进行其他错误处理
	}
	// 使用 defer 确保在不需要时关闭语句
	defer stmtXtag.Close()

	stmtTset, err := db.Prepare(insertTsetSQL)
	if err != nil {
		log.Printf("准备 SQL 语句失败: %v\n", err)
		return // 或者进行其他错误处理
	}
	// 使用 defer 确保在不需要时关闭语句
	defer stmtTset.Close()

	start := time.Now()

	keyword := "F1438"
	op := 1
	Max := 12
	srch := 1
	updt := 12
	kw := []byte("0123456789123456")
	kt := []byte("0123456789123456")
	kx := []byte("0123456789123456")
	ky := []byte("0123456789123456")
	kz := []byte("0123456789123456")
	for i := range 100 {
		msgLen := len(keyword) + len(big.NewInt(int64(Max)).Bytes()) + 1
		msg := make([]byte, 0, msgLen)
		msg = append(msg, []byte(keyword)...)
		msg = append(msg, big.NewInt(int64(Max)).Bytes()...)
		id := strconv.Itoa(i)
		
		addr, _ := utils.PrfF(kw, append(msg, byte(0)))
		val, _ := utils.PrfF(kw, append(msg, byte(1)))
		
		val, _ = utils.BytesXORWithOp(val, []byte(id), int(op))
		
		msgLen = len(keyword) + len(big.NewInt(int64(srch)).Bytes()) + len(big.NewInt(int64(updt)).Bytes()) + 1
		msg = make([]byte, 0, msgLen)
		msg = append(msg, []byte(keyword)...)
		msg = append(msg, big.NewInt(int64(srch)).Bytes()...)
		msg = append(msg, big.NewInt(int64(updt)).Bytes()...)
		l, _ := utils.PrfF(kt, append(msg, byte(0)))
		tt, _ := utils.PrfF(kt, append(msg, byte(1)))
		xtag1, _ := pbcUtil.PrfToZr(kx, []byte(keyword))
		xtag2, _ := pbcUtil.PrfToZr(ky, append([]byte(id), byte(op)))
		xtag := pbcUtil.GToPower2(xtag1, xtag2)
		c := utils.BytesXOR(xtag.Bytes(), tt)
		alpha, _, _ := utils.ComputeAlpha(ky, kz, []byte(id), int(op), append([]byte(keyword), big.NewInt(int64(Max)).Bytes()...))
		
		_, err = stmtXtag.Exec(string(l), string(c))
		if err != nil {
			log.Printf("执行插入语句失败: %v\n", err)
			return
		}
		_, err = stmtTset.Exec(string(addr), string(val), string(alpha.Bytes()))
		if err != nil {
			log.Printf("执行插入语句失败: %v\n", err)
			return
		}
	}
	

	elapsed := time.Since(start)
	fmt.Printf("执行插入语句完成，用时: %s\n", elapsed)
}
