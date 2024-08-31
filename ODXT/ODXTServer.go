package ODXT

import (
	"ConjunctiveSSE/util"
	"github.com/bits-and-blooms/bloom/v3"
	"math/big"
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

// Setup 初始化 Server
func (server *Server) Setup() {
	// 初始化 Server 结构体
	server.EDB = EDB{
		TSet: make(map[string]Record),
		XSet: bloom.NewWithEstimates(1000000, 0.01), // 可以存储100万个元素,错误率为1%
	}
	server.p, _ = new(big.Int).SetString("69445180235231407255137142482031499329548634082242122837872648805446522657159", 10)
}

// Update 更新数据
func (server *Server) Update(data util.DataPacket) {
	// 将数据包中的数据存储到 EDB 中

	server.EDB.TSet[string(data.Address)] = Record{
		Value: data.Val,
		Alpha: data.Alpha,
	}
	server.EDB.XSet.Add(data.Xtag.Bytes())
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
