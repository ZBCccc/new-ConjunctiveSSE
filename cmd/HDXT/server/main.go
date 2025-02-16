package main

import (
    "log"
    "net"
    "ConjunctiveSSE/pkg/HDXT/server"
    pb "ConjunctiveSSE/pkg/HDXT/proto"
    "google.golang.org/grpc"
)

func main() {
    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }
    
    s := grpc.NewServer(
        grpc.MaxRecvMsgSize(100 * 1024 * 1024),
        grpc.MaxSendMsgSize(100 * 1024 * 1024),
    )
    pb.RegisterHDXTServiceServer(s, server.NewHDXTServer())
    
    log.Printf("server listening at %v", lis.Addr())
    if err := s.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}