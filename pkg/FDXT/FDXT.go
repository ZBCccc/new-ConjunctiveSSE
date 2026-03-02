package FDXT

import (
	"ConjunctiveSSE/pkg/utils"
	"fmt"
	"math/big"
	"sync"
	"time"

	pbcUtil "ConjunctiveSSE/pkg/utils/pbc"

	"github.com/Nik-U/pbc"
)

type Counter struct {
	srch, updt, Max int
}

type TsetValue struct {
	Val   string
	Alpha *pbc.Element
}

type Operation int

const (
	Add Operation = iota
	Delete
)

type TKL struct {
	L, T string
}

type FDXT struct {
	Keys    [5][]byte
	Count   map[string]*Counter
	CDBXtag map[string]string
	CDBTSet map[string]*TsetValue
	XSet    map[string]int
}

var (
	err error
)

func (fdxt *FDXT) Encrypt(keyword string, ids []string, op Operation) (time.Duration, error) {
	kw, kt, kx, ky, kz := fdxt.Keys[0], fdxt.Keys[1], fdxt.Keys[2], fdxt.Keys[3], fdxt.Keys[4]
	if _, ok := fdxt.Count[keyword]; !ok {
		fdxt.Count[keyword] = &Counter{srch: 0, updt: 0, Max: 0}
	}
	clientTime := time.Duration(0)
	for _, id := range ids {
		start := time.Now()
		fdxt.Count[keyword].updt++
		fdxt.Count[keyword].Max++
		msgLen := len(keyword) + len(big.NewInt(int64(fdxt.Count[keyword].Max)).Bytes()) + 1
		msg := make([]byte, 0, msgLen)
		msg = append(msg, []byte(keyword)...)
		msg = append(msg, big.NewInt(int64(fdxt.Count[keyword].Max)).Bytes()...)

		addr, _ := utils.PrfF(kw, append(msg, byte(0)))
		val, _ := utils.PrfF(kw, append(msg, byte(1)))

		val, err = utils.BytesXORWithOp(val, []byte(id), int(op))
		if err != nil {
			return 0, fmt.Errorf("BytesXORWithOp: %w", err)
		}

		msgLen = len(keyword) + len(big.NewInt(int64(fdxt.Count[keyword].srch)).Bytes()) + len(big.NewInt(int64(fdxt.Count[keyword].updt)).Bytes()) + 1
		msg = make([]byte, 0, msgLen)
		msg = append(msg, []byte(keyword)...)
		msg = append(msg, big.NewInt(int64(fdxt.Count[keyword].srch)).Bytes()...)
		msg = append(msg, big.NewInt(int64(fdxt.Count[keyword].updt)).Bytes()...)
		l, _ := utils.PrfF(kt, append(msg, byte(0)))
		t, _ := utils.PrfF(kt, append(msg, byte(1)))
		xtag1, _ := pbcUtil.PrfToZr(kx, []byte(keyword))
		xtag2, _ := pbcUtil.PrfToZr(ky, append([]byte(id), byte(op)))
		xtag := pbcUtil.GToPower2(xtag1, xtag2)
		c := utils.BytesXOR(xtag.Bytes(), t)
		alpha, _, err := utils.ComputeAlpha(ky, kz, []byte(id), int(op), append([]byte(keyword), big.NewInt(int64(fdxt.Count[keyword].Max)).Bytes()...))
		if err != nil {
			return 0, err
		}
		clientTime += time.Since(start)

		// server part
		fdxt.CDBXtag[string(l)] = string(c)
		fdxt.CDBTSet[string(addr)] = &TsetValue{Val: string(val), Alpha: alpha}
	}
	return clientTime, nil
}

func (fdxt *FDXT) ClientSearchStep1(w1 string, q []string) ([]*TKL, []string, [][]*pbc.Element, error) {
	kw, kt, kx, kz := fdxt.Keys[0], fdxt.Keys[1], fdxt.Keys[2], fdxt.Keys[4]
	tklList := make([]*TKL, 0, fdxt.Count[q[0]].updt+fdxt.Count[q[1]].updt)
	for _, w := range q {
		if _, ok := fdxt.Count[w]; !ok {
			return nil, nil, nil, fmt.Errorf("keyword %s not found", w)
		}
		srch, updt := fdxt.Count[w].srch, fdxt.Count[w].updt
		for i := updt; i >= 1; i-- {
			// Pre-calculate total length
			msgLen := len(w) + len(big.NewInt(int64(srch)).Bytes()) + len(big.NewInt(int64(i)).Bytes()) + 1
			msg := make([]byte, 0, msgLen)

			// Build message in one go
			msg = append(msg, []byte(w)...)
			msg = append(msg, big.NewInt(int64(srch)).Bytes()...)
			msg = append(msg, big.NewInt(int64(i)).Bytes()...)

			l, _ := utils.PrfF(kt, append(msg, byte(0)))

			t, _ := utils.PrfF(kt, append(msg, byte(1)))

			tklList = append(tklList, &TKL{L: string(l), T: string(t)})
		}
	}
	STKL := make([]string, fdxt.Count[w1].Max)
	xtkList := make([][]*pbc.Element, fdxt.Count[w1].Max+1)
	qt := utils.RemoveElement(q, w1)
	var wg sync.WaitGroup
	for j := 1; j <= fdxt.Count[w1].Max; j++ {
		wg.Add(1)
		go func(j int) {
			defer wg.Done()
			msg := make([]byte, 0, len(w1)+len(big.NewInt(int64(j)).Bytes())+1)
			msg = append(msg, []byte(w1)...)
			msg = append(msg, big.NewInt(int64(j)).Bytes()...)
			addr, _ := utils.PrfF(kw, append(msg, byte(0)))

			STKL[j-1] = string(addr)
			xtkList[j] = make([]*pbc.Element, 0, len(qt))
			for _, w := range qt {
				xtk1, _ := pbcUtil.PrfToZr(kx, []byte(w))
				xtk2, _ := pbcUtil.PrfToZr(kz, msg)
				xtk := pbcUtil.GToPower2(xtk1, xtk2)
				xtkList[j] = append(xtkList[j], xtk)
			}
		}(j)
	}
	wg.Wait()
	return tklList, STKL, xtkList, nil
}

type RES struct {
	Val string
	Cnt int
}

func (fdxt *FDXT) ClientSearchStep2(w1 string, ws []string, resList []*RES) []string {
	IDL := make([]string, 0, len(resList))
	for j, res := range resList {
		val, cnt := res.Val, res.Cnt
		msgLen := len(w1) + len(big.NewInt(int64(j+1)).Bytes()) + 1
		msg := make([]byte, 0, msgLen)
		msg = append(msg, []byte(w1)...)
		msg = append(msg, big.NewInt(int64(j+1)).Bytes()...)
		msg = append(msg, byte(1))
		dec1, _ := utils.PrfF(fdxt.Keys[1], msg)
		valBytes := []byte(val)
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
	return IDL
}
