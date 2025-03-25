package ODXT

import (
	"ConjunctiveSSE/pkg/utils"
	pbcUtil "ConjunctiveSSE/pkg/utils/pbc"
	"fmt"
	"math/big"
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
func (odxt *ODXT) Search(stokenList []string, xtokenList [][]*pbc.Element) ([]utils.SEOp) {
	sEOpList := make([]utils.SEOp, len(stokenList))
	// 搜索数据
	var wg sync.WaitGroup
	for j, stoken := range stokenList {
		wg.Add(1)
		go func(j int) {
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
	return sEOpList
}

// ClientSearchStep1 生成陷门
func (odxt *ODXT) ClientSearchStep1(w1 string, q []string) ([]string, [][]*pbc.Element) {
	// 读取密钥
	kt, kx, kz := odxt.Keys[0], odxt.Keys[1], odxt.Keys[3]
	counter := odxt.UpdateCnt[w1]

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
			msgLen := len(w1) + len(big.NewInt(int64(j+1)).Bytes())
			msg := make([]byte, 0, msgLen)

			msg = append(msg, []byte(w1)...)
			msg = append(msg, big.NewInt(int64(j+1)).Bytes()...)
			saddr, err := utils.PrfF(kt, append(msg, byte(0)))
			if err != nil {
				fmt.Println(err)
			}
			stokenList[j] = string(saddr)
			xtoken2, _ := pbcUtil.PrfToZr(kz, msg)
			for i := range qt {
				xtoken1 := xtoken1List[i]
				xtoken := pbcUtil.GToPower2(xtoken1, xtoken2)
				xtokenList[j][i] = xtoken
			}
		}(j)
	}
	wg.Wait()
	return stokenList, xtokenList
}

// ClientSearchStep2 解密
func (odxt *ODXT) ClientSearchStep2(w1 string, q []string, sEOpList []utils.SEOp) []string {
	kt := odxt.Keys[0]

	sIdList := make([]string, 0)
	for _, sEOp := range sEOpList {
		j, sval, cnt := sEOp.J, sEOp.Sval, sEOp.Cnt
		w1Andj := append(append([]byte(w1), big.NewInt(int64(j)).Bytes()...), big.NewInt(int64(1)).Bytes()...)
		tmp, _ := utils.PrfF(kt, w1Andj)
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

	return sIdList
}
