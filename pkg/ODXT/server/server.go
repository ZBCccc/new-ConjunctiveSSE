package server

import (
	"ConjunctiveSSE/pkg/ODXT"
	pb "ConjunctiveSSE/pkg/ODXT/proto"
	"ConjunctiveSSE/pkg/utils"
	pbcUtil "ConjunctiveSSE/pkg/utils/pbc"
	"context"
	"encoding/base64"
	"github.com/Nik-U/pbc"
	"io"
)

type ODXTServer struct {
	pb.UnimplementedODXTServiceServer
	XSet map[string]int
	TSet map[string]*ODXT.TsetValue
}

func NewODXTServer() *ODXTServer {
	return &ODXTServer{
		XSet: make(map[string]int),
		TSet: make(map[string]*ODXT.TsetValue),
	}
}

//func (s *ODXTServer) Update(ctx context.Context, req *pb.UpdateRequest) (*pb.UpdateResponse, error) {
//	s.XSet[req.Xtag] = 1
//	s.TSet[req.Address] = &ODXT.TsetValue{
//		Val:   req.Val,
//		Alpha: pbcUtil.BytesToElement(req.Alpha),
//	}
//	return &pb.UpdateResponse{Success: true}, nil
//}

func (s *ODXTServer) Update(stream pb.ODXTService_UpdateServer) error {
	xSet := make(map[string]int)
	tSet := make(map[string]*ODXT.TsetValue)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// 流结束，返回结果
			s.TSet = tSet
			s.XSet = xSet
			return stream.SendAndClose(&pb.UpdateResponse{})
		}
		if err != nil {
			return err
		}

		// 合并每个批次的 map
		for k, v := range req.TSet {
			tSet[k] = &ODXT.TsetValue{
				Val:   v.Val,
				Alpha: pbcUtil.BytesToZr(v.Alpha),
			}
		}
		for k, v := range req.XSet {
			xSet[k] = int(v)
		}
	}
}

func (s *ODXTServer) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	stokenList := req.StokenList
	xtokenList := make([][]*pbc.Element, len(req.XtokenLists))
	for i, xtokenLists := range req.XtokenLists {
		xtokenList[i] = make([]*pbc.Element, sizeOfXtokenList_2D(xtokenLists))
		for j, xtoken := range xtokenLists.XtokenList {
			xtokenList[i][j] = pbcUtil.BytesToG1(xtoken)
		}
	}
	sEOpList := make([]utils.SEOp, len(stokenList))
	// 搜索数据
	for j, stoken := range stokenList {
		cnt := 1
		val, alpha := s.TSet[stoken].Val, s.TSet[stoken].Alpha
		// 遍历 xtokenList
		for _, xtoken := range xtokenList[j] {
			// 判断 xtag 是否匹配
			xtag := pbcUtil.Pow(xtoken, alpha)
			if _, ok := s.XSet[base64.StdEncoding.EncodeToString(xtag.Bytes())]; ok {
				cnt++
			}
		}
		sEOpList[j] = utils.SEOp{
			J:    j + 1,
			Sval: val,
			Cnt:  cnt,
		}
	}
	return &pb.SearchResponse{SeopList: convertSEOpList(sEOpList)}, nil
}
