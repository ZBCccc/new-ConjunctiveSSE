package server

import (
	"ConjunctiveSSE/pkg/FDXT"
	pb "ConjunctiveSSE/pkg/FDXT/proto"
)

func sizeOfXtokenList_2D(xtokenList_2D *pb.XtokenList_2D) int {
	return len(xtokenList_2D.XtokenList)
}


func convertToResList(resList []*FDXT.RES) []*pb.RES {
	res := make([]*pb.RES, 0, len(resList))
	for _, r := range resList {
		res = append(res, &pb.RES{
			Sval: r.Val,
			Cnt:  int64(r.Cnt),
		})
	}
	return res
}