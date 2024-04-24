package proto

import (
	"bytes"
	"fmt"

	"github.com/tidwall/resp"
)

const (
	CommandSET     = "set"
	CommandGET     = "get"
	CommandHELLO   = "hello"
	CommandClient  = "client"
	CommandCOMMAND = "COMMAND"
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
