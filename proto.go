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

// SetCommand our basic representation for the SET command in Redis.
type SetCommand struct {
	key, value []byte
}

// GetCommand our basic representation for the GET command in redis
type GetCommand struct {
	key []byte
}

// parseCommand parses the raw string that we get from the TCP connection
// and it will return the necessary elements for each of the commands we can handle.
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
					return parseGetCommand(v)
				}
			}
		}
	}
	return nil, fmt.Errorf("invalid or unknown command received: %s", raw)
}

// parseSetCommand parses the set command in redis
// basically something like this: "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"
// TL;DR this mainly means:
// *3 => the number of arguments we are sending
// \r\n => separators basically carriage return and new line
// $3 => the length of the first argument
// SET => the first argument you have to care about
// $3 => the length of the next argument
// foo => the second argument
// $3 => the length of the third argument
// bar => the third argument.
// func parseSetCommand(v resp.Value) (Command, error) {
// 	// TODO: This is more of a hack we need to do proper parsing/handling with no assumptions
// 	// because I can send whatever I want here, then we can return proper errors
// 	if len(v.Array()) != 3 {
// 		return nil, fmt.Errorf("invalid number of variables for SET command")
// 	}
// 	cmd := SetCommand{
// 		key:   v.Array()[1].Bytes(),
// 		value: v.Array()[2].Bytes(),
// 	}
// 	return cmd, nil
// }

// // parseGetCommand is the same thing as parseSetCommand just the number of arguments that's different.
// func parseGetCommand(v resp.Value) (Command, error) {
// 	if len(v.Array()) != 2 {
// 		return nil, fmt.Errorf("invalid number of variables for GET command")
// 	}
// 	cmd := GetCommand{
// 		key: v.Array()[1].Bytes(),
// 	}

// 	return cmd, nil
// }
