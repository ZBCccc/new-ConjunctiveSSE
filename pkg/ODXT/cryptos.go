package ODXT

import (
	"ConjunctiveSSE/pkg/utils"
	pbcUtil "ConjunctiveSSE/pkg/utils/pbc"
	"fmt"
	"math/big"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/Nik-U/pbc"
)

func (odxt *ODXT) Encrypt(keyword string, ids []string, operation utils.Operation) (time.Duration, error) {
	var encryptedTime time.Duration
	if _, ok := odxt.UpdateCnt[keyword]; !ok {
		odxt.UpdateCnt[keyword] = 0
	}
	for _, id := range ids {
		// 在id后添加一个随机数生成新的id,避免重复,例如：假如id为123，则新的id为123@123
		id = id + "@" + strconv.Itoa(rand.Intn(1000000))
		encryptedTimeAOperation, err := odxt.encrypt(keyword, id, operation)
		if err != nil {
			return encryptedTime, err
		}
		encryptedTime += encryptedTimeAOperation
	}

	return encryptedTime, nil
}

func (odxt *ODXT) encrypt(keyword string, id string, operation utils.Operation) (time.Duration, error) {
	kt, kx, ky, kz := odxt.Keys[0], odxt.Keys[1], odxt.Keys[2], odxt.Keys[3]
	odxt.UpdateCnt[keyword]++
	msgLen := len(keyword) + len(big.NewInt(int64(odxt.UpdateCnt[keyword])).Bytes())
	wWc := make([]byte, 0, msgLen)
	wWc = append(wWc, []byte(keyword)...)
	wWc = append(wWc, big.NewInt(int64(odxt.UpdateCnt[keyword])).Bytes()...)

	var encryptedTime time.Duration
	start := time.Now()

	// address = PRF(kt, w||wc||0)
	address, _ := utils.PrfF(kt, append(wWc, byte(0)))

	// val = PRF(kt, w||wc||1) xor (id||op)
	val, _ := utils.PrfF(kt, append(wWc, byte(1)))
	val, _ = utils.BytesXORWithOp(val, []byte(id), int(operation))

	// alpha = Fp(ky, id||op) * Fp(kz, w||wc)^-1
	alpha, alpha1, _ := utils.ComputeAlpha(ky, kz, []byte(id), int(operation), wWc)

	// xtag = g^{Fp(Kx, w)*Fp(Ky, id||op)} mod p-1
	xtag1, _ := pbcUtil.PrfToZr(kx, []byte(keyword))

	Xtag := pbcUtil.GToPower2(xtag1, alpha1)

	encryptedTime += time.Since(start)

	// Encoded the ciphertext
	odxt.XSet[string(Xtag.Bytes())] = 1
	odxt.TSet[string(address)] = &TsetValue{string(val), alpha}

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
	var wg sync.WaitGroup
	for j, stoken := range stokenList {
		wg.Add(1)
		go func (j int)  {
			defer wg.Done()
			cnt := 1
			val, alpha := odxt.TSet[stoken].Val, odxt.TSet[stoken].Alpha
			// 遍历 xtokenList
			for _, xtoken := range xtokenList[j] {
				// 判断 xtag 是否匹配
				xtag := pbcUtil.Pow(xtoken, alpha)
				if _, ok := odxt.XSet[string(xtag.Bytes())]; ok {
					cnt++
				}
			}
			sEOpList[j] = utils.SEOp{
				J:    j + 1,
				Sval: val,
				Cnt:  cnt,
			}
		}(j)
	}
	wg.Wait()
	serverTime := time.Since(start)
	return trapdoorTime, serverTime, sEOpList
}

// ClientSearchStep1 生成陷门
func (odxt *ODXT) ClientSearchStep1(q []string) ([]string, [][]*pbc.Element) {
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
	xtokenList := make([][]*pbc.Element, counter)
	for i := range xtokenList {
		xtokenList[i] = make([]*pbc.Element, len(q)-1)
	}
	qt := utils.RemoveElement(q, w1)
	xtoken1List := make([]*pbc.Element, len(qt))
	
	for i, wi := range qt {
		xtoken1, _ := pbcUtil.PrfToZr(kx, []byte(wi))
		xtoken1List[i] = xtoken1
	}
	var wg sync.WaitGroup
	for j := 0; j < counter; j++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			saddr, err := utils.PrfF(kt, append(append([]byte(w1), big.NewInt(int64(j+1)).Bytes()...), byte(0)))
			if err != nil {
				fmt.Println(err)
			}
			stokenList[j] = string(saddr)
			xtoken2, _ := pbcUtil.PrfToZr(kz, append([]byte(w1), big.NewInt(int64(j+1)).Bytes()...))
			var wgg sync.WaitGroup
			for i := range qt {
				wgg.Add(1)
				go func(i, j int) {
					defer wgg.Done()
					xtoken1 := xtoken1List[i]
					xtoken := pbcUtil.GToPower2(xtoken1, xtoken2)
					xtokenList[j][i] = xtoken
				}(i, j)
			}
			wgg.Wait()
		}(j)
	}
	wg.Wait()
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
		val := []byte(sval)
		for i := range 31 {
			id[i] = tmp[i] ^ val[i]
		}
		var op = utils.Operation(tmp[31] ^ val[31])
		if op == utils.Add && cnt == len(q) {
			sIdList = append(sIdList, string(id))
		} else if op == utils.Del && cnt > 0 {
			sIdList = utils.RemoveElement(sIdList, string(id))
		}
	}

	return sIdList, nil
}
