package client

import (
	"bytes"
	"context"
	"net"

	"github.com/tidwall/resp"
)

type Client struct {
	addr string
	conn net.Conn
}

func New(addr string) (*Client, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return &Client{
		addr: addr,
		conn: conn,
	}, nil
}

// Set sends a SET RPC to the server.
func (c *Client) Set(ctx context.Context, key string, val any) error {
    buf := &bytes.Buffer{} // NOTE: Hmm this can be done differently

	wr := resp.NewWriter(buf)
	err := wr.WriteArray([]resp.Value{
		resp.StringValue("set"),
		resp.StringValue(key),
		resp.AnyValue(val),
	})
	if err != nil {
		return err
	}

	_, err = c.conn.Write(buf.Bytes())

	buf.Reset()
	return err
}

// Get sends a GET RPC to the server and get's a response back.
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	buf := &bytes.Buffer{}

	wr := resp.NewWriter(buf)
	err := wr.WriteArray([]resp.Value{
		resp.StringValue("get"),
		resp.StringValue(key),
	})
	if err != nil {
		return "", err
	}

	_, err = c.conn.Write(buf.Bytes())
	if err != nil {
		return "", err
	}

	respBuffer := make([]byte, 1024)
	n, err := c.conn.Read(respBuffer)

	return string(respBuffer[:n]), err
}

func (c *Client) Close() error {
	return c.conn.Close()
}
