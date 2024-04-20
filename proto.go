package main

import (
	"bytes"
	"fmt"
	"io"

	"github.com/tidwall/resp"
)

const (
	CommandSET = "SET"
	CommandGET = "GET"
)

type Command interface{}

type SetCommand struct {
	key, value []byte
}

type GetCommand struct {
	key, value []byte
}

func parseCommand(raw string) (Command, error) {
	rd := resp.NewReader(bytes.NewBufferString(raw))

	for {
		v, _, err := rd.ReadValue()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if v.Type() == resp.Array {
			for _, value := range v.Array() {
				switch value.String() {
				case CommandSET:
					return parseSetCommand(v)
				case CommandGET:
					// TODO: Implement the GET command stuff
					return parseGetCommand(v)
				}
			}
		}
	}
	return nil, fmt.Errorf("invalid or unknown command received: %s", raw)
}

func parseSetCommand(v resp.Value) (Command, error) {
	if len(v.Array()) != 3 {
		return nil, fmt.Errorf("invalid number of variables for SET command")
	}
	cmd := SetCommand{
		key:   v.Array()[1].Bytes(),
		value: v.Array()[2].Bytes(),
	}
	return cmd, nil
}

func parseGetCommand(v resp.Value) (Command, error) {
	if len(v.Array()) != 2 {
		return nil, fmt.Errorf("invalid number of variables for GET command")
	}
	cmd := GetCommand{
		key: v.Array()[1].Bytes(),
	}

	return cmd, nil
}
