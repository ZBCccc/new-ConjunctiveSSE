package FDXT

import (
	"ConjunctiveSSE/pkg/utils"
	pbcUtil "ConjunctiveSSE/pkg/utils/pbc"
	"math/big"
	"testing"
	"time"
)

func TestSearchCommunication(t *testing.T) {
	wk := "F1438"
	w1 := "F27628"
	k := []byte("0123456789123456")
	xtkStorage := 0
	for j := 1; j <= 10; j++ {
		msg := make([]byte, 0, len(w1)+len(big.NewInt(int64(j)).Bytes())+1)
		msg = append(msg, []byte(w1)...)
		msg = append(msg, big.NewInt(int64(j)).Bytes()...)

		xtk1, _ := pbcUtil.PrfToZr(k, []byte(wk))
		xtk2, _ := pbcUtil.PrfToZr(k, msg)
		xtk := pbcUtil.GToPower2(xtk1, xtk2)
		xtkStorage += xtk.BytesLen()

	}
	t.Log("Storage cost:", xtkStorage)
}

func TestSearch(t *testing.T) {
	updt := 16644
	w := "F1438"
	srch := 1
	kt := []byte("0123456789123456")
	tklList := make([]*TKL, 0, updt)
	start := time.Now()
	for i := updt; i >= 1; i-- {
		// 预计算总长度
		msgLen := len(w) + len(big.NewInt(int64(srch)).Bytes()) + len(big.NewInt(int64(i)).Bytes()) + 1
		msg := make([]byte, 0, msgLen)

		// 一次性构建消息
		msg = append(msg, []byte(w)...)
		msg = append(msg, big.NewInt(int64(srch)).Bytes()...)
		msg = append(msg, big.NewInt(int64(i)).Bytes()...)

		l, _ := utils.PrfF(kt, append(msg, byte(0)))

		t, _ := utils.PrfF(kt, append(msg, byte(1)))

		tklList = append(tklList, &TKL{L: string(l), T: string(t)})
	}
	t.Log("Time cost:", time.Since(start))
}
