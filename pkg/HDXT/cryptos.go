package HDXT

import (
	"ConjunctiveSSE/pkg/utils"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"log"
	"math/big"
	"strings"
)

func mitraEncrypt(hdxt *HDXT, keyword string, id string, operation int) (string, string, error) {
	hdxt.FileCnt[keyword]++
	k := hdxt.Mitra.Key

	wWc := append([]byte(keyword), big.NewInt(int64(hdxt.FileCnt[keyword])).Bytes()...)

	// address = PRF(kt, w||wc||0)
	address, err := utils.PrfF(k, append(wWc, big.NewInt(int64(0)).Bytes()...))
	if err != nil {
		return "", "", err
	}

	// val = PRF(kt, w||wc||1) xor (id||op)
	val, err := utils.PrfF(k, append(wWc, big.NewInt(int64(1)).Bytes()...))
	if err != nil {
		return "", "", err
	}
	val, err = utils.BytesXORWithOp(val, []byte(id), operation)
	if err != nil {
		return "", "", err
	}

	return base64.StdEncoding.EncodeToString(address), base64.StdEncoding.EncodeToString(val), nil
}

func auhmeEncrypt(hdxt *HDXT, keyword string, id string, flag int, cnt int) (string, string, error) {
	// label = PRF(k1, w||id)
	k1, k2, k3 := hdxt.Auhme.Keys[0], hdxt.Auhme.Keys[1], hdxt.Auhme.Keys[2]
	wId := append([]byte(keyword), []byte(id)...)
	label, err := utils.FAesni(k1, wId, 1)
	if err != nil {
		return "", "", err
	}

	v := append([]byte(label), byte(flag))
	enc1, err := utils.FAesni(k2, v, 1)
	if err != nil {
		return "", "", err
	}

	v = append([]byte(label), byte(cnt))
	enc2, err := utils.FAesni(k3, v, 1)
	if err != nil {
		return "", "", err
	}

	// xor enc1 and enc2
	enc := utils.Xor(enc1, enc2)
	return base64.StdEncoding.EncodeToString(label), base64.StdEncoding.EncodeToString(enc), nil
}

type Operation int

const (
	Add Operation = iota
	Edit
	EditMinus
	EditPlus
)

type Delta struct {
	cnt   int
	t     map[string]int
	delta int
	s     []string
}

type UTok struct {
	tok map[string]string
	op  Operation
}

func auhmeGenUpd(hdxt *HDXT, op Operation, ku string, vu int) (*UTok, *Delta, error) {
	k1, k2, k3 := hdxt.Auhme.Keys[0], hdxt.Auhme.Keys[1], hdxt.Auhme.Keys[2]
	cnt, s, t, delta := hdxt.Auhme.Deltas.cnt, hdxt.Auhme.Deltas.s, hdxt.Auhme.Deltas.t, hdxt.Auhme.Deltas.delta
	tok := make(map[string]string)
	if op == Add {
		// l ← F (k1, ku )
		l, err := utils.FAesni(k1, []byte(ku), 1)
		if err != nil {
			return nil, nil, err
		}
		tok1, err := utils.FAesni(k2, append([]byte(l), byte(vu)), 1)
		if err != nil {
			return nil, nil, err
		}
		tok2, err := utils.FAesni(k3, append([]byte(l), byte(cnt)), 1)
		if err != nil {
			return nil, nil, err
		}
		tok[base64.StdEncoding.EncodeToString(l)] = base64.StdEncoding.EncodeToString(utils.Xor(tok1, tok2))
		return &UTok{tok, op}, &Delta{cnt, t, delta, s}, nil
	}

	var err error
	t, err = CInsert(k1, ku, vu, t)
	if err != nil {
		return nil, nil, err
	}

	if len(t)+1 < delta {
		return nil, &Delta{cnt, t, delta, nil}, nil
	} else {
		s = make([]string, 0)
		for k := range hdxt.AuhmeCipherList {
			s = append(s, k)
		}
		tok, err = CEvict(hdxt, s)
		if err != nil {
			return nil, nil, err
		}
		CClear(hdxt)
		return &UTok{tok, Edit}, &Delta{cnt + 1, t, delta, nil}, nil
	}
}

