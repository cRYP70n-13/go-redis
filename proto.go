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
