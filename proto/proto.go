package proto

import (
	"bytes"
	"fmt"

	"github.com/tidwall/resp"
)

const (
	CommandSET     = "SET"
	CommandGET     = "GET"
	CommandHELLO   = "hello"
	CommandClient  = "client"
	CommandCOMMAND = "COMMAND"
	CommandPING    = "PING"
	CommandConfig  = "CONFIG"
)

// NOTE: actually this can be not just GET but can also be: SET, RESETSTAT and REWRITE
// Okay now I need a way to handle complex commands

type Command interface{}

// SetCommand our basic representation for the SET command in Redis.
type SetCommand struct {
	Key, Value []byte
}

// GetCommand our basic representation for the GET command in redis
type GetCommand struct {
	Key []byte
}

type HelloCommand struct {
	Value string
}

type ClientCommand struct {
	Value string
}

type CommandCommand struct {
	Value string
}

type CommandPing struct {
	Value string
}

type CommandConfigGet struct {
	Key, Value string	
}

func WriteRespMap(m map[string]string) []byte {
	buf := &bytes.Buffer{}
	buf.WriteString("%" + fmt.Sprintf("%d\r\n", len(m)))
	rw := resp.NewWriter(buf)
	for k, v := range m {
		_ = rw.WriteString(k)
		_ = rw.WriteString(":" + v)
	}
	return buf.Bytes()
}
