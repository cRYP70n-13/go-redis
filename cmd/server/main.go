package main

import (
	"log"

	"redis-clone/server"
)

func main() {
	s := server.NewServer(server.Config{})
	log.Fatal(s.Start())
}
