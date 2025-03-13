package HDXT

import (
	"ConjunctiveSSE/pkg/utils"
	"bytes"
	"crypto/aes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"math/big"
)

func PrfF(key, message []byte) ([]byte, error) {
    // 检查密钥长度是否为16字节(128位)
    if len(key) != 16 {
        return nil, errors.New("key must be 16 bytes for AES-128")
    }

    // 1. 首先使用AES-ECB-128
    cipher, err := aes.NewCipher(key)
    if err != nil {
        return nil, err
    }

    // 确保消息长度是16字节的倍数
    paddedMessage := pkcs7Padding(message, 16)
    encrypted := make([]byte, len(paddedMessage))

    // 实现ECB模式加密
    for i := 0; i < len(paddedMessage); i += 16 {
        cipher.Encrypt(encrypted[i:i+16], paddedMessage[i:i+16])
    }

    // 2. 然后进行SHA-256哈希
    hash := sha256.Sum256(encrypted)
    
    return hash[:], nil
}


// PKCS7填充
func pkcs7Padding(data []byte, blockSize int) []byte {
    padding := blockSize - len(data)%blockSize
    padText := bytes.Repeat([]byte{byte(padding)}, padding)
    return append(data, padText...)
}


// mitraEncrypt generates encrypted address and value for the given keyword and id.
// Parameters:
//   - hdxt: HDXT instance containing encryption keys and counters
//   - keyword: the keyword to encrypt
//   - id: document identifier
//   - operation: operation type (should be documented what values are valid)
//
// Returns:
//   - string: base64 encoded address
//   - string: base64 encoded value
//   - error: error if any occurred during encryption
func mitraEncrypt(hdxt *HDXT, keyword string, id string, operation int) (string, string, error) {
	hdxt.FileCnt[keyword]++
	k := hdxt.Mitra.Key

	wWc := append([]byte(keyword+"#"), big.NewInt(int64(hdxt.FileCnt[keyword])).Bytes()...)

	// address = PRF(kt, w||wc||0)
	address, err := PrfF(k, append(wWc, byte(0)))
	if err != nil {
		return "", "", err
	}

	// val = PRF(kt, w||wc||1) xor (id||op)
	val, err := PrfF(k, append(wWc, byte(1)))
	if err != nil {
		return "", "", err
	}
	val, err = utils.BytesXORWithOp(val, []byte(id), operation)
	if err != nil {
		return "", "", err
	}
	return base64.StdEncoding.EncodeToString(address), base64.StdEncoding.EncodeToString(val), nil
}

func auhmeEncrypt(hdxt *HDXT, keyword string, id string, va int) (string, string, error) {
	// label = PRF(k1, w||id)
	k1, k2, k3 := hdxt.Auhme.Keys[0], hdxt.Auhme.Keys[1], hdxt.Auhme.Keys[2]
	wid := keyword + "#" + id
	wId := []byte(wid)
	label, err := PrfF(k1, wId)
	if err != nil {
		return "", "", err
	}

	v := append(label, byte(va))
	enc1, err := PrfF(k2, v)
	if err != nil {
		return "", "", err
	}

	v = append(label, byte(0))
	enc2, err := PrfF(k3, v)
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
	Tok map[string]string
	Op  Operation
}

func auhmeGenUpd(hdxt *HDXT, op Operation, ku string, vu int) (*UTok, error) {
	k1, k2, k3 := hdxt.Auhme.Keys[0], hdxt.Auhme.Keys[1], hdxt.Auhme.Keys[2]
	cnt := hdxt.Auhme.Deltas.cnt
	tok := make(map[string]string)
	if op == Add {
		// l ← F (k1, ku )
		l, err := PrfF(k1, []byte(ku))
		if err != nil {
			return nil, err
		}
		tok1, err := PrfF(k2, append(l, byte(vu)))
		if err != nil {
			return nil, err
		}
		tok2, err := PrfF(k3, append(l, byte(cnt)))
		if err != nil {
			return nil, err
		}
		tok[base64.StdEncoding.EncodeToString(l)] = base64.StdEncoding.EncodeToString(utils.Xor(tok1, tok2))
		return &UTok{tok, op}, nil
	}

	err := CInsert(hdxt, ku, vu)
	if err != nil {
		return nil, err
	}

	if len(hdxt.Auhme.Deltas.t)+1 < hdxt.Auhme.Deltas.delta {
		hdxt.Auhme.Deltas.s = nil
		return nil, nil
	} else {
		s := make([]string, 0, len(hdxt.AuhmeCipherList))
		for k := range hdxt.AuhmeCipherList {
			s = append(s, k)
		}
		tok, err = CEvict(hdxt, s)
		if err != nil {
			return nil, err
		}
		CClear(hdxt)
		hdxt.Auhme.Deltas.s, hdxt.Auhme.Deltas.cnt = nil, cnt+1
		return &UTok{tok, Edit}, nil
	}
}

