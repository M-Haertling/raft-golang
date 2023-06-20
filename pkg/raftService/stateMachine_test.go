package raftService

import (
	"errors"
	"fmt"
	"testing"
)

func executeCommandSet(state *StateMachine, commands []string) error {
	for idx, command := range commands {
		_, err := state.executeCommand(command)
		if err != nil {
			return errors.Join(fmt.Errorf("an error occurred running the command at index %v -> %s", idx, command), err)
		}
	}
	return nil
}

func TestPut(t *testing.T) {
	state := StateMachine{state: map[string]string{}}
	if _, err := state.executeCommand("put test 100"); err != nil {
		t.Error(err)
	}
	if state.state["test"] != "100" {
		t.Errorf("key test had an unexpected value; expected: 100; actual: %s", state.state["test"])
	}
}

func TestKeyOverwrite(t *testing.T) {
	commands := []string{
		"put test 100",
		"put test 500",
		"get test",
	}
	state := StateMachine{state: map[string]string{}}
	if err := executeCommandSet(&state, commands); err != nil {
		t.Error(err)
	}
	if state.state["test"] != "500" {
		t.Errorf("key test had an unexpected value; expected: 500; actual: %s", state.state["test"])
	}
}

func TestInvalidRetrieve(t *testing.T) {
	commands := []string{
		"put test 500",
		"get test2",
	}
	state := StateMachine{state: map[string]string{}}
	if err := executeCommandSet(&state, commands); err == nil {
		t.Errorf("the retrieval of key test2 should result in an error")
	}
}

func TestHelpValid(t *testing.T) {
	state := StateMachine{state: map[string]string{}}
	if _, err := state.executeCommand("help get"); err != nil {
		t.Error(err)
	}
}

func TestHelpInvalid(t *testing.T) {
	state := StateMachine{state: map[string]string{}}
	if _, err := state.executeCommand("help foo"); err == nil {
		t.Errorf("help foo should have resulted in an error due to the unknown function")
	}
}

func TestAppend(t *testing.T) {
	commands := []string{
		"append test one",
		"append test two",
		"append test three",
		"get test",
	}
	state := StateMachine{state: map[string]string{}}
	if err := executeCommandSet(&state, commands); err != nil {
		t.Error(err)
	}
	expect := "[\"one\",\"two\",\"three\"]"
	if state.state["test"] != expect {
		t.Errorf("key test had an unexpected value; expected: %s; actual: %s", expect, state.state["test"])
	}
}

func TestAppendReservedChar(t *testing.T) {
	commands := []string{
		"append test double-quotes\"",
		"get test",
	}
	state := StateMachine{state: map[string]string{}}
	if err := executeCommandSet(&state, commands); err != nil {
		t.Error(err)
	}
	expect := "[\"double-quotes\\\"\"]"
	if state.state["test"] != expect {
		t.Errorf("key test had an unexpected value; expected: %s; actual: %s", expect, state.state["test"])
	}
}

func TestAppendInvalidJSON1(t *testing.T) {
	commands := []string{
		"put test not.json",
		"append test one",
	}
	state := StateMachine{state: map[string]string{}}
	if err := executeCommandSet(&state, commands); err == nil {
		t.Errorf("command should fail due to an invalid JSON array in test")
	}
}

func TestAppendInvalidJSON2(t *testing.T) {
	commands := []string{
		"put test {\"value\":\"test\"}",
		"append test one",
	}
	state := StateMachine{state: map[string]string{}}
	if err := executeCommandSet(&state, commands); err == nil {
		t.Errorf("command should fail due to an invalid JSON array in test")
	}
}

func TestAddInteger(t *testing.T) {
	commands := []string{
		"put test 100",
		"add test 100",
		"get test",
	}
	state := StateMachine{state: map[string]string{}}
	if err := executeCommandSet(&state, commands); err != nil {
		t.Error(err)
	}
	expect := "200"
	if state.state["test"] != expect {
		t.Errorf("key test had an unexpected value; expected: %s; actual: %s", expect, state.state["test"])
	}
}

func TestAddDecimal(t *testing.T) {
	commands := []string{
		"put test 4.5",
		"add test 6.8",
		"get test",
	}
	state := StateMachine{state: map[string]string{}}
	if err := executeCommandSet(&state, commands); err != nil {
		t.Error(err)
	}
	expect := "11.3"
	if state.state["test"] != expect {
		t.Errorf("key test had an unexpected value; expected: %s; actual: %s", expect, state.state["test"])
	}
}
