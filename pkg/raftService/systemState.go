package raftService

import (
	"errors"
)

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

var (
	ErrIndexTooHigh = errors.New("the requested index is not in the system (either it has been ingested into the snapshot or the current state is illegal)")
	ErrIndexTooLow  = errors.New("the requested index is newer than the latest recorded index")
)

func (log Log) getCommitIndex() int32 {
	return log.lastEntry.index
}

func (log Log) getLastAppliedIndex() int32 {
	return log.lastEntry.index
}

func (log Log) getCurrentTerm() int32 {
	return log.lastEntry.term
}

func (log Log) get(index int32) (*LogEntry, error) {
	currentEntry := log.lastEntry
	for ; currentEntry.index > index; currentEntry = currentEntry.previousEntry {
	}
	if currentEntry.index > index {
		return nil, ErrIndexTooHigh
	}
	if currentEntry.index < index {
		return nil, ErrIndexTooLow
	}
	return currentEntry, nil
}

func (log Log) applyEntries(firstEntry *LogEntry, entries *LogEntry) (bool, error) {
	conflictLogEntry, err := log.get(firstEntry.index - 1)
	if err != nil {
		switch {
		case errors.Is(err, ErrIndexTooHigh):
			if firstEntry.index == log.lastEntry.index+1 {
				// Append the logs to the end of the chain
				firstEntry.previousEntry = log.lastEntry
				log.lastEntry = entries
				return true, nil
			} else {
				// The leader needs to send an earlier index
				return false, nil
			}
		default:
			// This case should only occur if the startIndex has been incorperated into a snapshot
			// but the snapshot should only contain committed indexes. Therefore, this is not a legal
			// state and implies that a critical bug exists.
			return false, errors.New("The start index has been incorperated into a snapshot. This is an illegal system state.")
		}
	}
	// There were conflicting logs. The leader's logs should be taken as truth.
	firstEntry.previousEntry = conflictLogEntry.previousEntry
	log.lastEntry = entries
	return true, nil
}

func (log Log) commitIndex(index int32) error {
	for commitEntry := log.lastEntry; ; commitEntry = commitEntry.previousEntry {
		if commitEntry.index == index {
			log.lastCommittedEntry = commitEntry
			return nil
		}
		if commitEntry.previousEntry == nil || commitEntry.index < index {
			return errors.New("index not found")
		}
	}
}

type LogEntry struct {
	term          int32
	index         int32
	data          string
	previousEntry *LogEntry
}

type Snapshot struct {
	lastTerm  int32
	lastIndex int32
	state     map[string]string
}
