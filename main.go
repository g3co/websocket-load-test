package main

import (
	"context"
	"github.com/gorilla/websocket"
	"log"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"
)

var count int64

func main()  {
	ctx, cancel := context.WithCancel(context.Background())

	attempts := 0
	go func() {
		buf := make(chan struct{}, 900)
		for {
			buf <- struct{}{}
			attempts++
			go connection(ctx, buf)
		}
	}()

	go func() {
		tick := time.Tick(time.Second*5)
		prevAttempts := 0
		for {
			<-tick
			if attempts != prevAttempts {
				log.Printf("total connection attempts %d", attempts)
				prevAttempts = attempts
			}
		}
	}()

	go func() {
		tick := time.Tick(time.Second*5)

		for {
			select {
			case <-tick:
				log.Printf("Handled %d messages", count)
				count = 0
			case <-ctx.Done():
				return
			}
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	log.Printf("terminated with sig %q", <-sigChan)
	cancel()
}

func connection(ctx context.Context, buf <-chan struct{})  {
	defer func() {
		<-buf
	}()

	c, _, err := websocket.DefaultDialer.Dial("wss://rw-manager.rock-west.net/ticks_feed", nil)
	if err != nil {
		log.Println("dial:", err)
		return
	}

	defer func() {
		c.Close()
		c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, _, err := c.ReadMessage()
			atomic.AddInt64(&count, 1)
			if err != nil {
				log.Println("read:", err)
				return
			}
		}
	}
}