// CInsert inserts a key-value pair into the map. If the key already exists, it deletes the existing key-value pair.
func CInsert(k1 []byte, k string, v int, t map[string]int) (map[string]int, error) {
	l, err := utils.FAesni(k1, []byte(k), 1)
	if err != nil {
		return nil, err
	}
	delete(t, base64.StdEncoding.EncodeToString(l))
	t[base64.StdEncoding.EncodeToString(l)] = v
	return t, nil
}

func CEvict(hdxt *HDXT, s []string) (tok map[string]string, err error) {
	k2, k3 := hdxt.Auhme.Keys[1], hdxt.Auhme.Keys[2]
	cnt, t := hdxt.Auhme.Deltas.cnt, hdxt.Auhme.Deltas.t
	tok = make(map[string]string)
	for _, l := range s {
		if _, ok := t[l]; !ok {
			u1, err := utils.FAesni(k3, append([]byte(l), byte(cnt)), 1)
			if err != nil {
				return nil, err
			}
			u2, err := utils.FAesni(k3, append([]byte(l), byte(cnt+1)), 1)
			if err != nil {
				return nil, err
			}
			tok[l] = base64.StdEncoding.EncodeToString(utils.Xor(u1, u2))
		}
		b := t[l]
		u1, err := utils.FAesni(k2, append([]byte(l), byte(b)), 1)
		if err != nil {
			return nil, err
		}
		u2, err := utils.FAesni(k2, append([]byte(l), byte(1-b)), 1)
		if err != nil {
			return nil, err
		}
		u3, err := utils.FAesni(k3, append([]byte(l), byte(cnt)), 1)
		if err != nil {
			return nil, err
		}
		u4, err := utils.FAesni(k3, append([]byte(l), byte(cnt+1)), 1)
		if err != nil {
			return nil, err
		}
		tok[l] = base64.StdEncoding.EncodeToString(utils.Xor(utils.Xor(u1, u2), utils.Xor(u3, u4)))
	}
	return tok, nil
}

func CClear(hdxt *HDXT) {
	// delete all keys in t
	hdxt.Auhme.Deltas.t = make(map[string]int)
}

func auhmeApplyUpd(hdxt *HDXT, utok *UTok) {
	tok, op := utok.tok, utok.op
	for l, v := range tok {
		if op == Add {
			hdxt.AuhmeCipherList[l] = v
		} else {
			dec := xor(hdxt.AuhmeCipherList[l], v)
			hdxt.AuhmeCipherList[l] = dec
		}
	}
}

func xor(s1, s2 string) string {
	// 将字符串转换为字节切片
	b1 := []byte(s1)
	b2 := []byte(s2)

	// 获取较短的长度
	minLen := len(b1)
	if len(b2) < minLen {
		minLen = len(b2)
	}

	// 使用较长的切片作为结果
	var result []byte
	if len(b1) > len(b2) {
		result = make([]byte, len(b1))
		copy(result, b1)
	} else {
		result = make([]byte, len(b2))
		copy(result, b2)
	}

	// 对最小长度的部分进行异或操作
	for i := 0; i < minLen; i++ {
		result[i] = b1[i] ^ b2[i]
	}

	return string(result)
}

func auhmeGenKey(hdxt *HDXT, mp map[string]int) (*dk, error) {
	k1, k2, k3 := hdxt.Auhme.Keys[0], hdxt.Auhme.Keys[1], hdxt.Auhme.Keys[2]
	cnt := hdxt.Auhme.Deltas.cnt
	L := make([]string, 0)
	beta := 1
	xors := strings.Repeat("0", 16)
	for k, v := range mp {
		l, err := utils.FAesni(k1, []byte(k), 1)
		if err != nil {
			return nil, err
		}
		L = append(L, base64.StdEncoding.EncodeToString(l))
		cv, err := CFind(hdxt, k)
		if err != nil {
			return nil, err
		}
		if cv == 1-v {
			beta = 0
		} else if cv == v {
			v1, err := utils.FAesni(k2, append([]byte(l), byte(1-v)), 1)
			if err != nil {
				return nil, err
			}
			v2, err := utils.FAesni(k3, append([]byte(l), byte(cnt)), 1)
			if err != nil {
				return nil, err
			}
			xors = xor(xors, base64.StdEncoding.EncodeToString(utils.Xor(v1, v2)))
		} else if cv == -1 {
			v1, err := utils.FAesni(k2, append(l, byte(v)), 1)
			if err != nil {
				return nil, err
			}
			v2, err := utils.FAesni(k3, append(l, byte(cnt)), 1)
			if err != nil {
				return nil, err
			}
			xors = xor(xors, base64.StdEncoding.EncodeToString(utils.Xor(v1, v2)))
		}
	}
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, err
	}
	r := base64.StdEncoding.EncodeToString(randomBytes)
	if beta == 1 {
		h := sha256.New()
		h.Write([]byte(r + xors))
		d := base64.StdEncoding.EncodeToString(h.Sum(nil))
		return &dk{L, r, d}, nil
	}
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, err
	}
	d := base64.StdEncoding.EncodeToString(randomBytes)
	return &dk{L, r, d}, nil
}

