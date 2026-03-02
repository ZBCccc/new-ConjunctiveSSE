package sdssecqClient

import (
	pb "ConjunctiveSSE/pkg/SDSSE-CQ/proto"
	"ConjunctiveSSE/pkg/utils"
	"context"
	"encoding/base64"
	"encoding/binary"
	"log"
	"math"
	"math/big"
	"time"

	pbcUtil "ConjunctiveSSE/pkg/utils/pbc"

	"github.com/Nik-U/pbc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	sseclient "github.com/ZBCccc/Aura/Core/SSEClient"
	util "github.com/ZBCccc/Aura/Util"
)

// SDSSEcqClient wraps the SDSSE-CQ client with gRPC connectivity.
type SDSSEcqClient struct {
	client pb.SDSSEcqServiceClient
	conn   *grpc.ClientConn

	TSet          *sseclient.SSEClient
	XSet          *sseclient.SSEClient
	CT            map[string]int
	k, kx, ki, kz []byte
	iv            []byte
}

// NewSDSSEcqClient creates a new SDSSE-CQ gRPC client.
func NewSDSSEcqClient(serverAddr string) (*SDSSEcqClient, error) {
	conn, err := grpc.NewClient(
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	client := pb.NewSDSSEcqServiceClient(conn)

	return &SDSSEcqClient{
		client: client,
		conn:   conn,
		TSet:   sseclient.NewSSEClient(),
		XSet:   sseclient.NewSSEClient(),
		CT:     make(map[string]int),
		k:      []byte("0123456789123456"),
		kx:     []byte("0123456789123456"),
		ki:     []byte("0123456789123456"),
		kz:     []byte("0123456789123456"),
		iv:     []byte("0123456789123456"),
	}, nil
}

// Close closes the gRPC connection.
func (c *SDSSEcqClient) Close() error {
	return c.conn.Close()
}

// Setup sends the encrypted TSet and XSet to the server via streaming RPC.
func (c *SDSSEcqClient) Setup() error {
	stream, err := c.client.Setup(context.Background())
	if err != nil {
		return err
	}

	// Send empty request to complete setup
	err = stream.Send(&pb.SetupRequest{
		Tset: make(map[string]string),
		Xset: make(map[string]string),
	})
	if err != nil {
		return err
	}

	_, err = stream.CloseAndRecv()
	return err
}

// Update encrypts and sends an update operation to the server.
func (c *SDSSEcqClient) Update(op util.Operation, keyword string, id string) error {
	// Update local CT
	if _, exists := c.CT[keyword]; !exists {
		c.CT[keyword] = -1
	}
	c.CT[keyword]++

	// Compute e, xind, z, y, xTag
	kw, _ := utils.PrfF(c.k, []byte(keyword))
	e, err := util.AesEncrypt([]byte(id), kw, c.iv)
	if err != nil {
		log.Fatal("Failed to AesEncrypt", err)
	}

	xind, _ := pbcUtil.PrfToZr(c.ki, []byte(id))
	z, _ := pbcUtil.PrfToZr(c.kz, append([]byte(keyword), big.NewInt(int64(c.CT[keyword])).Bytes()...))
	y := pbcUtil.ZrDiv(xind, z)
	xTagHead, _ := pbcUtil.PrfToZr(c.kx, []byte(keyword))
	xTag := pbcUtil.GToPower2(xTagHead, xind)

	// Serialize data
	serializedData := serializeData(e, y, c.CT[keyword])
	tval := base64.StdEncoding.EncodeToString(serializedData)
	xval := base64.StdEncoding.EncodeToString(xTag.Bytes())

	// Update local structures
	c.TSet.Update(op, keyword, tval)
	c.XSet.Update(op, keyword, xval)

	// Send to server
	_, err = c.client.Update(context.Background(), &pb.UpdateRequest{
		Op:      int64(op),
		Keyword: keyword,
		Id:      id,
		Tval:    tval,
		Xval:    xval,
	})

	return err
}

// Search performs a conjunctive search via gRPC.
func (c *SDSSEcqClient) Search(keywords []string) ([]string, time.Duration, error) {
	clientStart := time.Now()

	// Find the keyword with minimum count
	minCount := math.MaxInt
	w1 := keywords[0]
	for _, keyword := range keywords {
		if count, exists := c.CT[keyword]; exists {
			if count < minCount {
				minCount = count
				w1 = keyword
			}
		} else {
			return nil, 0, nil
		}
	}

	// Generate xtokenList
	xtokenList := make([][]*pbc.Element, minCount+1)
	for i := range xtokenList {
		xtokenList[i] = make([]*pbc.Element, len(keywords)-1)
	}

	qt := utils.RemoveElement(keywords, w1)
	for i := 0; i <= minCount; i++ {
		for j, wj := range qt {
			xtoken1, _ := pbcUtil.PrfToZr(c.kx, []byte(wj))
			xtoken2, _ := pbcUtil.PrfToZr(c.kz, append([]byte(w1), big.NewInt(int64(i)).Bytes()...))
			xtoken := pbcUtil.GToPower2(xtoken1, xtoken2)
			xtokenList[i][j] = xtoken
		}
	}

	// Convert xtokenList to bytes for gRPC
	xtokenLists := make([][]byte, 0)
	for i := range xtokenList {
		for _, xt := range xtokenList[i] {
			xtokenLists = append(xtokenLists, xt.Bytes())
		}
	}

	clientTime := time.Since(clientStart)

	// Call gRPC Search
	resp, err := c.client.Search(context.Background(), &pb.SearchRequest{
		W1:          w1,
		Keywords:    qt,
		XtokenLists: xtokenLists,
	})
	if err != nil {
		return nil, clientTime, err
	}

	// Decrypt results
	kw1, _ := utils.PrfF(c.k, []byte(w1))
	ResInd := make([]string, 0, len(resp.ResultList))
	for _, v := range resp.ResultList {
		decoded, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			log.Fatal("Failed to decode string", err)
		}
		e, _, _ := deserializeData(decoded)
		ind, err := util.AesDecrypt([]byte(e), kw1, c.iv)
		if err != nil {
			log.Fatal("Failed to AesDecrypt", err)
		}
		ResInd = append(ResInd, string(ind))
	}

	totalTime := time.Since(clientStart)
	return ResInd, totalTime, nil
}

