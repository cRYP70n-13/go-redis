package client

import (
	"context"
	"fmt"
	"log"
	"testing"
)

func init() {
	// TODO: Decouple the server from the main package make a separated one
	// then use it here to test this $hit
}

// TODO: With this approach we basically gonna break our CI because the server is not running
// Will need to run the server here in the same test and then run the client, looks like a good
// candidate for init function.
func TestNewClient(t *testing.T) {
	t.Skip(t)
	client, err := New("localhost:5001")
	if err != nil {
		t.Error(err)
	}

	for i := 0; i < 20; i++ {
		if err := client.Set(context.Background(), fmt.Sprintf("Otmane_%d", i), fmt.Sprintf("Kimdil_%d", i)); err != nil {
			log.Fatal(err)
		}

		value, err := client.Get(context.Background(), fmt.Sprintf("Otmane_%d", i))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("GET =>", value)
	}
}
