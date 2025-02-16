package client

import (
	"ConjunctiveSSE/pkg/HDXT"
	pb "ConjunctiveSSE/pkg/HDXT/proto"
)

func convertToIntList(PosList []int32) []int {
	intList := make([]int, len(PosList))
	for i, v := range PosList {
		intList[i] = int(v)
	}
	return intList
}

func convertToPbDK(dkList []*HDXT.Dk) []*pb.DK {
	pbDKList := make([]*pb.DK, len(dkList))
	for i, dk := range dkList {
		pbDKList[i] = &pb.DK{
			L: dk.L,
			R: dk.R,
			D: dk.D,
		}
	}
	return pbDKList
}

func convertToPbUTok(tokList []*HDXT.UTok) []*pb.UTok {
	pbTokList := make([]*pb.UTok, 0, len(tokList))
	for _, tok := range tokList {
		pbTokList = append(pbTokList, &pb.UTok{
			Tok: tok.Tok,
			Op:  pb.Operation(tok.Op),
		})
	}
	return pbTokList
}