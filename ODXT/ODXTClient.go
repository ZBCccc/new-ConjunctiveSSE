package ODXT

import (
	"ConjunctiveSSE/util"
	"crypto/rand"
	"encoding/gob"
	"fmt"
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
}

func (client *Client) Setup() (err error) {
	// 生成4个32字节长度的随机私钥
	keyLen := 32

	for i := 0; i < 4; i++ {
		key := make([]byte, keyLen)
		if _, err := rand.Read(key); err != nil {
			return fmt.Errorf("error generating key %d: %v", i, err)
		}
		client.Keys[i] = key
	}

	// 初始化 UpdateCnt
	client.UpdateCnt = make(map[string]int)
	client.g = big.NewInt(65537)
	client.p, _ = new(big.Int).SetString("69445180235231407255137142482031499329548634082242122837872648805446522657159", 10)
	client.addr = "localhost:8080"

	return nil
}

func (client *Client) Update(id string, w string, op util.Operation) error {
	kt, kx, ky, kz := client.Keys[0], client.Keys[1], client.Keys[2], client.Keys[3]
	_, ok := client.UpdateCnt[w]
	if !ok {
		client.UpdateCnt[w] = 0
	}
	client.UpdateCnt[w] += 1

	// 计算HMAC-SHA256 PRF值
	m := []byte(w)
	m = append(m, big.NewInt(int64(client.UpdateCnt[w])).Bytes()...)

	address, err := util.PrfF(kt, append(m, big.NewInt(int64(0)).Bytes()...))
	if err != nil {
		fmt.Println(err)
		return err
	}

	val, err := util.PrfF(kx, append(m, big.NewInt(int64(1)).Bytes()...))
	if err != nil {
		fmt.Println(err)
		return err
	}
	for i := 0; i < 65; i++ {
		val[i] = val[i] ^ id[i]
	}
	if op == util.Add {
		val[65] = val[65] ^ 0xFF // 255 == xor 1
	} else {
		val[65] = val[65] ^ 0x00 // 0 == xor 0
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
	//p := new(big.Int).Sub(client.p, big.NewInt(1))
	alpha2 = new(big.Int).ModInverse(alpha2, client.p)
	alpha := new(big.Int).Mul(alpha1, alpha2)

	C, err := util.PrfFp(kx, []byte(w), client.p, client.g)
	if err != nil {
		fmt.Println(err)
		return err
	}
	A := new(big.Int).Mul(C, new(big.Int).SetBytes(idBytes))
	xtag := new(big.Int).Exp(client.g, A, client.p)

	// 准备需要发送的数据
	data := util.DataPacket{
		Address: address,
		Val:     val,
		Alpha:   alpha,
		Xtag:    xtag,
	}

	// 连接服务器
	conn, err := net.Dial("tcp", client.addr)
	if err != nil {
		fmt.Println("Error connecting:", err)
		return err
	}
	defer conn.Close()

	// 创建一个 gob 编码器
	encoder := gob.NewEncoder(conn)

	// 发送数据
	err = encoder.Encode(data)
	if err != nil {
		fmt.Println("Error sending data:", err)
		return err
	}

	fmt.Println("Data sent successfully")

	return nil
}

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
		m = append(m, big.NewInt(int64(j)).Bytes()...)

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
			xtoken2, _ := util.PrfFp(kz, []byte(w1), client.p, client.g)
			xtokenList[j][i] = new(big.Int).Exp(xtoken1, xtoken2, client.p)
			i++
		}

		// 打乱切片中的元素
		mrand.Shuffle(len(xtokenList[j]), func(i, j int) {
			xtokenList[j][i], xtokenList[j][j] = xtokenList[j][j], xtokenList[j][i]
		})
	}

	// 连接服务器
	conn, err := net.Dial("tcp", client.addr)
	if err != nil {
		fmt.Println("Error connecting:", err)
		return err
	}
	defer conn.Close()

	// 创建一个 gob 编码器
	encoder := gob.NewEncoder(conn)

	// 发送数据
	err = encoder.Encode(stokenList)
	if err != nil {
		fmt.Println("Error sending stokenList:", err)
		return err
	}

	err = encoder.Encode(xtokenList)
	if err != nil {
		fmt.Println("Error sending xtokenList:", err)
		return err
	}

	fmt.Println("Data sent successfully")

	return nil
}
