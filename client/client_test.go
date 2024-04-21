package client

import (
	// "bytes"
	"context"
	"fmt"
	"log"
	"sync"
	"testing"

	"github.com/redis/go-redis/v9"
	// "github.com/tidwall/resp"
)

func TestNewClientRedisClient(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr:       "localhost:5001",
		Password:   "",
		DB:         0,
		MaxRetries: 2,
	})
	fmt.Println(rdb)
	fmt.Println("This shit is working")

	if err := rdb.Set(context.Background(), "otmane", "kimdil", 0).Err(); err != nil {
		panic(err)
	}

	res := rdb.Get(context.Background(), "otmane")
	fmt.Println("Server's response for the GET => ", res.Val())
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
