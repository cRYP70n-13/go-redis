package main

import (
	"bytes"
	"fmt"

	"github.com/tidwall/resp"
)

const (
	CommandSET    = "set"
	CommandGET    = "get"
	CommandHELLO  = "hello"
	CommandClient = "client"
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

type HelloCommand struct {
	value string
}

type ClientCommand struct {
	value string
}

func writeRespMap(m map[string]string) []byte {
	buf := &bytes.Buffer{}
	buf.WriteString("%" + fmt.Sprintf("%d\r\n", len(m)))
	rw := resp.NewWriter(buf)
	for k, v := range m {
        _ = rw.WriteString(k)
        _ = rw.WriteString(":" + v)
	}
	return buf.Bytes()
}
