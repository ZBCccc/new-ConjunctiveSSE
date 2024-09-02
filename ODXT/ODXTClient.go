package ODXT

import (
	"ConjunctiveSSE/util"
	"bytes"
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"log"
	"math/big"
	mrand "math/rand"
	"net"
)

type Client struct {
	Keys      [4][]byte
	UpdateCnt map[string]int
	g         *big.Int
	p         *big.Int
	addr      string
	Conn      net.Conn
}

func (client *Client) Setup() error {
	// 生成4个32字节长度的随机私钥
	keyLen := 32

	for i := 0; i < 4; i++ {
		key := make([]byte, keyLen)
		if _, err := rand.Read(key); err != nil {
			log.Println("Error generating random key:", err)
			return err
		}
		client.Keys[i] = key
	}

	// 初始化 UpdateCnt
	client.UpdateCnt = make(map[string]int)
	client.g = big.NewInt(65537)
	client.p, _ = new(big.Int).SetString("69445180235231407255137142482031499329548634082242122837872648805446522657159", 10)
	client.addr = "localhost:8308"

	var err error
	client.Conn, err = net.Dial("tcp", client.addr)
	if err != nil {
		log.Println("Error connecting:", err)
		return err
	}

	return nil
}

// Update 向服务器发送更新请求
func (client *Client) Update(id string, w string, op util.Operation) error {
	kt, kx, ky, kz := client.Keys[0], client.Keys[1], client.Keys[2], client.Keys[3]
	_, ok := client.UpdateCnt[w]
	if !ok {
		client.UpdateCnt[w] = 0
	}
	client.UpdateCnt[w]++

	// 计算HMAC-SHA256 PRF值
	m := append([]byte(w), big.NewInt(int64(client.UpdateCnt[w])).Bytes()...)
	address, err := util.PrfF(kt, append(m, big.NewInt(int64(0)).Bytes()...))
	if err != nil {
		fmt.Println(err)
		return err
	}

	val, err := util.PrfF(kt, append(m, big.NewInt(int64(1)).Bytes()...))
	if err != nil {
		fmt.Println(err)
		return err
	}
	val, err = util.BytesXORWithOp(val, []byte(id), int(op))
	if err != nil {
		fmt.Println(err)
		return err
	}

	idBytes := []byte(id)
	idBytes = append(idBytes, byte(op))
	alpha1, err := util.PrfFp(ky, idBytes, client.p, client.g)
	if err != nil {
		fmt.Println(err)
		return err
	}

	alpha2, err := util.PrfFp(kz, m, client.p, client.g)
	if err != nil {
		fmt.Println(err)
		return err
	}
	p := new(big.Int).Sub(client.p, big.NewInt(1))
	alpha2 = new(big.Int).ModInverse(alpha2, p)
	alpha := new(big.Int).Mul(alpha1, alpha2)

	C, err := util.PrfFp(kx, []byte(w), client.p, client.g)
	if err != nil {
		fmt.Println(err)
		return err
	}
	A := new(big.Int).Mul(C, alpha1)
	xtag := new(big.Int).Exp(client.g, A, client.p)

	// 构造更新请求
	req := util.Request{
		Type: util.Update,
		Payload: util.UpdatePayload{
			Address: address,
			Val:     val,
			Alpha:   alpha,
			Xtag:    xtag,
		},
	}

	// 创建一个 gob 编码器
	encoder := gob.NewEncoder(client.Conn)

	// 发送数据
	err = encoder.Encode(req)
	if err != nil {
		fmt.Println("Error sending data:", err)
		return err
	}

	return nil
}

// Search 向服务器发送搜索请求并处理响应y
func (client *Client) Search(q []string) error {
	// 读取密钥
	kt, kx, kz := client.Keys[0], client.Keys[1], client.Keys[3]
	counter, w1, st := 1000000, q[0], client.UpdateCnt

	// 选择查询频率最低的关键字
	for _, w := range q {
		num := st[w]
		if num < counter {
			w1 = w
			counter = num
		}
	}

	// 初始化stokenList和xtokenList
	stokenList := make([][]byte, counter)
	xtokenList := make([][]*big.Int, counter)
	for i := range xtokenList {
		xtokenList[i] = make([]*big.Int, len(q)-1)
	}
	for j := 0; j < counter; j++ {
		m := []byte(w1)
		m = append(m, big.NewInt(int64(j+1)).Bytes()...)

		saddr, err := util.PrfF(kt, append(m, big.NewInt(int64(0)).Bytes()...))
		if err != nil {
			fmt.Println(err)
			return err
		}
		stokenList[j] = saddr

		i := 0
		for _, wi := range q {
			if wi == w1 {
				continue
			}
			xtoken1, _ := util.PrfFp(kx, []byte(wi), client.p, client.g)
			xtoken2, _ := util.PrfFp(kz, append([]byte(w1), big.NewInt(int64(j+1)).Bytes()...), client.p, client.g)
			xtokenList[j][i] = new(big.Int).Exp(client.g, new(big.Int).Mul(xtoken1, xtoken2), client.p)
			i++
		}

		// 打乱切片中的元素
		mrand.Shuffle(len(xtokenList[j]), func(i, j int) {
			xtokenList[j][i], xtokenList[j][j] = xtokenList[j][j], xtokenList[j][i]
		})
	}

	// 构造搜索请求
	req := util.Request{
		Type: util.Search,
		Payload: util.SearchPayload{
			StokenList: stokenList,
			XtokenList: xtokenList,
		},
	}

	// 创建一个 gob 编码器
	encoder := gob.NewEncoder(client.Conn)

	// 发送数据
	err := encoder.Encode(req)
	if err != nil {
		fmt.Println("Error sending data:", err)
		return err
	}

	// 接收数据
	decoder := gob.NewDecoder(client.Conn)
	var resp util.Response
	err = decoder.Decode(&resp)
	if err != nil {
		fmt.Println("Error receiving data:", err)
		return err
	}
	sEOpList := resp.SEOpList

	// 本地检查
	sIdList := make([][]byte, 0)
	for _, sEOp := range sEOpList {
		j, sval, cnt := sEOp.J, sEOp.Sval, sEOp.Cnt
		w1Andj := append(append([]byte(w1), big.NewInt(int64(j)).Bytes()...), big.NewInt(int64(1)).Bytes()...)
		tmp, err := util.PrfF(kt, w1Andj)
		if err != nil {
			fmt.Println(err)
			return err
		}
		id := make([]byte, 31)
		for i := 0; i < 31; i++ {
			id[i] = tmp[i] ^ sval[i]
		}
		var op util.Operation = util.Operation(tmp[31] ^ sval[31])
		if op == util.Add && cnt == len(q) {
			sIdList = append(sIdList, id)
		} else if op == util.Del && cnt > 0 {
			sIdList = removeElement(sIdList, id)
		}
	}

	return nil
}

// 删除sIdList中的特定元素
func removeElement(slice [][]byte, target []byte) [][]byte {
	for i, v := range slice {
		if bytes.Equal(v, target) {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
