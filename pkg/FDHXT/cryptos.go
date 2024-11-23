package FDXT

import (
	"ConjunctiveSSE/pkg/utils"
	"encoding/base64"
	"fmt"
	"math"
	"math/big"
)

type Operation int

const (
	Add Operation = iota
	Delete
)

type TKL struct {
	l, t string
}

func (fdxt *FDXT) ClientSearchStep1(q []string) ([]*TKL, []string, [][]*big.Int, error) {
	kw, kt, kx, kz := fdxt.Keys[0], fdxt.Keys[1], fdxt.Keys[2], fdxt.Keys[4]
	counter, w1 := math.MaxInt64, q[0]
	for _, w := range q {
		num := fdxt.Count[w].max
		if num < counter {
			w1 = w
			counter = num
		}
	}
	tklList := make([]*TKL, 0, 1000)
	for _, w := range q {
		if _, ok := fdxt.Count[w]; !ok {
			return nil, nil, nil, fmt.Errorf("keyword %s not found", w)
		}
		srch, updt := fdxt.Count[w].srch, fdxt.Count[w].updt
		for i := updt; i >= 1; i-- {
			// 预计算总长度
			msgLen := len(w) + len(big.NewInt(int64(srch)).Bytes()) + len(big.NewInt(int64(i)).Bytes()) + 1
			msg := make([]byte, 0, msgLen)

			// 一次性构建消息
			msg = append(msg, []byte(w)...)
			msg = append(msg, big.NewInt(int64(srch)).Bytes()...)
			msg = append(msg, big.NewInt(int64(i)).Bytes()...)

			l, err := utils.PrfF(kt, append(msg, byte(0)))
			if err != nil {
				return nil, nil, nil, err
			}
			t, err := utils.PrfF(kt, append(msg, byte(1)))
			if err != nil {
				return nil, nil, nil, err
			}
			tklList = append(tklList, &TKL{l: base64.StdEncoding.EncodeToString(l), t: base64.StdEncoding.EncodeToString(t)})
		}
	}
	STKL := make([]string, 0, 1000)
	xtkList := make([][]*big.Int, fdxt.Count[w1].max+1)
	// fmt.Printf("Initial w1: %s, max: %d\n", w1, fdxt.Count[w1].max)
	qt := utils.RemoveElement(q, w1)
	for j := 1; j <= fdxt.Count[w1].max; j++ {
		msg := make([]byte, 0, len(w1)+len(big.NewInt(int64(j)).Bytes())+1)
		msg = append(msg, []byte(w1)...)
		msg = append(msg, big.NewInt(int64(j)).Bytes()...)
		addr, err := utils.PrfF(kw, append(msg, byte(0)))
		if err != nil {
			return nil, nil, nil, err
		}
		STKL = append(STKL, base64.StdEncoding.EncodeToString(addr))
		// fmt.Printf("j=%d, len(q)=%d, len(qt)=%d, qt=%v\n", j, len(q), len(qt), qt)
		xtkList[j] = make([]*big.Int, 0, len(qt))
		// fmt.Printf("Before loop - j=%d, xtkList[j] initialized with length: %d\n", j, len(xtkList[j]))
		for _, w := range qt {
			// fmt.Printf("Processing w=%s for j=%d\n", w, j)
			xtk1, err := utils.PrfFp(kx, []byte(w), p, g)
			if err != nil {
				return nil, nil, nil, err
			}
			xtk2, err := utils.PrfFp(kz, msg, p, g)
			if err != nil {
				return nil, nil, nil, err
			}
			xtk := new(big.Int).Exp(g, new(big.Int).Mul(xtk1, xtk2), p)
			xtkList[j] = append(xtkList[j], xtk)
			// fmt.Printf("After adding xtk - j=%d, current xtkList[j] length: %d\n", j, len(xtkList[j]))
		}
		// fmt.Printf("Final len(xtkList[j]) for j=%d: %d\n", j, len(xtkList[j]))
		// fmt.Println("len(xtkList[j]):", len(xtkList[j]))
	}
	return tklList, STKL, xtkList, nil
}

type RES struct {
	val string
	cnt int
}

func (fdxt *FDXT) ServerSearch(n int, tklList []*TKL, stklList []string, xtkList [][]*big.Int) ([]*RES, error) {
	resList := make([]*RES, 0, len(stklList))
	for _, tkl := range tklList {
		l, t := tkl.l, tkl.t
		c, err := base64.StdEncoding.DecodeString(fdxt.CDBXtag[l])
		if err != nil {
			return nil, err
		}
		tBytes, err := base64.StdEncoding.DecodeString(t)
		if err != nil {
			return nil, err
		}
		xtag := utils.BytesXOR(c, tBytes)
		fdxt.XSet[base64.StdEncoding.EncodeToString(xtag)] = 1
	}
	for j, stkl := range stklList {
		cnt := 1
		val, alpha := fdxt.CDBTSet[stkl].val, fdxt.CDBTSet[stkl].alpha
		// fmt.Println("len(xtkList[j]):", len(xtkList[j]))
		for k := 0; k < n-1; k++ {
			xtk := xtkList[j+1][k]
			xtag := new(big.Int).Exp(xtk, alpha, p).Bytes()
			if _, ok := fdxt.XSet[base64.StdEncoding.EncodeToString(xtag)]; ok {
				cnt++
			}
		}
		resList = append(resList, &RES{val: val, cnt: cnt})
	}
	return resList, nil
}

func (fdxt *FDXT) ClientSearchStep2(w1 string, ws []string, resList []*RES) ([]string, error) {
	IDL := make([]string, 0, len(resList))
	for j, res := range resList {
		val, cnt := res.val, res.cnt
		msgLen := len(w1) + len(big.NewInt(int64(j+1)).Bytes()) + 1
		msg := make([]byte, 0, msgLen)
		msg = append(msg, []byte(w1)...)
		msg = append(msg, big.NewInt(int64(j+1)).Bytes()...)
		msg = append(msg, byte(1))
		dec1, err := utils.PrfF(fdxt.Keys[1], msg)
		if err != nil {
			return nil, err
		}
		valBytes, err := base64.StdEncoding.DecodeString(val)
		if err != nil {
			return nil, err
		}
		idOp := utils.BytesXOR(valBytes, dec1)
		op := Operation(idOp[len(idOp)-1])
		// var end int
		// for end = 0; end < len(idOp); end++ {
		// 	if idOp[end] == 0 || idOp[end] == 0x80 {
		// 		break
		// 	}
		// }
		// id := string(idOp[:end])
		id := string(idOp[:len(idOp)-1])
		// fmt.Println("id:", id)
		// fmt.Println("op:", op)
		// fmt.Println("cnt:", cnt)
		if op == Add && cnt == len(ws) {
			IDL = append(IDL, id)
		}
		if op == Delete && cnt > 0 {
			IDL = utils.RemoveElement(IDL, id)
		}
	}
	for _, w := range ws {
		fdxt.Count[w].srch++
		fdxt.Count[w].updt = 0
	}
	return IDL, nil
}
