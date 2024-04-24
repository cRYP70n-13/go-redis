package client

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/redis/go-redis/v9"
)

func TestNewClientRedisClient(t *testing.T) {
	t.Skip()
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
	t.Skip()
	client, err := New("localhost:5001")
	if err != nil {
		t.Error(err)
	}
	defer client.Close()

	if err := client.Set(context.Background(), "client_foo_1", "client_bar_1"); err != nil {
		log.Fatal(err)
	}

	val, err := client.Get(context.Background(), "client_foo_1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Got this: ", val)
}
