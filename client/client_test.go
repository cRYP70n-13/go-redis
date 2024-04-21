package client

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
)

func init() {
	// TODO: Decouple the server from the main package make a separated one
	// then use it here to test this $hit
}

func TestNewClients(t *testing.T) {
	nClients := 10
	wg := sync.WaitGroup{}
	wg.Add(nClients)
	for i := 0; i < nClients; i++ {
		go func(iterator int) {
			client, err := New("localhost:5001")
			if err != nil {
				t.Error(err)
			}
			defer client.Close()

			key := fmt.Sprintf("client_foo_%d", iterator)
			value := fmt.Sprintf("client_bar_%d", iterator)
			if err := client.Set(context.Background(), key, value); err != nil {
				log.Fatal(err)
			}

			val, err := client.Get(context.Background(), key)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("client number %d, sent RPC GET -> %s and got this value back => %s\n", iterator, key, val)

			wg.Done()
		}(i)
	}
	wg.Wait()
}

