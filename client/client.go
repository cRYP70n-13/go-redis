package client

import (
	"bytes"
	"context"
	"net"

	"github.com/tidwall/resp"
)

type Client struct {
	addr string
}

func New(addr string) *Client {
	return &Client{
		addr: addr,
	}
}

// Set sends a set RPC to the server.
func (c *Client) Set(ctx context.Context, key string, val string) error {
	// FIXME: This is very bad we are dialing each time we get a SET RPC which should not be the case
	// We have to dial only once and keep this connection ongoing between the client and the server until one of them goes down.
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}

	wr := resp.NewWriter(buf)
	err = wr.WriteArray([]resp.Value{
		resp.StringValue("SET"),
		resp.StringValue(key),
		resp.StringValue(val),
	})
	if err != nil {
		return err
	}

	_, err = conn.Write(buf.Bytes())

	buf.Reset()
	return err
}

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	conn, err := net.Dial("tcp", c.addr)
	if err != nil {
		return "", err
	}

	buf := &bytes.Buffer{}

	wr := resp.NewWriter(buf)
	err = wr.WriteArray([]resp.Value{
		resp.StringValue("GET"),
		resp.StringValue(key),
	})
	if err != nil {
		return "", err
	}

	_, err = conn.Write(buf.Bytes())
	if err != nil {
		return "", err
	}

	respBuffer := make([]byte, 1024)
	n, err := conn.Read(respBuffer)

	return string(respBuffer[:n]), err
}
