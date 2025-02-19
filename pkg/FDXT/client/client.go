package client

import (
	"ConjunctiveSSE/pkg/FDXT"
	pb "ConjunctiveSSE/pkg/FDXT/proto"
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type FDXTClient struct {
	fdxt   *FDXT.FDXT
	client pb.FDXTServiceClient
}

func NewFDXTClient(serverAddr string) (*FDXTClient, error) {
	conn, err := grpc.NewClient(serverAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	var fdxt FDXT.FDXT
	Init(&fdxt, "Crime_USENIX_REV")
	return &FDXTClient{
		fdxt:   &fdxt,
		client: pb.NewFDXTServiceClient(conn),
	}, nil
}

func Init(f *FDXT.FDXT, s string) {
	f.Keys[0] = []byte("0123456789123456")
	f.Keys[1] = []byte("0123456789123456")
	f.Keys[2] = []byte("0123456789123456")
	f.Keys[3] = []byte("0123456789123456")
	f.Keys[4] = []byte("0123456789123456")
	f.Count = make(map[string]*FDXT.Counter)
	f.CDBXtag = make(map[string]string)
	f.CDBTSet = make(map[string]*FDXT.TsetValue)
	f.XSet = make(map[string]int, 1000000)
}

func (f *FDXTClient) GetFDXT() *FDXT.FDXT {
	return f.fdxt
}

func (f *FDXTClient) Update(cdbXtag map[string]string, cdbTset map[string]*FDXT.TsetValue) error {
	stream, err := f.client.Update(context.Background())
	if err != nil {
		return err
	}
	// 分批发送数据
	const batchSize = 1000
	count := 0
	batch := &pb.UpdateRequest{
		CDBXtag: make(map[string]string),
		CDBTset: make(map[string]*pb.TsetValue),
	}
	// 发送Xtag
	for k, v := range cdbXtag {
		batch.CDBXtag[k] = v
		count++
		if count >= batchSize {
			if err := stream.Send(batch); err != nil {
				return err
			}
			count = 0
			batch = &pb.UpdateRequest{
				CDBXtag: make(map[string]string),
				CDBTset: make(map[string]*pb.TsetValue),
			}
		}
	}
	// 发送Tset
	for k, v := range cdbTset {
		batch.CDBTset[k] = &pb.TsetValue{
			Val:   v.Val,
			Alpha: v.Alpha.Bytes(),
		}
		count++
		if count >= batchSize {
			if err := stream.Send(batch); err != nil {
				return err
			}
			count = 0
			batch = &pb.UpdateRequest{
				CDBXtag: make(map[string]string),
				CDBTset: make(map[string]*pb.TsetValue),
			}
		}
	}
	// 发送最后一批数据并关闭流
	if count > 0 {
		if err := stream.Send(batch); err != nil {
			return err
		}
	}

	_, err = stream.CloseAndRecv()
	return err
}

func (f *FDXTClient) Search(keywords []string) ([]string, error) {
	// client search step 1
	w1, tkl, stkl, xtkList, err := f.GetFDXT().ClientSearchStep1(keywords)
	if err != nil {
		return nil, err
	}

	// send search request to server
	resp, err := f.client.Search(context.Background(), &pb.SearchRequest{
		TklList:     convertToTklList(tkl),
		StklList:    stkl,
		XtokenLists: convertToXtokenList_2D(xtkList),
		N:           int64(len(keywords)),
	})
	if err != nil {
		return nil, err
	}

	// client search step 2
	return f.GetFDXT().ClientSearchStep2(w1, keywords, convertToRESList(resp.ResList))
}
