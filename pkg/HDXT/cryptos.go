package HDXT

import (
	"ConjunctiveSSE/pkg/utils"
	"encoding/base64"
	"math/big"
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
