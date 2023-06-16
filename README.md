# Generating grpc Code
```
go install google.golang.org/protobuf/cmd/protoc-gen-go
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc

export PATH="$PATH:$(go env GOPATH)/bin"

protoc proto/raftService.proto \
--go_out=. \
--go-grpc_out=.
```