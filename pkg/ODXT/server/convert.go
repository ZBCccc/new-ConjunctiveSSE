package server

import (
	pb "ConjunctiveSSE/pkg/ODXT/proto"
	"ConjunctiveSSE/pkg/utils"
)

func sizeOfXtokenList_2D(xtokenList_2D *pb.XtokenList_2D) int {
	return len(xtokenList_2D.XtokenList)
}

func convertSEOpList(sEOpList []utils.SEOp) []*pb.SEOp {
	seopList := make([]*pb.SEOp, len(sEOpList))
	for i, seop := range sEOpList {
		seopList[i] = &pb.SEOp{
			J:    int64(seop.J),
			Sval: seop.Sval,
			Cnt:  int64(seop.Cnt),
		}
	}
	return seopList
}
