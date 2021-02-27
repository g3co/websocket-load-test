package main

import (
	"context"
	"github.com/gorilla/websocket"
	"log"
	"os"
	"os/signal"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"
)

var count int64

func main() {
	args := os.Args[1:]
	if len(args) != 2 {
		log.Fatal("Mismatch arguments quantity: websocket-load-test [ws address] [threads num]")
	}

	threads, err := strconv.Atoi(args[1])
	if err != nil {
		log.Fatal("can't parse threads num", err)
	}

	log.Printf("connecting to the %s", args[0])

	ctx, cancel := context.WithCancel(context.Background())

	attempts := 0
	go func() {
		buf := make(chan struct{}, threads)
		for {
			buf <- struct{}{}
			attempts++
			go connection(ctx, args[0], buf)
		}
	}()

	go func() {
		tick := time.Tick(time.Second * 1)
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
		tick := time.Tick(time.Second * 30)

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

func connection(ctx context.Context, host string, buf <-chan struct{}) {
	defer func() {
		<-buf
	}()

	c, _, err := websocket.DefaultDialer.Dial(host, nil)
	if err != nil {
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
