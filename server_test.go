package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"redis-clone/client"
)

func TestClientHello(t *testing.T) {
	in := map[string]string{
		"server":  "redis",
		"version": "6.0.0",
	}
	out := writeRespMap(in)
	fmt.Println(string(out))
}

// TODO: fix this potato server to correctly communicate when it's done and when it's still up and running.
// to avoid the time.Sleeps
func TestServerWithClients(t *testing.T) {
	server := NewServer(Config{})
	go func() {
		log.Fatal(server.Start())
	}()
	// FIXME: This is more of a hack we need to sync this with a waitgroup or something
	time.Sleep(time.Second)

	nClients := 10
	wg := sync.WaitGroup{}
	wg.Add(nClients)
	for i := 0; i < nClients; i++ {
		go func(iterator int) {
			client, err := client.New("localhost:5001")
			if err != nil {
				log.Fatal(err)
			}
			defer client.Close()

			key := fmt.Sprintf("client_foo_%d", iterator)
			// val := fmt.Sprintf("client_bar_%d", iterator)
			if err := client.Set(context.Background(), key, struct {
				fistName string
				lastName string
			}{fistName: "Otmane", lastName: "Kimdil"}); err != nil {
				log.Fatal(err)
			}

			value, err := client.Get(context.Background(), key)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("Client number %d, made a RPC to get %s and got value => %s\n", iterator, key, value)
			wg.Done()
		}(i)
	}
	wg.Wait()

	// FIXME: Same here :)
	time.Sleep(time.Second)
	if len(server.Peers) != 0 {
		t.Fatalf("expected 0 peers but got %d", len(server.Peers))
	}
}
