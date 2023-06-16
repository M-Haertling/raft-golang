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
	*raftServer
}

type followerState struct {
	votedFor int32
}

type leaderState struct {
	nextIndex  []int32
	matchIndex []int32
}

type serverConfiguration struct {
	address string
	port    string
}

type raftServer struct {
	id            int32
	servers       []serverConfiguration
	followerState followerState
	leaderState   leaderState
	SystemState
}

// AppendEntries is invoked by the leader to replicate log entries.
// Leader entries may be rejected if the leader is identified out outdated or indexes were skipped.
func (follower raftService) AppendEntries(ctx context.Context, in *proto.AppendEntriesRequest) (*proto.AppendEntriesResponse, error) {
	// Whitepaper Section 5.1
	// The leader term should always be equal or greater.
	// If it isn't, this request must be from some outdated leader.
	if in.LeaderTerm < follower.log.getCurrentTerm() {
		return &proto.AppendEntriesResponse{Term: follower.log.getCurrentTerm(), Success: false}, nil
	}

	// Whitepaper Section 5.3
	// We only want to append these new entries if they are successors to the expecter term.
	// This protects against an outdated leader. ???
	if follower.log.getCurrentTerm() != in.LeaderTerm {
		return &proto.AppendEntriesResponse{Term: follower.log.getCurrentTerm(), Success: false}, nil
	}

	// Whitepaper Section 5.3
	// Any index conflicts will take the leader's entries as truth.
	// New entries are appended to the log.

	// Generate a log chain
	var firstEntry *LogEntry
	var chainHead *LogEntry
	for index, inputEntry := range in.GetEntries() {
		e := LogEntry{term: inputEntry.Term, index: inputEntry.Index, data: inputEntry.Data}
		if index == 0 {
			firstEntry = &e
			chainHead = &e
		} else {
			e.previousEntry = chainHead
			chainHead = &e
		}
	}

	// Record entries
	success, err := follower.log.applyEntries(firstEntry, chainHead)
	if err != nil {
		return nil, err
	}

	if success && follower.log.getCommitIndex() < in.LeaderCommitIndex {
		follower.log.commitIndex(in.LeaderCommitIndex)
	}

	return &proto.AppendEntriesResponse{Term: follower.log.getCurrentTerm(), Success: success}, nil
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
