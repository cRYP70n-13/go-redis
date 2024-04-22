package main

import (
	"fmt"
	"io"
	"net"

	"github.com/tidwall/resp"
)

type Peer struct {
	conn      net.Conn
	msgCh     chan Message
	delPeerCh chan *Peer
}

func NewPeer(conn net.Conn, msgCh chan Message, delCh chan *Peer) *Peer {
	return &Peer{
		conn:      conn,
		msgCh:     msgCh,
		delPeerCh: delCh,
	}
}

func (p *Peer) Send(msg []byte) (int, error) {
	return p.conn.Write(msg)
}

// readLoop will read whatever we receive in the connection and
// sends it to our server via the msg channel
func (p *Peer) readLoop() error {
	rd := resp.NewReader(p.conn)

	for {
		v, _, err := rd.ReadValue()
		if err == io.EOF {
			p.delPeerCh <- p
			break
		}
		if err != nil {
			return err
		}

		if v.Type() == resp.Array {
			var cmd Command
			rawCmd := v.Array()[0]
			switch rawCmd.String() {
			case CommandClient:
				cmd = ClientCommand{
					value: v.Array()[1].String(),
				}
			case CommandSET:
				cmd, err = parseSetCommand(v)
				if err != nil {
					return err
				}
			case CommandGET:
				cmd, err = parseGetCommand(v)
				if err != nil {
					return err
				}
			case CommandHELLO:
				cmd, err = parseHelloCommand(v)
				if err != nil {
					return err
				}
			default:
				fmt.Println("That's a case that we cannot handle atm: ", v.Array())
			}

			p.msgCh <- Message{
				Cmd:  cmd,
				Peer: p,
			}
		}
	}
	return nil
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
func parseSetCommand(v resp.Value) (SetCommand, error) {
	// TODO: This is more of a hack we need to do proper parsing/handling with no assumptions
	// because I can send whatever I want here, then we can return proper errors
	if len(v.Array()) != 3 {
		return SetCommand{}, fmt.Errorf("invalid number of variables for SET command")
	}
	cmd := SetCommand{
		key:   v.Array()[1].Bytes(),
		value: v.Array()[2].Bytes(),
	}
	return cmd, nil
}

// parseGetCommand is the same thing as parseSetCommand just the number of arguments that's different.
func parseGetCommand(v resp.Value) (GetCommand, error) {
	if len(v.Array()) != 2 {
		return GetCommand{}, fmt.Errorf("invalid number of variables for GET command")
	}
	cmd := GetCommand{
		key: v.Array()[1].Bytes(),
	}

	return cmd, nil
}

func parseHelloCommand(v resp.Value) (HelloCommand, error) {
	if len(v.Array()) != 2 {
		return HelloCommand{}, fmt.Errorf("invalid number of variables for HELLO command")
	}
	cmd := HelloCommand{
		value: v.Array()[0].String(),
	}

	return cmd, nil
}
