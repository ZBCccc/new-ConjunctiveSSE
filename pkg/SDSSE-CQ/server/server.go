package server

import (
	pb "ConjunctiveSSE/pkg/SDSSE-CQ/proto"
	"context"
	"encoding/base64"
	"io"
	"log"
	"sync"

	"github.com/Nik-U/pbc"
	pbcUtil "ConjunctiveSSE/pkg/utils/pbc"
	sseclient "github.com/ZBCccc/Aura/Core/SSEClient"
)

type SDSSEcqServer struct {
	pb.UnimplementedSDSSEcqServiceServer
	TSet *sseclient.SSEClient
	XSet *sseclient.SSEClient
	mu   sync.Mutex
}

func NewSDSSEcqServer() *SDSSEcqServer {
	return &SDSSEcqServer{
		TSet: sseclient.NewSSEClient(),
		XSet: sseclient.NewSSEClient(),
	}
}

func (s *SDSSEcqServer) Setup(stream pb.SDSSEcqService_SetupServer) error {
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&pb.SetupResponse{Success: true})
		}
		if err != nil {
			return err
		}

		s.mu.Lock()
		// Merge TSet
		for keyword, tval := range req.Tset {
			s.TSet.Update(0, keyword, tval)
		}
		// Merge XSet
		for keyword, xval := range req.Xset {
			s.XSet.Update(0, keyword, xval)
		}
		s.mu.Unlock()
	}
}

func (s *SDSSEcqServer) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	op := req.Op
	keyword := req.Keyword
	tval := req.Tval
	xval := req.Xval

	// Apply update to TSet and XSet
	s.TSet.Update(int(op), keyword, tval)
	s.XSet.Update(int(op), keyword, xval)

	return &pb.UpdateResponse{Success: true}, nil
}

func (s *SDSSEcqServer) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	w1 := req.W1
	keywords := req.Keywords
	xtokenLists := req.XtokenLists

	s.mu.Lock()
	defer s.mu.Unlock()

	// Get TSet results for w1
	ResT := s.TSet.Search(w1)
	if ResT == nil {
		return &pb.SearchResponse{ResultList: nil}, nil
	}

	// Build XSet map from all keywords (except w1)
	XSet := make(map[string]bool, 100)
	for _, kw := range keywords {
		ResX := s.XSet.Search(kw)
		if ResX == nil {
			continue
		}
		for _, x := range ResX {
			XSet[x] = true
		}
	}

	// Filter results: for each TSet entry, check if all xtokens match
	Res := make([]string, 0, len(ResT))
	for _, v := range ResT {
		decoded, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			log.Fatal("Failed to decode string", err)
		}

		// Deserialize: e, y, counter
		eLen := decodeUint64(decoded[0:8])
		yLen := decodeUint64(decoded[8:16])

		if int(eLen) > len(decoded)-16 || int(yLen) > len(decoded)-16-int(eLen) {
			continue
		}

		y := pbcUtil.BytesToZr(decoded[16+eLen : 16+eLen+yLen])
		c := int(decodeUint64(decoded[16+eLen+yLen:]))

		// Check if xtokens match for counter c
		flag := true
		if c < len(xtokenLists) {
			for _, xtokenBytes := range xtokenLists[c] {
				xtoken := pbcUtil.BytesToG1(xtokenBytes)
				powed := pbcUtil.Pow(xtoken, y)
				xTagStr := base64.StdEncoding.EncodeToString(powed.Bytes())

				if _, exists := XSet[xTagStr]; !exists {
					flag = false
					break
				}
			}
		} else {
			flag = false
		}

		if flag {
			Res = append(Res, v)
		}
	}

	return &pb.SearchResponse{ResultList: Res}, nil
}

func decodeUint64(data []byte) uint64 {
	var result uint64
	for i := 0; i < 8; i++ {
		result = result<<8 | uint64(data[i])
	}
	return result
}
