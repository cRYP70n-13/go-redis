package server

import (
	"fmt"

	"redis-clone/peer"
	"redis-clone/proto"

	"github.com/tidwall/resp"
)

func unhandledCommand(msg peer.Message) error {
	return resp.
		NewWriter(msg.Peer.Conn).
		WriteString("This is not yet handled in our redis")
}

func clientCommandHandler(msg peer.Message) error {
	return resp.
		NewWriter(msg.Peer.Conn).
		WriteString("OK")
}

func commandCommandHandler(msg peer.Message) error {
	spec := map[string]string{
		"server":  "redis",
		"role":    "master",
		"version": "6.0.0",
		"mode":    "standalone",
		"proto":   "3",
		"Author":  "Otmane",
	}
	resMap := proto.WriteRespMap(spec)
	_, err := msg.Peer.Send(resMap)
	if err != nil {
		return fmt.Errorf("error sending response to peer: %s", err)
	}
	return nil
}

func helloCommandHandler(msg peer.Message) error {
	spec := map[string]string{
		"server":  "redis",
		"role":    "master",
		"version": "6.0.0",
		"mode":    "standalone",
		"proto":   "3",
		"Author":  "Otmane",
	}
	resMap := proto.WriteRespMap(spec)
	_, err := msg.Peer.Send(resMap)
	if err != nil {
		return fmt.Errorf("error sending response to peer: %s", err)
	}
	return nil
}

func pingCommandHandler(msg peer.Message) error {
	return resp.NewWriter(msg.Peer.Conn).WriteString("PONG")
}

func getCommandHandler(s *Server, v proto.GetCommand, msg peer.Message) error {
	val, ok := s.Kv.Get(v.Key)
	if !ok {
		return resp.
			NewWriter(msg.Peer.Conn).
			WriteString("Key not found Timmy")
	}

	return resp.
		NewWriter(msg.Peer.Conn).
		WriteString(string(val))
}

func setCommandHandler(s *Server, v proto.SetCommand, msg peer.Message) error {
	if err := s.Kv.Set(v.Key, v.Value); err != nil {
		return resp.
			NewWriter(msg.Peer.Conn).
			WriteString(err.Error())
	}
	// FIXME: We have a bug with our OWN WRITTEN CLIENT here
	// When we send get request to get the value associated with the key
	// we get the OK message which is not fine we have to send the value
	// but with the official redis client this is working fine
	return resp.
		NewWriter(msg.Peer.Conn).
		WriteString("OK")
}

func configCommandGetHandler(msg peer.Message) error {
	return resp.NewWriter(msg.Peer.Conn).WriteArray([]resp.Value{
		resp.StringValue("save"),
		resp.StringValue("3600 1 300 100 60 10000"),
	})
}

func existCommandHandler(s *Server, v proto.ExistCommand, msg peer.Message) error {
	_, ok := s.Kv.Get([]byte(v.Key))
	if !ok {
		return resp.NewWriter(msg.Peer.Conn).WriteInteger(0)
	}

	return resp.NewWriter(msg.Peer.Conn).WriteInteger(1)
}

func delCommandHandler(s *Server, v proto.DelCommand, msg peer.Message) error {
	s.Kv.Del([]byte(v.Key))
	return resp.NewWriter(msg.Peer.Conn).WriteString("OK")
}

func decrCommandHandler(s *Server, v proto.DecrCommand, msg peer.Message) error {
	res, err :=s.Kv.Decr([]byte(v.Key))
	if err != nil {
		return resp.NewWriter(msg.Peer.Conn).WriteError(err)
	}

	return resp.NewWriter(msg.Peer.Conn).WriteInteger(res)
}
	
func incrCommandHandler(s *Server, v proto.IncrCommand, msg peer.Message) error {
	res, err := s.Kv.Incr([]byte(v.Key))
	if err != nil {
		return resp.NewWriter(msg.Peer.Conn).WriteError(err)
	}

	return resp.NewWriter(msg.Peer.Conn).WriteInteger(res)
}