package client

import (
	"ConjunctiveSSE/pkg/HDXT"
	pb "ConjunctiveSSE/pkg/HDXT/proto"
	"ConjunctiveSSE/pkg/utils"
	"context"
	"log"
	"math"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type HDXTClient struct {
	hdxt   *HDXT.HDXT
	client pb.HDXTServiceClient
	conn   *grpc.ClientConn
}

func NewHDXTClient(serverAddr, dbName, mongoURI string) (*HDXTClient, error) {
	conn, err := grpc.NewClient(
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(100*1024*1024), // 100MB
			grpc.MaxCallSendMsgSize(100*1024*1024), // 100MB
		),
		grpc.WithTimeout(10*time.Minute),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, err
	}
	var hdxt HDXT.HDXT
	hdxt.Init(dbName, false, mongoURI)
	return &HDXTClient{
		hdxt:   &hdxt,
		client: pb.NewHDXTServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *HDXTClient) GetHDXT() *HDXT.HDXT {
	return c.hdxt
}

func (c *HDXTClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *HDXTClient) Setup(mitraCipherList map[string]string, auhmeCipherList map[string]string) error {
	// Check connection state
	state := c.conn.GetState()
	log.Printf("Connection state before Setup: %v", state)
	// Send ciphertext to server
	stream, err := c.client.Setup(context.Background())
	if err != nil {
		return err
	}

	// Send data in batches
	const batchSize = 1000
	count := 0
	batch := &pb.SetupRequest{
		MitraCiphers: make(map[string]string),
		AuhmeCiphers: make(map[string]string),
	}

	// Send MitraCiphers
	for k, v := range mitraCipherList {
		batch.MitraCiphers[k] = v
		count++

		if count >= batchSize {
			if err := stream.Send(batch); err != nil {
				return err
			}
			batch = &pb.SetupRequest{
				MitraCiphers: make(map[string]string),
				AuhmeCiphers: make(map[string]string),
			}
			count = 0
		}
	}

	// Send AuhmeCiphers
	for k, v := range auhmeCipherList {
		batch.AuhmeCiphers[k] = v
		count++

		if count >= batchSize {
			if err := stream.Send(batch); err != nil {
				return err
			}
			batch = &pb.SetupRequest{
				MitraCiphers: make(map[string]string),
				AuhmeCiphers: make(map[string]string),
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
	if err != nil {
		return err
	}
	state = c.conn.GetState()
	log.Printf("Connection state after Setup: %v", state)
	return err
}

func (c *HDXTClient) Update(id string, keywords []string, operation HDXT.Operation) error {
	// Generate ciphertext locally
	_, tokList, err := c.hdxt.Encrypt(id, keywords, operation)
	if err != nil {
		return err
	}

	// Send ciphertext to server
	_, err = c.client.Update(context.Background(), &pb.UpdateRequest{
		UpdateTokens: convertToPbUTok(tokList),
	})
	return err
}

func (c *HDXTClient) SearchOneKeyword(keyword string) ([]string, error) {
	// Generate trapdoor
	tList, err := HDXT.MitraGenTrapdoor(c.hdxt, keyword)
	if err != nil {
		return nil, err
	}
	resp, err := c.client.SearchOneKeyword(context.Background(), &pb.SearchOneKeywordRequest{
		Trapdoors: tList,
	})
	if err != nil {
		return nil, err
	}
	ids, err := HDXT.MitraDecrypt(c.hdxt, keyword, resp.GetEncryptedIds())
	if err != nil {
		return nil, err
	}
	return ids, nil
}

func (c *HDXTClient) Search(keywords []string) ([]string, error) {
	// Single keyword search, mitra part
	// Select the keyword with the lowest query frequency
	counter, w1 := math.MaxInt64, keywords[0]
	for _, w := range keywords {
		num := c.hdxt.FileCnt[w]
		if num < counter {
			w1 = w
			counter = num
		}
	}
	w1Ids, err := c.SearchOneKeyword(w1)
	if err != nil {
		return nil, err
	}

	// auhme part
	// client search step 1
	q := utils.RemoveElement(keywords, w1)
	dkList, err := HDXT.AuhmeClientSearchStep1(c.hdxt, w1Ids, q)
	if err != nil {
		return nil, err
	}

	// Send search request
	resp, err := c.client.Search(context.Background(), &pb.SearchRequest{
		DkList: convertToPbDK(dkList),
	})
	if err != nil {
		return nil, err
	}

	// client search step 2
	sIdList := HDXT.AuhmeClientSearchStep2(w1Ids, convertToIntList(resp.PosList))
	return sIdList, nil
}
