package main

import (
	"context"
	"log"
	"os"
	"rush/ws"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	authStatusCh := make(chan ws.AuthStatus)
	key := os.Getenv("RUSH_LICENSE_KEY")
	if key == "" {
		key = "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"
	}
	go ws.AuthPoll(key, authStatusCh, ctx)
	for {
		select {
		case authStatus := <-authStatusCh:
			log.Printf("%+v", authStatus)
			// cancel()
		}
	}
}
