package ODXT

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

func (server *Server) Setup() {
	// 初始化 Server 结构体
	server.EDB = EDB{
		TSet: make(map[string]Record),
		XSet: make(map[string]int),
	}
}
