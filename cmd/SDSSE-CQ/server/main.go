package main

import (
	pb "ConjunctiveSSE/pkg/SDSSE-CQ/proto"
	"ConjunctiveSSE/pkg/SDSSE-CQ/server"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.MaxRecvMsgSize(100*1024*1024),
		grpc.MaxSendMsgSize(100*1024*1024),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	pb.RegisterSDSSEcqServiceServer(s, server.NewSDSSEcqServer())

	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
