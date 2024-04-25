package server

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestWithRedisGoClient(t *testing.T) {
	listenAddr := ":5001"
	s := NewServer(Config{
		ListenAddress: listenAddr,
	})
	go func() {
		log.Fatal(s.Start())
	}()
	time.Sleep(time.Millisecond * 400)

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("localhost%s", ":5001"),
		Password: "",
		DB:       0,
	})

	testCases := map[string]string{
		"foo":    "bar",
		"redis":  "server",
		"GoLang": "is_the_best",
	}
	for key, val := range testCases {
		if err := rdb.Set(context.Background(), key, val, 0).Err(); err != nil {
			t.Fatal(err)
		}
		newVal, err := rdb.Get(context.Background(), key).Result()
		if err != nil {
			t.Fatal(err)
		}
        fmt.Println(newVal)
        if newVal != val {
			t.Fatalf("expected %s but got %s", val, newVal)
		}
	}
}
