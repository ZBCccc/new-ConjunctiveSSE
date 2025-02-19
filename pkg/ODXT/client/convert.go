package client

import (
	pb "ConjunctiveSSE/pkg/ODXT/proto"
	"ConjunctiveSSE/pkg/utils"
	"github.com/Nik-U/pbc"
)

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

func convertToSEOp(sEopList []*pb.SEOp) []utils.SEOp {
	seopList := make([]utils.SEOp, len(sEopList))
	for i, seop := range sEopList {
		seopList[i] = utils.SEOp{
			J:    int(seop.J),
			Sval: seop.Sval,
			Cnt:  int(seop.Cnt),
		}
	}
	return seopList
}
