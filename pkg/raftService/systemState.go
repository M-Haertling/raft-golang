package raftService

import "errors"

type SystemState struct {
	state    map[string]string
	log      *Log
	snapshot *Snapshot
}

func (state SystemState) generateState() error {
	constructedState := make(map[string]string)
	// Apply logs
	for entry := state.log.lastEntry; entry.previousEntry != nil; entry = entry.previousEntry {

	}
	// Fill any gaps with
	for key, value := range state.snapshot.state {
		constructedState[key] = value
	}
	state.state = constructedState
	return nil
}

type Log struct {
	lastEntry          *LogEntry
	lastCommittedEntry *LogEntry
}

func (log Log) commitIndex(index int) error {
	for commitEntry := log.lastEntry; ; commitEntry = commitEntry.previousEntry {
		if commitEntry.index == index {
			log.lastCommittedEntry = commitEntry
			return nil
		}
		if commitEntry.previousEntry == nil || commitEntry.index < index {
			return errors.New("Index not found")
		}
	}
}

type LogEntry struct {
	term          int
	index         int
	data          string
	previousEntry *LogEntry
}

type Snapshot struct {
	lastTerm  int
	lastIndex int
	state     map[string]string
}
