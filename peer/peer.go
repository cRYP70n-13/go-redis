package peer

import (
	"fmt"
	"io"
	"net"
	"strings"

	"redis-clone/proto"

	"github.com/tidwall/resp"
)

type Peer struct {
	Conn      net.Conn
	msgCh     chan Message
	errorsCh  chan Errors
	delPeerCh chan *Peer
}

type Message struct {
	Cmd  proto.Command
	Peer *Peer
}

type Errors struct {
	Err  error
	Peer *Peer
}

func NewPeer(conn net.Conn, msgCh chan Message, delCh chan *Peer, errorsCh chan Errors) *Peer {
	return &Peer{
		Conn:      conn,
		errorsCh:  errorsCh,
		msgCh:     msgCh,
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
			cmd, err := parseCommand(v)
			if err != nil {
				fmt.Println("Error parsing command:", err)
				p.errorsCh <- Errors{
					Err:  err,
					Peer: p,
				}
				continue
			}

			p.msgCh <- Message{
				Cmd:  cmd,
				Peer: p,
			}
		}
	}
	return nil
}

func parseCommand(v resp.Value) (proto.Command, error) {
	if len(v.Array()) < 1 {
		return nil, fmt.Errorf("invalid command format: empty array")
	}

	rawCmd := v.Array()[0].String()
	cmdType := strings.ToUpper(rawCmd)

	switch cmdType {
	case proto.CommandCLIENT:
		return parseClientCommand(v)
	case proto.CommandSET:
		return parseSetCommand(v)
	case proto.CommandGET:
		return parseGetCommand(v)
	case proto.CommandHELLO:
		return parseHelloCommand(v)
	case proto.CommandCOMMAND:
		return parseCommandCommand(v)
	case proto.CommandPING:
		return parsePingCommand(v)
	case proto.CommandCONFIG:
		return parseConfigGetCommand(v)
	case proto.CommandEXIST:
		return parseExistCommand(v)
	case proto.CommandDEL:
		return parseDelCommand(v)
	case proto.CommandINCR:
		return parseIncrCommand(v)
	case proto.CommandDECR:
		return parseDecrCommand(v)
	case proto.CommandLPUSH:
		return parseLpushCommand(v)
	default:
		return nil, fmt.Errorf("unsupported command: %s", cmdType)
	}
}

func parseClientCommand(v resp.Value) (proto.ClientCommand, error) {
	cmd := proto.ClientCommand{
		Value: v.Array()[1].String(),
	}

	return cmd, nil
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

func parsePingCommand(v resp.Value) (proto.PingCommand, error) {
	if len(v.Array()) > 2 {
		return proto.PingCommand{}, fmt.Errorf("invalid number of variables for PING command")
	}
	cmd := proto.PingCommand{
		Value: v.Array()[0].String(),
	}

	return cmd, nil
}

func parseExistCommand(v resp.Value) (proto.ExistCommand, error) {
	if len(v.Array()) != 2 {
		return proto.ExistCommand{}, fmt.Errorf("invalid number of variables for GET command")
	}
	cmd := proto.ExistCommand{
		Key: v.Array()[1].String(),
	}

	return cmd, nil
}

func parseConfigGetCommand(v resp.Value) (proto.ConfigGetCommand, error) {
	if len(v.Array()) < 2 {
		return proto.ConfigGetCommand{}, fmt.Errorf("invalid number of variables for CONFIG command")
	}

	var cmd proto.ConfigGetCommand

	if len(v.Array()) == 2 {
		cmd = proto.ConfigGetCommand{
			Key:   v.Array()[1].String(),
			Value: "",
		}
	} else if len(v.Array()) > 2 {
		cmd = proto.ConfigGetCommand{
			Key:   v.Array()[1].String(),
			Value: v.Array()[2].String(),
		}
	}

	return cmd, nil
}

func parseDelCommand(v resp.Value) (proto.DelCommand, error) {
	if len(v.Array()) != 2 {
		return proto.DelCommand{}, fmt.Errorf("invalid number of variables for GET command")
	}
	cmd := proto.DelCommand{
		Key: v.Array()[1].String(),
	}

	return cmd, nil
}

func parseIncrCommand(v resp.Value) (proto.IncrCommand, error) {
	if len(v.Array()) != 2 {
		return proto.IncrCommand{}, fmt.Errorf("invalid number of variables for GET command")
	}
	cmd := proto.IncrCommand{
		Key: v.Array()[1].String(),
	}

	return cmd, nil
}

func parseDecrCommand(v resp.Value) (proto.DecrCommand, error) {
	if len(v.Array()) != 2 {
		return proto.DecrCommand{}, fmt.Errorf("invalid number of variables for GET command")
	}
	cmd := proto.DecrCommand{
		Key: v.Array()[1].String(),
	}

	return cmd, nil
}

func parseLpushCommand(v resp.Value) (proto.LpushCommand, error) {
	if len(v.Array()) < 3 {
		return proto.LpushCommand{}, fmt.Errorf("invalid number of arguments for LPUSH command")
	}

	cmd := proto.LpushCommand{
		Key:   v.Array()[1].String(),
		Value: getElement(v),
	}

	return cmd, nil
}

func getElement(v resp.Value) []string {
	ret := make([]string, 0)
    for _, v := range v.Array()[2:] {
		ret = append(ret, v.String())
	}

	return ret
}