// CInsert inserts a key-value pair into the map. If the key already exists, it deletes the existing key-value pair.
func CInsert(hdxt *HDXT, k string, v int) error {
	k1 := hdxt.Auhme.Keys[0]
	l, err := PrfF(k1, []byte(k))
	if err != nil {
		return err
	}
	delete(hdxt.Auhme.Deltas.t, base64.StdEncoding.EncodeToString(l))
	hdxt.Auhme.Deltas.t[base64.StdEncoding.EncodeToString(l)] = v
	return nil
}

func CEvict(hdxt *HDXT, s []string) (tok map[string]string, err error) {
	k2, k3 := hdxt.Auhme.Keys[1], hdxt.Auhme.Keys[2]
	cnt, t := hdxt.Auhme.Deltas.cnt, hdxt.Auhme.Deltas.t
	tok = make(map[string]string)
	for _, l := range s {
		lBytes, err := base64.StdEncoding.DecodeString(l)
		if err != nil {
			return nil, err
		}
		if _, ok := t[l]; !ok {
			u1, err := PrfF(k3, append(lBytes, byte(cnt)))
			if err != nil {
				return nil, err
			}
			u2, err := PrfF(k3, append(lBytes, byte(cnt+1)))
			if err != nil {
				return nil, err
			}
			tok[l] = base64.StdEncoding.EncodeToString(utils.BytesXOR(u1, u2))
		}
		b := t[l]
		u1, err := PrfF(k2, append(lBytes, byte(b)))
		if err != nil {
			return nil, err
		}
		u2, err := PrfF(k2, append(lBytes, byte(1-b)))
		if err != nil {
			return nil, err
		}
		u3, err := PrfF(k3, append(lBytes, byte(cnt)))
		if err != nil {
			return nil, err
		}
		u4, err := PrfF(k3, append(lBytes, byte(cnt+1)))
		if err != nil {
			return nil, err
		}
		tok[l] = base64.StdEncoding.EncodeToString(utils.BytesXOR(utils.BytesXOR(u1, u2), utils.BytesXOR(u3, u4)))
	}
	return tok, nil
}

func CClear(hdxt *HDXT) {
	// delete all keys in t
	hdxt.Auhme.Deltas.t = make(map[string]int)
}

func auhmeApplyUpd(hdxt *HDXT, utok *UTok) {
	tok, op := utok.Tok, utok.Op
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

func auhmeGenKey(hdxt *HDXT, mp map[string]int) (*Dk, error) {
	k1, k2, k3, cnt := hdxt.Auhme.Keys[0], hdxt.Auhme.Keys[1], hdxt.Auhme.Keys[2], hdxt.Auhme.Deltas.cnt
	L := make([]string, 0, len(mp))
	beta := 1
	xors := make([]byte, 16)
	for k, v := range mp {
		l, err := PrfF(k1, []byte(k))
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
			v1, err := PrfF(k2, append(l, byte(1-v)))
			if err != nil {
				return nil, err
			}
			v2, err := PrfF(k3, append(l, byte(cnt)))
			if err != nil {
				return nil, err
			}
			xors = utils.BytesXOR(xors, utils.BytesXOR(v1, v2))
		} else if cv == -1 {
			v1, err := PrfF(k2, append(l, byte(v)))
			if err != nil {
				return nil, err
			}
			v2, err := PrfF(k3, append(l, byte(cnt)))
			if err != nil {
				return nil, err
			}
			xors = utils.BytesXOR(xors, utils.BytesXOR(v1, v2))
		}
	}
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, err
	}
	r := base64.StdEncoding.EncodeToString(randomBytes)
	if beta == 1 {
		h := sha256.New()
		h.Write(append(randomBytes, xors...))
		d := base64.StdEncoding.EncodeToString(h.Sum(nil))
		return &Dk{L, r, d}, nil
	}
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, err
	}
	d := base64.StdEncoding.EncodeToString(randomBytes)
	return &Dk{L, r, d}, nil
}

