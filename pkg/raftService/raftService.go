package raftService

import (
	"context"
	"fmt"
	"log"
	"net"

	proto "github.com/m-haertling/raft-server/pkg/proto"
	grpc "google.golang.org/grpc"
)

type raftService struct {
	proto.UnimplementedRaftServiceServer
}

func (server raftService) AppendEntries(ctx context.Context, in *proto.AppendEntriesRequest) (*proto.AppendEntriesResponse, error) {
	return nil, nil
}

func (server raftService) RequestVote(ctx context.Context, in *proto.RequestVoteRequest) (*proto.RequestVoteResponse, error) {
	return nil, nil
}

func main() {
	port := 2020
	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("Failed to listen %d: %v", port, err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	proto.RegisterRaftServiceServer(grpcServer, raftService{})
	grpcServer.Serve(listener)
}

type coreState struct {
	currentTerm int
	votedFor    int
	commitIndex int
	lastApplied int
	systemState []SystemState
}

type leaderState struct {
	nextIndex  []int
	matchIndex []int
}

type serverConfiguration struct {
	address string
	port    string
}

type raftServer struct {
	id      int
	servers []serverConfiguration
	coreState
	leaderState
}
