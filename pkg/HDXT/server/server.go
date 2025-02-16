package server

import (
	pb "ConjunctiveSSE/pkg/HDXT/proto"
	"ConjunctiveSSE/pkg/utils"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

type HDXTServer struct {
	pb.UnimplementedHDXTServiceServer
	mitraCipherList map[string]string
	auhmeCipherList map[string]string
}

func NewHDXTServer() *HDXTServer {
	return &HDXTServer{
		mitraCipherList: make(map[string]string),
		auhmeCipherList: make(map[string]string),
	}
}

func (s *HDXTServer) Setup(stream pb.HDXTService_SetupServer) error {
	mitraCipherMap := make(map[string]string)
    auhmeCipherMap := make(map[string]string)
	
	for {
        req, err := stream.Recv()
        if err == io.EOF {
            // 流结束，返回结果
			s.mitraCipherList = mitraCipherMap
			s.auhmeCipherList = auhmeCipherMap
            return stream.SendAndClose(&pb.SetupResponse{})
        }
        if err != nil {
            return err
        }

        // 合并每个批次的 map
        for k, v := range req.MitraCiphers {
            mitraCipherMap[k] = v
        }
        for k, v := range req.AuhmeCiphers {
            auhmeCipherMap[k] = v
        }
    }
}

func (s *HDXTServer) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
	for _, tok := range req.UpdateTokens {
		// 实现 auhmeApplyUpd 逻辑
		auhmeApplyUpd(s.auhmeCipherList, tok)
	}
	return &pb.UpdateResponse{Success: true}, nil
}

func (s *HDXTServer) SearchOneKeyword(ctx context.Context, req *pb.SearchOneKeywordRequest) (*pb.SearchOneKeywordResponse, error) {
	// 实现 mitraServerSearch 逻辑
	result := make([]string, 0, len(req.Trapdoors))
	for _, t := range req.Trapdoors {
		if _, ok := s.mitraCipherList[t]; ok {
			result = append(result, s.mitraCipherList[t])
		}
	}
	return &pb.SearchOneKeywordResponse{EncryptedIds: result}, nil
}

func (s *HDXTServer) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	result := make([]int32, 0, len(req.DkList))
	for i, dk := range req.DkList {
		if auhmeQuery(s.auhmeCipherList, dk) == 1 {
			result = append(result, int32(i))
		}
	}

	return &pb.SearchResponse{
		PosList: result,
	}, nil
}

func auhmeQuery(auhmeCipherList map[string]string, dk *pb.DK) int {
	xors := make([]byte, 16)
	for _, l := range dk.L {
		lBytes, err := base64.StdEncoding.DecodeString(auhmeCipherList[l])
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

func auhmeApplyUpd(auhmeCipherList map[string]string, tok *pb.UTok) {
	for l, v := range tok.Tok {
		if tok.Op == pb.Operation_ADD {
			auhmeCipherList[l] = v
		} else {
			auhmeCipherList[l] = xor(auhmeCipherList[l], v)
		}
	}
}
