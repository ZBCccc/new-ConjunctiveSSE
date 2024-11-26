package ODXT

import (
	"ConjunctiveSSE/pkg/utils"
	"encoding/base64"
	"fmt"
	"log"
	"math/big"
	"time"
)

func (odxt *ODXT) Encrypt(keyword string, ids []string, operation utils.Operation) (time.Duration, error) {
	kt, kx, ky, kz := odxt.Keys[0], odxt.Keys[1], odxt.Keys[2], odxt.Keys[3]
	var encryptedTime time.Duration
	if _, ok := odxt.UpdateCnt[keyword]; !ok {
		odxt.UpdateCnt[keyword] = 0
	}
	for _, id := range ids {
		start := time.Now()
		odxt.UpdateCnt[keyword]++
		msgLen := len(keyword) + len(big.NewInt(int64(odxt.UpdateCnt[keyword])).Bytes())
		wWc := make([]byte, 0, msgLen)
		wWc = append(wWc, []byte(keyword)...)
		wWc = append(wWc, big.NewInt(int64(odxt.UpdateCnt[keyword])).Bytes()...)

		// address = PRF(kt, w||wc||0)
		address, err := utils.PrfF(kt, append(wWc, byte(0)))
		if err != nil {
			return encryptedTime, err
		}

		// val = PRF(kt, w||wc||1) xor (id||op)
		val, err := utils.PrfF(kt, append(wWc, byte(1)))
		if err != nil {
			return encryptedTime, err
		}
		val, err = utils.BytesXORWithOp(val, []byte(id), int(operation))
		if err != nil {
			return encryptedTime, err
		}

		// alpha = Fp(ky, id||op) * Fp(kz, w||wc)^-1
		alpha, alpha1, err := utils.ComputeAlpha(ky, kz, []byte(id), int(operation), wWc, p, g)
		if err != nil {
			return encryptedTime, err
		}

		// xtag = g^{Fp(Kx, w)*Fp(Ky, id||op)} mod p
		C, err := utils.PrfFp(kx, []byte(keyword), p, g)
		if err != nil {
			return encryptedTime, err
		}
		A := new(big.Int).Mul(C, alpha1)
		xtag := new(big.Int).Exp(g, A, p)

		encryptedTime += time.Since(start)

		// Encoded the ciphertext
		odxt.XSet[base64.StdEncoding.EncodeToString(xtag.Bytes())] = 1
		odxt.TSet[base64.StdEncoding.EncodeToString(address)] = &tsetValue{base64.StdEncoding.EncodeToString(val), alpha}
	}

	return encryptedTime, nil
}

// Search 搜索，生成search token
func (odxt *ODXT) Search(q []string) (time.Duration, time.Duration, []utils.SEOp) {
	start := time.Now()
	// 生成陷门
	stokenList, xtokenList := odxt.ClientSearchStep1(q)
	trapdoorTime := time.Since(start)

	sEOpList := make([]utils.SEOp, len(stokenList))
	start = time.Now()
	// 搜索数据
	for j, stoken := range stokenList {
		cnt := 1
		val, alpha := odxt.TSet[stoken].val, odxt.TSet[stoken].alpha
		// 遍历 xtokenList
		for _, xtoken := range xtokenList[j] {
			// 判断 xtag 是否匹配
			xtag := new(big.Int).Exp(xtoken, alpha, p)
			if _, ok := odxt.XSet[base64.StdEncoding.EncodeToString(xtag.Bytes())]; ok {
				cnt++
			}
		}
		sEOpList[j] = utils.SEOp{
			J:    j + 1,
			Sval: val,
			Cnt:  cnt,
		}
	}
	serverTime := time.Since(start)
	return trapdoorTime, serverTime, sEOpList
}

// ClientSearchStep1 生成陷门
func (odxt *ODXT) ClientSearchStep1(q []string) ([]string, [][]*big.Int) {
	// 读取密钥
	kt, kx, kz := odxt.Keys[0], odxt.Keys[1], odxt.Keys[3]
	counter, w1, st := 1000000, q[0], odxt.UpdateCnt

	// 选择查询频率最低的关键字
	for _, w := range q {
		num := st[w]
		if num < counter {
			w1 = w
			counter = num
		}
	}

	// 初始化stokenList和xtokenList
	stokenList := make([]string, counter)
	xtokenList := make([][]*big.Int, counter)
	for i := range xtokenList {
		xtokenList[i] = make([]*big.Int, len(q)-1)
	}
	qt := utils.RemoveElement(q, w1)
	for j := 0; j < counter; j++ {
		saddr, err := utils.PrfF(kt, append(append([]byte(w1), big.NewInt(int64(j+1)).Bytes()...), byte(0)))
		if err != nil {
			fmt.Println(err)
		}
		stokenList[j] = base64.StdEncoding.EncodeToString(saddr)

		for i, wi := range qt {
			xtoken1, _ := utils.PrfFp(kx, []byte(wi), p, g)
			xtoken2, _ := utils.PrfFp(kz, append([]byte(w1), big.NewInt(int64(j+1)).Bytes()...), p, g)
			xtoken := new(big.Int).Exp(g, new(big.Int).Mul(xtoken1, xtoken2), p)
			xtokenList[j][i] = xtoken
		}
	}
	return stokenList, xtokenList
}

// Decrypt 解密
func (odxt *ODXT) Decrypt(q []string, sEOpList []utils.SEOp) ([]string, error) {
	kt := odxt.Keys[0]
	counter, w1, st := 1000000, q[0], odxt.UpdateCnt

	// 选择查询频率最低的关键字
	for _, w := range q {
		num := st[w]
		if num < counter {
			w1 = w
			counter = num
		}
	}

	sIdList := make([]string, 0)
	for _, sEOp := range sEOpList {
		j, sval, cnt := sEOp.J, sEOp.Sval, sEOp.Cnt
		w1Andj := append(append([]byte(w1), big.NewInt(int64(j)).Bytes()...), big.NewInt(int64(1)).Bytes()...)
		tmp, err := utils.PrfF(kt, w1Andj)
		if err != nil {
			fmt.Println(err)
			return nil, err
		}
		id := make([]byte, 31)
		val, err := base64.StdEncoding.DecodeString(sval)
		if err != nil {
			log.Fatal(err)
			return nil, err
		}
		for i := 0; i < 31; i++ {
			id[i] = tmp[i] ^ val[i]
		}
		var op = utils.Operation(tmp[31] ^ val[31])
		if op == utils.Add && cnt == len(q) {
			sIdList = append(sIdList, base64.StdEncoding.EncodeToString(id))
		} else if op == utils.Del && cnt > 0 {
			sIdList = utils.RemoveElement(sIdList, base64.StdEncoding.EncodeToString(id))
		}
	}

	return sIdList, nil
}
