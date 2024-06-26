package main

import (
	"log"

	"redis-clone/server"
)

func main() {
	s := server.NewServer(server.Config{
        ListenAddress: "localhost:6379",
    })
	log.Fatal(s.Start())
}
