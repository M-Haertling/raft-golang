package raftService

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	proto "github.com/m-haertling/raft-server/pkg/proto"
	grpc "google.golang.org/grpc"
)

type RaftMode int64

const (
	Leader RaftMode = iota
	Follower
	Candidate
)

type raftService struct {
	mode RaftMode
	proto.UnimplementedRaftServiceServer
	*raftServerState
	invokeHeartbeat func()
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

type raftServerState struct {
	id            int32
	servers       []serverConfiguration
	followerState followerState
	leaderState   leaderState
	SystemState
}

// AppendEntries is invoked by the leader to replicate log entries.
// Leader entries may be rejected if the leader is identified out outdated or indexes were skipped.
func (follower raftService) AppendEntries(ctx context.Context, in *proto.AppendEntriesRequest) (*proto.AppendEntriesResponse, error) {
	if follower.mode != Follower {
		return &proto.AppendEntriesResponse{Term: follower.log.getCurrentTerm(), Success: false}, nil
	}

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

	// Invoke the heartbeat to reset the timer for leader election
	follower.invokeHeartbeat()

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

// RequestVote is invoked by election candidates to gather votes
func (server raftService) RequestVote(ctx context.Context, in *proto.RequestVoteRequest) (*proto.RequestVoteResponse, error) {
	// Rejection scenarios
	if server.log.getCurrentTerm() > in.Term || // The requestor is on an old term - Section 5.1
		server.followerState.votedFor != 0 || // This server has already cast a vote - Section 5.2, 5.4
		server.log.getLastAppliedIndex() > in.LastLogIndex { // This server is ahead in log entries over the requestor  - Section 5.2, 5.4
		return &proto.RequestVoteResponse{Term: server.log.getCurrentTerm(), VoteGranted: false}, nil
	}
	return &proto.RequestVoteResponse{Term: server.log.getCurrentTerm(), VoteGranted: true}, nil
}

func (server raftService) activateElectionClock() {
	// A random wait time helps minimize instances of deadlocked elections
	seed := rand.NewSource(time.Now().UnixMicro())
	randGen := rand.New(seed)
	waitDuration := time.Duration(randGen.Intn(5)+5) * time.Second
	// Start the timer
	timer := time.NewTimer(waitDuration)
	reset := make(chan bool)
	go func() {
		for {
			select {
			case <-timer.C:
				server.runForElection()
				return
			case <-reset:
				if !timer.Stop() {
					<-timer.C
				}
				timer.Reset(waitDuration)
			}
		}
	}()
	// Set the heartbeat function
	server.invokeHeartbeat = func() { reset <- true }
}

func (server raftService) start() {
	server.activateElectionClock()

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

func (server raftService) runForElection() {

}

func (server raftService) revertToFollower() {

}
