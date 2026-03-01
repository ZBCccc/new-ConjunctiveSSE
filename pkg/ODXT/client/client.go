package client

import (
	"ConjunctiveSSE/pkg/ODXT"
	pb "ConjunctiveSSE/pkg/ODXT/proto"
	"ConjunctiveSSE/pkg/utils"
	pbcUtil "ConjunctiveSSE/pkg/utils/pbc"
	"context"
	"encoding/base64"
	"math"
	"math/big"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type ODXTClient struct {
	odxt   *ODXT.ODXT
	client pb.ODXTServiceClient
}

func NewODXTClient(serverAddr string) (*ODXTClient, error) {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	var odxt ODXT.ODXT
	Init(&odxt)
	return &ODXTClient{
		odxt:   &odxt,
		client: pb.NewODXTServiceClient(conn),
	}, nil
}

func Init(o *ODXT.ODXT) {
	// Fixed keys for reproducible benchmarking as used in the paper.
	o.Keys[0] = []byte("0123456789123456")
	o.Keys[1] = []byte("0123456789123456")
	o.Keys[2] = []byte("0123456789123456")
	o.Keys[3] = []byte("0123456789123456")
	o.UpdateCnt = make(map[string]int)
	o.TSet = make(map[string]*ODXT.TsetValue)
	o.XSet = make(map[string]int)
}

func (c *ODXTClient) GetODXT() *ODXT.ODXT {
	return c.odxt
}

//func (c *ODXTClient) Update(keyword string, ids []string, operation utils.Operation) error {
//	if _, ok := c.odxt.UpdateCnt[keyword]; !ok {
//		c.odxt.UpdateCnt[keyword] = 0
//	}
//	for _, id := range ids {
//		// Generate ciphertext locally
//		xtag, address, val, alpha := Encrypt(c.odxt, keyword, id, operation)
//		_, err := c.client.Update(context.Background(), &pb.UpdateRequest{
//			Xtag:    xtag,
//			Address: address,
//			Val:     val,
//			Alpha:   alpha,
//		})
//		if err != nil {
//			return err
//		}
//	}
//	return nil
//}

func (c *ODXTClient) Update(tSet map[string]*ODXT.TsetValue, xSet map[string]int) error {
	stream, err := c.client.Update(context.Background())
	if err != nil {
		return err
	}
	// Send data in batches
	const batchSize = 1000
	count := 0
	batch := &pb.UpdateRequest{
		TSet: make(map[string]*pb.TsetValue),
		XSet: make(map[string]int64),
	}
	// Send TSet
	for k, v := range tSet {
		batch.TSet[k] = &pb.TsetValue{
			Val:   v.Val,
			Alpha: v.Alpha.Bytes(),
		}
		count++
		if count >= batchSize {
			if err := stream.Send(batch); err != nil {
				return err
			}
			batch = &pb.UpdateRequest{
				TSet: make(map[string]*pb.TsetValue),
				XSet: make(map[string]int64),
			}
			count = 0
		}
	}
	// Send XSet
	for k, v := range xSet {
		batch.XSet[k] = int64(v)
		count++
		if count >= batchSize {
			if err := stream.Send(batch); err != nil {
				return err
			}
			batch = &pb.UpdateRequest{
				TSet: make(map[string]*pb.TsetValue),
				XSet: make(map[string]int64),
			}
			count = 0
		}
	}
	// Send final batch data and close stream
	if count > 0 {
		if err := stream.Send(batch); err != nil {
			return err
		}
	}

	_, err = stream.CloseAndRecv()
	return err
}

func Encrypt(odxt *ODXT.ODXT, keyword string, id string, operation utils.Operation) (string, string, string, []byte) {
	kt, kx, ky, kz := odxt.Keys[0], odxt.Keys[1], odxt.Keys[2], odxt.Keys[3]
	odxt.UpdateCnt[keyword]++
	msgLen := len(keyword) + len(big.NewInt(int64(odxt.UpdateCnt[keyword])).Bytes())
	wWc := make([]byte, 0, msgLen)
	wWc = append(wWc, []byte(keyword)...)
	wWc = append(wWc, big.NewInt(int64(odxt.UpdateCnt[keyword])).Bytes()...)

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

	return base64.StdEncoding.EncodeToString(Xtag.Bytes()), base64.StdEncoding.EncodeToString(address), base64.StdEncoding.EncodeToString(val), alpha.Bytes()
}

func (c *ODXTClient) Search(keywords []string) (time.Duration, time.Duration, []string, error) {
	// Select the keyword with the lowest update count
	counter := math.MaxInt
	var w1 string
	for _, w := range keywords {
		num := c.GetODXT().UpdateCnt[w]
		if num < counter {
			counter = num
			w1 = w
		}
	}

	// client search step 1
	clientTime := time.Now()
	stokenList, xtokenList := c.GetODXT().ClientSearchStep1(w1, keywords)
	clientTimeDura := time.Since(clientTime)

	// send search request
	serverTime := time.Now()
	resp, err := c.client.Search(context.Background(), &pb.SearchRequest{
		StokenList:  stokenList,
		XtokenLists: convertToXtokenList_2D(xtokenList),
	})
	if err != nil {
		return 0, 0, nil, err
	}
	serverTimeDura := time.Since(serverTime)

	// client search step 2
	ids := c.GetODXT().ClientSearchStep2(w1, keywords, convertToSEOp(resp.GetSeopList()))
	return clientTimeDura, serverTimeDura, ids, nil
}
