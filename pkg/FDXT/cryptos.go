package FDXT

import (
	"ConjunctiveSSE/pkg/utils"
	pbcUtil "ConjunctiveSSE/pkg/utils/pbc"
	"encoding/base64"
	"fmt"
	"math"
	"math/big"
	"sync"

	"github.com/Nik-U/pbc"
)

type Operation int

const (
	Add Operation = iota
	Delete
)

type TKL struct {
	L, T string
}

func (fdxt *FDXT) ClientSearchStep1(q []string) (string, []*TKL, []string, [][]*pbc.Element, error) {
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
			return "", nil, nil, nil, fmt.Errorf("keyword %s not found", w)
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
				return "", nil, nil, nil, err
			}
			t, err := utils.PrfF(kt, append(msg, byte(1)))
			if err != nil {
				return "", nil, nil, nil, err
			}
			tklList = append(tklList, &TKL{L: base64.StdEncoding.EncodeToString(l), T: base64.StdEncoding.EncodeToString(t)})
		}
	}
	STKL := make([]string, 0, 1000)
	xtkList := make([][]*pbc.Element, fdxt.Count[w1].max+1)
	qt := utils.RemoveElement(q, w1)
	for j := 1; j <= fdxt.Count[w1].max; j++ {
		msg := make([]byte, 0, len(w1)+len(big.NewInt(int64(j)).Bytes())+1)
		msg = append(msg, []byte(w1)...)
		msg = append(msg, big.NewInt(int64(j)).Bytes()...)
		addr, err := utils.PrfF(kw, append(msg, byte(0)))
		if err != nil {
			return "", nil, nil, nil, err
		}
		STKL = append(STKL, base64.StdEncoding.EncodeToString(addr))
		xtkList[j] = make([]*pbc.Element, 0, len(qt))
		for _, w := range qt {
			xtk1, err := pbcUtil.PrfToZr(kx, []byte(w))
			if err != nil {
				return "", nil, nil, nil, err
			}
			xtk2, err := pbcUtil.PrfToZr(kz, msg)
			if err != nil {
				return "", nil, nil, nil, err
			}
			xtk := pbcUtil.GToPower2(xtk1, xtk2)
			xtkList[j] = append(xtkList[j], xtk)
		}
	}
	return w1, tklList, STKL, xtkList, nil
}

type RES struct {
	Val string
	Cnt int
}

func (fdxt *FDXT) ServerSearch(n int, tklList []*TKL, stklList []string, xtkList [][]*pbc.Element) ([]*RES, error) {
	resList := make([]*RES, len(stklList))
	for _, tkl := range tklList {
		l, t := tkl.L, tkl.T
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
	var wg sync.WaitGroup
	for j := range stklList {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			cnt := 1
			val, alpha := fdxt.CDBTSet[stklList[j]].Val, fdxt.CDBTSet[stklList[j]].Alpha
			for k := 0; k < n-1; k++ {
				xtk := xtkList[j+1][k]
				xtag := pbcUtil.Pow(xtk, alpha)
				if _, ok := fdxt.XSet[base64.StdEncoding.EncodeToString(xtag.Bytes())]; ok {
					cnt++
				}
			}
			resList = append(resList, &RES{Val: val, Cnt: cnt})
		}(j)
		wg.Wait()
	}
	return resList, nil
}

func (fdxt *FDXT) ClientSearchStep2(w1 string, ws []string, resList []*RES) ([]string, error) {
	IDL := make([]string, 0, len(resList))
	for j, res := range resList {
		val, cnt := res.Val, res.Cnt
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
		id := string(idOp[:len(idOp)-1])
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