// GetTSet returns the local TSet for reading data
func (c *SDSSEcqClient) GetTSet() *sseclient.SSEClient {
	return c.TSet
}

// GetXSet returns the local XSet for reading data
func (c *SDSSEcqClient) GetXSet() *sseclient.SSEClient {
	return c.XSet
}

// GetCT returns the counter map
func (c *SDSSEcqClient) GetCT() map[string]int {
	return c.CT
}

// GetKeys returns the cryptographic keys
func (c *SDSSEcqClient) GetKeys() ([]byte, []byte, []byte, []byte, []byte) {
	return c.k, c.kx, c.ki, c.kz, c.iv
}

// serializeData encodes e, y, and counter into a single byte slice.
func serializeData(e []byte, y *pbc.Element, counter int) []byte {
	eLen := len(e)
	yBytes := y.Bytes()
	yLen := len(yBytes)

	result := make([]byte, 8+8+eLen+yLen+8)

	binary.BigEndian.PutUint64(result[0:8], uint64(eLen))
	binary.BigEndian.PutUint64(result[8:16], uint64(yLen))

	copy(result[16:16+eLen], e)
	copy(result[16+eLen:16+eLen+yLen], yBytes)
	binary.BigEndian.PutUint64(result[16+eLen+yLen:], uint64(counter))

	return result
}

func deserializeData(data []byte) (e []byte, y *pbc.Element, counter int) {
	eLen := binary.BigEndian.Uint64(data[0:8])
	yLen := binary.BigEndian.Uint64(data[8:16])

	e = data[16 : 16+eLen]
	y = pbcUtil.BytesToZr(data[16+eLen : 16+eLen+yLen])
	counter = int(binary.BigEndian.Uint64(data[16+eLen+yLen:]))

	return
}
