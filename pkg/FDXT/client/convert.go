package client

import (
	"ConjunctiveSSE/pkg/FDXT"
	pb "ConjunctiveSSE/pkg/FDXT/proto"

	"github.com/Nik-U/pbc"
)

func convertToTklList(tklList []*FDXT.TKL) []*pb.TKL {
	var res []*pb.TKL
	for _, tkl := range tklList {
		res = append(res, &pb.TKL{
			L: tkl.L,
			T: tkl.T,
		})
	}
	return res
}

func convertToXtokenList_2D(xtokenList [][]*pbc.Element) []*pb.XtokenList_2D {
	xtokenList_2D := make([]*pb.XtokenList_2D, len(xtokenList))
	for i, xtoken := range xtokenList {
		xtokens := make([][]byte, len(xtoken))
		for j, token := range xtoken {
			xtokens[j] = token.Bytes()
		}
		xtokenList_2D[i] = &pb.XtokenList_2D{
			XtokenList: xtokens,
		}
	}
	return xtokenList_2D
}

func convertToRESList(resList []*pb.RES) []*FDXT.RES {
	var res []*FDXT.RES
	for _, r := range resList {
		res = append(res, &FDXT.RES{
			Val: r.Sval,
			Cnt: int(r.Cnt),
		})
	}
	return res
}
