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
}

// Setup 初始化 Server
func (server *Server) Setup() {
	// 初始化 Server 结构体
	server.EDB = EDB{
		TSet: make(map[string]Record),
		XSet: bloom.NewWithEstimates(1000000, 0.01), // 可以存储100万个元素,错误率为1%
	}
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
func (server *Server) Search(stokenList [][]byte, xtokenList [][]*big.Int) error {

}
