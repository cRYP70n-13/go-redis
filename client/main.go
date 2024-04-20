package client

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net"

	"github.com/tidwall/resp"
)

type Client struct {
	addr string
	conn net.Conn
}

func New(addr string) *Client {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		slog.Error("cannot create a connection", "client constructor", err)
	}

	return &Client{
		addr: addr,
        conn: conn,
	}
}

// Set sends a set RPC to the server.
func (c *Client) Set(ctx context.Context, key string, val string) error {
	buf := bytes.Buffer{}

	wr := resp.NewWriter(&buf)
	err := wr.WriteArray([]resp.Value{
		resp.StringValue("SET"),
		resp.StringValue(key),
		resp.StringValue(val),
	})
	if err != nil {
		return err
	}

	// NOTE: This is just some silly bullshit we can just write this slice of bytes directly to the connection but w/e.
	_, err = io.Copy(c.conn, &buf)
	if err != nil {
		return err
	}

	return nil
}
