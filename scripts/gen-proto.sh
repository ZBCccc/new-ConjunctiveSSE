# 生成proto文件
protoc --go_out=. --go-grpc_out=. pkg/ODXT/proto/odxt.proto
protoc --go_out=. --go-grpc_out=. pkg/HDXT/proto/hdxt.proto
protoc --go_out=. --go-grpc_out=. pkg/FDXT/proto/fdxt.proto