type Dk struct {
	L []string
	R string
	D string
}

func (d *Dk) Size() int {
	size := len(d.R) + len(d.D)
	for _, l := range d.L {
		size += len(l)
	}
	return size
}

func CalculateDkListSize(dkList []*Dk) int {
	size := 0
	for _, d := range dkList {
		size += d.Size()
	}
	return size
}

func CFind(hdxt *HDXT, k string) (int, error) {
	k1 := hdxt.Auhme.Keys[0]
	l, err := PrfF(k1, []byte(k))
	if err != nil {
		return -1, err
	}
	if v, ok := hdxt.Auhme.Deltas.t[base64.StdEncoding.EncodeToString(l)]; ok {
		return v, nil
	}
	return -1, nil
}

func auhmeQuery(hdxt *HDXT, dk *Dk) int {
	xors := make([]byte, 16)
	for _, l := range dk.L {
		lBytes, err := base64.StdEncoding.DecodeString(hdxt.AuhmeCipherList[l])
		if err != nil {
			fmt.Println("lbytes decode error")
			return 0
		}
		xors = utils.BytesXOR(xors, lBytes)
	}
	h := sha256.New()
	rBytes, err := base64.StdEncoding.DecodeString(dk.R)
	if err != nil {
		fmt.Println("rbytes decode error")
		return 0
	}
	h.Write(append(rBytes, xors...))
	d := base64.StdEncoding.EncodeToString(h.Sum(nil))
	if d == dk.D {
		return 1
	}
	return 0
}

func MitraGenTrapdoor(hdxt *HDXT, keyword string) ([]string, error) {
	tList := make([]string, 0, hdxt.FileCnt[keyword])
	for i := 1; i <= hdxt.FileCnt[keyword]; i++ {
		//Ti = PrfF(kt, w||i||0)
		address, err := PrfF(hdxt.Mitra.Key, append(append([]byte(keyword+"#"), big.NewInt(int64(i)).Bytes()...), byte(0)))
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

func MitraDecrypt(hdxt *HDXT, keyword string, encs []string) ([]string, error) {
	dec := make([]string, 0, len(encs))
	for i, e := range encs {
		laber, err := PrfF(hdxt.Mitra.Key, append(append([]byte(keyword+"#"), big.NewInt(int64(i+1)).Bytes()...), byte(1)))
		if err != nil {
			return nil, err
		}
		eBytes, err := base64.StdEncoding.DecodeString(e)
		if err != nil {
			return nil, err
		}
		idOp := utils.BytesXOR(eBytes, laber)
		var end int
		for end = 0; end < len(idOp); end++ {
			if idOp[end] == 0 || idOp[end] == 0x80 {
				break
			}
		}
		id := string(idOp[:end])
		dec = append(dec, id)
	}
	return dec, nil
}

func AuhmeClientSearchStep1(hdxt *HDXT, w1Ids []string, q []string) ([]*Dk, error) {
	DK := make([]*Dk, 0, len(w1Ids))
	for _, id := range w1Ids {
		I := make(map[string]int, len(q))
		for _, w := range q {
			I[w+"#"+id] = 1
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

func auhmeServerSearch(hdxt *HDXT, DK []*Dk) []int {
	result := make([]int, 0, len(DK))
	for i, dk := range DK {
		if auhmeQuery(hdxt, dk) == 1 {
			result = append(result, i)
		}
	}
	return result
}

func AuhmeClientSearchStep2(w1Ids []string, posList []int) []string {
	result := make([]string, 0, len(posList))
	for _, pos := range posList {
		result = append(result, w1Ids[pos])
	}
	return result
}