type dk struct {
	L []string
	r string
	d string
}

func CFind(hdxt *HDXT, k string) (int, error) {
	k1 := hdxt.Auhme.Keys[0]
	l, err := utils.FAesni(k1, []byte(k), 1)
	if err != nil {
		return -1, err
	}
	if v, ok := hdxt.Auhme.Deltas.t[base64.StdEncoding.EncodeToString(l)]; ok {
		return v, nil
	}
	return -1, nil
}

func auhmeQuery(hdxt *HDXT, dk *dk) int {
	xors := strings.Repeat("0", 16)
	for _, l := range dk.L {
		xors = xor(xors, hdxt.AuhmeCipherList[l])
	}
	h := sha256.New()
	h.Write([]byte(dk.r + xors))
	d := base64.StdEncoding.EncodeToString(h.Sum(nil))
	if d == dk.d {
		return 1
	}
	return 0
}

func mitraGenTrapdoor(hdxt *HDXT, keyword string) ([]string, error) {
	tList := make([]string, 0, hdxt.FileCnt[keyword])
	for i := 1; i <= hdxt.FileCnt[keyword]; i++ {
		//Ti = PrfF(kt, w||i||0)
		address, err := utils.PrfF(hdxt.Mitra.Key, append(append([]byte(keyword), big.NewInt(int64(i)).Bytes()...), byte(0)))
		if err != nil {
			return nil, err
		}
		tList = append(tList, base64.StdEncoding.EncodeToString(address))
	}
	return tList, nil
}

func mitraServerSearch(hdxt *HDXT, tList []string) []string {
	result := make([]string, 0, len(tList))
	for _, t := range tList {
		if _, ok := hdxt.MitraCipherList[t]; ok {
			result = append(result, hdxt.MitraCipherList[t])
		}
	}
	return result
}

func mitraDecrypt(hdxt *HDXT, keyword string, encs []string) ([]string, error) {
	dec := make([]string, 0, len(encs))
	for i, e := range encs {
		laber, err := utils.PrfF(hdxt.Mitra.Key, append(append([]byte(keyword), big.NewInt(int64(i)).Bytes()...), byte(1)))
		if err != nil {
			return nil, err
		}
		eBytes, err := base64.StdEncoding.DecodeString(e)
		if err != nil {
			return nil, err
		}
		idOp := utils.BytesXOR(eBytes, laber)
		dec = append(dec, string(idOp[:len(idOp)-1]))
	}
	return dec, nil
}

func auhmeClientSearchStep1(hdxt *HDXT, w1Ids []string, q []string) ([]*dk, error) {
	DK := make([]*dk, 0, len(w1Ids))
	for _, id := range w1Ids {
		I := make(map[string]int, len(q))
		for _, w := range q {
			I[w+id] = 1
		}
		dk, err := auhmeGenKey(hdxt, I)
		if err != nil {
			log.Println(err)
			return nil, err
		}
		DK = append(DK, dk)
	}
	return DK, nil
}

func auhmeServerSearch(hdxt *HDXT, DK []*dk) []int {
	result := make([]int, 0, len(DK))
	for i, dk := range DK {
		if auhmeQuery(hdxt, dk) == 1 {
			result = append(result, i)
		}
	}
	return result
}

func auhmeClientSearchStep2(w1Ids []string, posList []int) []string {
	result := make([]string, 0, len(posList))
	for _, pos := range posList {
		result = append(result, w1Ids[pos])
	}
	return result
}
