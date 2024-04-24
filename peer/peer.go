package peer

import (
	"fmt"
	"io"
	"net"
	"redis-clone/proto"

	"github.com/tidwall/resp"
)

type Peer struct {
	Conn      net.Conn
	MsgCh     chan Message
	delPeerCh chan *Peer
}

type Message struct {
	Cmd  proto.Command
	Peer *Peer
}


func NewPeer(conn net.Conn, msgCh chan Message, delCh chan *Peer) *Peer {
	return &Peer{
		Conn:      conn,
		MsgCh:     msgCh,
		delPeerCh: delCh,
	}
}

func (p *Peer) Send(msg []byte) (int, error) {
	return p.Conn.Write(msg)
}

// readLoop will read whatever we receive in the connection and
// sends it to our server via the msg channel
func (p *Peer) ReadLoop() error {
	rd := resp.NewReader(p.Conn)

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
			var cmd proto.Command
			rawCmd := v.Array()[0]
			switch rawCmd.String() {
			case proto.CommandClient:
				cmd = proto.ClientCommand{
					Value: v.Array()[1].String(),
				}
			case proto.CommandSET:
				cmd, err = parseSetCommand(v)
				if err != nil {
					return err
				}
			case proto.CommandGET:
				cmd, err = parseGetCommand(v)
				if err != nil {
					return err
				}
			case proto.CommandHELLO:
				cmd, err = parseHelloCommand(v)
				if err != nil {
					return err
				}
			case proto.CommandCOMMAND:
				cmd, err = parseCommandCommand(v)
				if err != nil {
					return err
				}
			default:
				fmt.Println("That's a case that we cannot handle atm: ", v.Array())
			}

			p.MsgCh <- Message{
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
func parseSetCommand(v resp.Value) (proto.SetCommand, error) {
	// TODO: This is more of a hack we need to do proper parsing/handling with no assumptions
	// because I can send whatever I want here, then we can return proper errors
	if len(v.Array()) != 3 {
		return proto.SetCommand{}, fmt.Errorf("invalid number of variables for SET command")
	}
	cmd := proto.SetCommand{
		Key:   v.Array()[1].Bytes(),
		Value: v.Array()[2].Bytes(),
	}
	return cmd, nil
}

// parseGetCommand is the same thing as parseSetCommand just the number of arguments that's different.
func parseGetCommand(v resp.Value) (proto.GetCommand, error) {
	if len(v.Array()) != 2 {
		return proto.GetCommand{}, fmt.Errorf("invalid number of variables for GET command")
	}
	cmd := proto.GetCommand{
		Key: v.Array()[1].Bytes(),
	}

	return cmd, nil
}

func parseHelloCommand(v resp.Value) (proto.HelloCommand, error) {
	if len(v.Array()) != 2 {
		return proto.HelloCommand{}, fmt.Errorf("invalid number of variables for HELLO command")
	}
	cmd := proto.HelloCommand{
		Value: v.Array()[0].String(),
	}

	return cmd, nil
}

func parseCommandCommand(v resp.Value) (proto.CommandCommand, error) {
	if len(v.Array()) != 1 {
		return proto.CommandCommand{}, fmt.Errorf("invalid number of variables for COMMAND command")
	}
	cmd := proto.CommandCommand{
		Value: v.Array()[0].String(),
	}

	return cmd, nil
}
