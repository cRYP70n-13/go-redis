package proto

import (
	"bytes"
	"fmt"

	"github.com/tidwall/resp"
)

const (
	CommandSET     = "SET"
	CommandGET     = "GET"
	CommandHELLO   = "HELLO"
	CommandCLIENT  = "CLIENT"
	CommandCOMMAND = "COMMAND"
	CommandPING    = "PING"
	CommandCONFIG  = "CONFIG"
	CommandEXIST   = "EXISTS"
	CommandDEL     = "DEL"
	CommandINCR    = "INCR"
	CommandDECR    = "DECR"
	CommandLPUSH   = "LPUSH"
)

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

type PingCommand struct {
	Value string
}

type ConfigGetCommand struct {
	Key, Value string
}

type ExistCommand struct {
	Key string
}

type DelCommand struct {
	Key string
}

type IncrCommand struct {
	Key string
}

type DecrCommand struct {
	Key string
}

type LpushCommand struct {
	Key   string
	Value []string
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
