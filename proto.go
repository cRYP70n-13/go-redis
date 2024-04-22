package main

import (
	"bytes"
)

const (
	CommandSET   = "set"
	CommandGET   = "get"
	CommandHELLO = "hello"
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

// TODO: Change this to set whatever we have inside the incoming map
func writeRespMap(m map[string]string) []byte {
	buf := &bytes.Buffer{}
	buf.WriteString("%2\r\n+first\r\n:1\r\n+second\r\n:2\r\n")
	// buf.WriteString("%" + fmt.Sprintf("%d\r\n", len(m)))
	// wr := resp.NewWriter(buf)
	// for k, v := range m {
	// 	wr.WriteString(k)
	// 	wr.WriteString(":" + v)
	// }

	return buf.Bytes()
}
