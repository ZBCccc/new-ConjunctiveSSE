package server

import (
	"ConjunctiveSSE/pkg/FDXT"
	pb "ConjunctiveSSE/pkg/FDXT/proto"
	"ConjunctiveSSE/pkg/utils"
	pbcUtil "ConjunctiveSSE/pkg/utils/pbc"
	"context"
	"encoding/base64"
	"io"

	"github.com/Nik-U/pbc"
)

type FDXTServer struct {
	pb.UnimplementedFDXTServiceServer
	CDBXtag map[string]string
	CDBTSet map[string]*FDXT.TsetValue
	XSet    map[string]int
}

func NewFDXTServer() *FDXTServer {
	return &FDXTServer{
		CDBXtag: make(map[string]string),
		CDBTSet: make(map[string]*FDXT.TsetValue),
		XSet:    make(map[string]int, 1000000),
	}
}

func (s *FDXTServer) Update(stream pb.FDXTService_UpdateServer) error {
	cdbXtag := make(map[string]string)
	cdbTset := make(map[string]*FDXT.TsetValue)

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			// 流结束，返回结果
			s.CDBXtag = cdbXtag
			s.CDBTSet = cdbTset
			return stream.SendAndClose(&pb.UpdateResponse{})
		}
		if err != nil {
			return err
		}

		// 合并每个批次的 map
		for k, v := range req.CDBTset {
			cdbTset[k] = &FDXT.TsetValue{
				Val:   v.Val,
				Alpha: pbcUtil.BytesToZr(v.Alpha),
			}
		}
		for k, v := range req.CDBXtag {
			cdbXtag[k] = v
		}
	}
}

func (s *FDXTServer) Search(ctx context.Context, req *pb.SearchRequest) (*pb.SearchResponse, error) {
	stklList := req.StklList
	tklList := req.TklList
	xtkList := make([][]*pbc.Element, len(req.XtokenLists))
	n := int(req.N)
	for i, xtokenLists := range req.XtokenLists {
		xtkList[i] = make([]*pbc.Element, sizeOfXtokenList_2D(xtokenLists))
		for j, xtoken := range xtokenLists.XtokenList {
			xtkList[i][j] = pbcUtil.BytesToG1(xtoken)
		}
	}
	resList := make([]*FDXT.RES, 0, len(stklList))
	for _, tkl := range tklList {
		l, t := tkl.L, tkl.T
		c, err := base64.StdEncoding.DecodeString(s.CDBXtag[l])
		if err != nil {
			return nil, err
		}
		tBytes, err := base64.StdEncoding.DecodeString(t)
		if err != nil {
			return nil, err
		}
		xtag := utils.BytesXOR(c, tBytes)
		s.XSet[base64.StdEncoding.EncodeToString(xtag)] = 1
	}
	for j, stkl := range stklList {
		cnt := 1
		val, alpha := s.CDBTSet[stkl].Val, s.CDBTSet[stkl].Alpha
		for k := 0; k < n-1; k++ {
			xtk := xtkList[j+1][k]
			xtag := pbcUtil.Pow(xtk, alpha)
			if _, ok := s.XSet[base64.StdEncoding.EncodeToString(xtag.Bytes())]; ok {
				cnt++
			}
		}
		resList = append(resList, &FDXT.RES{Val: val, Cnt: cnt})
	}
	return &pb.SearchResponse{ResList: convertToResList(resList)}, nil
}

