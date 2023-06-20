package raftService

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type StateMachine struct {
	state map[string]string
}

type Command struct {
	name        string
	description string
	execute     func(commands map[string]Command, state StateMachine, args ...string) (output string, err error)
}

var Commands = map[string]Command{
	"help": {
		name:        "help",
		description: "Provides help for supported commands.",
		execute: func(commands map[string]Command, state StateMachine, args ...string) (string, error) {
			commandName := args[0]
			if command, ok := commands[commandName]; ok {
				return command.description, nil
			}
			return "", errors.New("command not found")
		},
	},
	"put": {
		name: "put",
		description: strings.Join([]string{
			"Set a key to a specified value. The value is stored as a string, but will allow non-string operations if the value is formatted appropriately.",
			"Arguments:",
			"\t0 - The target key",
			"\t1 - The value to store",
		}, "\n"),
		execute: func(commands map[string]Command, state StateMachine, args ...string) (string, error) {
			if len(args) != 2 {
				return "", errors.New("there must be exactly two arguments")
			}
			key := args[0]
			value := args[1]
			state.state[key] = value
			return "Success", nil
		},
	},
	"get": {
		name: "get",
		description: strings.Join([]string{
			"Retrieve the value for a target key. If the key has not been set, an error will be returned.",
			"Arguments:",
			"\t0 - The target key",
		}, "\n"),
		execute: func(commands map[string]Command, state StateMachine, args ...string) (string, error) {
			if len(args) != 1 {
				return "", errors.New("there must be exactly one argument")
			}
			key := args[0]
			if value, ok := state.state[key]; ok {
				return value, nil
			} else {
				return "", errors.New("key not set")
			}
		},
	},
	"append": {
		name: "append",
		description: strings.Join([]string{
			"Appends a value to a JSON array stored under the specified key.",
			"If the key has not been defined, a new JSON array with the provided value will be stored under that key.",
			"Arguments:",
			"\t0 - The target key (must contain a valid JSON array)",
			"\t1 - The string value to append to the array",
		}, "\n"),
		execute: func(commands map[string]Command, state StateMachine, args ...string) (string, error) {
			if len(args) != 2 {
				return "", errors.New("there must be exactly two arguments")
			}
			key := args[0]
			value := args[1]

			var (
				jsonString string
				strBytes   []byte
				err        error
			)
			if currentValue, ok := state.state[key]; ok {
				var parsedArray []string
				err = json.Unmarshal([]byte(currentValue), &parsedArray)
				if err != nil {
					return "", err
				}

				strBytes, err = json.Marshal(append(parsedArray, value))
				if err != nil {
					return "", err
				}
				jsonString = string(strBytes)

				state.state[key] = jsonString
				return "Success - value appended", nil

			} else {
				strBytes, err = json.Marshal([]string{value})
				if err != nil {
					return "", err
				}
				jsonString = string(strBytes)

				state.state[key] = jsonString
				return "Success - new array created", nil
			}
		},
	},
	"add": {
		name: "add",
		description: strings.Join([]string{
			"Add a specified amount to the numeric value stored at the specified key.",
			"Arguments:",
			"\t0 - The target key (must point to a number)",
			"\t1 - The amount to add to the value",
		}, "\n"),
		execute: func(commands map[string]Command, state StateMachine, args ...string) (string, error) {
			if len(args) != 2 {
				return "", errors.New("there must be exactly two arguments")
			}
			key := args[0]
			amountStr := args[1]

			if incAmt, err := strconv.ParseFloat(amountStr, 64); err != nil {
				return "", err
			} else {
				if curAmtStr, ok := state.state[key]; !ok {
					return "", errors.New("unable to parse numeric value at specified")
				} else {
					if curAmt, err := strconv.ParseFloat(curAmtStr, 64); err != nil {
						return "", err
					} else {
						newAmt := strconv.FormatFloat(curAmt+incAmt, 'G', -1, 64)
						state.state[key] = newAmt
						return fmt.Sprintf("Success - new value set to %s", newAmt), nil
					}
				}
			}

		},
	},
}

func (state StateMachine) executeCommand(command string) (string, error) {
	tokens := strings.Split(command, " ")
	if command, ok := Commands[tokens[0]]; ok {
		return command.execute(Commands, state, tokens[1:]...)
	}
	return "", errors.New("operation not recognized")
}
