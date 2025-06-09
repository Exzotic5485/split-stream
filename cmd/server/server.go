package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"image"
	"log"
	"net"
	"os/signal"
	"syscall"

	"github.com/exzotic5485/split-stream/splitstream"
)

func main() {
	var splits []splitstream.Split = []splitstream.Split{
		{
			Rect:   image.Rect(0, 0, 958, 538),
			Output: make(chan []byte),
		},
		{
			Rect:   image.Rect(960, 0, 1920, 538),
			Output: make(chan []byte),
		},
		{
			Rect:   image.Rect(0, 545, 958, 1080),
			Output: make(chan []byte),
		},
		// {
		// 	Rect:   image.Rect(960, 545, 1920, 1080),
		// 	Output: make(chan []byte),
		// },
	}

	ss := splitstream.NewSplitStream("/dev/video0", splits)

	go ss.Run()
	go createSplitServer(":3000", splits[0].Output)
	go createSplitServer(":3001", splits[1].Output)
	go createSplitServer(":3002", splits[2].Output)
	// go createSplitServer(":3003", splits[3].Output)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
}

func createSplitServer(address string, frames chan []byte) {
	l, err := net.Listen("tcp", address)

	if err != nil {
		log.Fatal(err)
	}

	defer l.Close()

	var clients map[net.Conn]bool = map[net.Conn]bool{}

	go func() {
		for {
			frame := <-frames

			for c := range clients {
				if err := binary.Write(c, binary.BigEndian, uint32(len(frame))); err != nil {
					c.Close()
					delete(clients, c)
				}

				_, err := c.Write(frame)

				if err != nil {
					c.Close()
					delete(clients, c)
				}
			}
		}
	}()

	for {
		c, err := l.Accept()

		if err != nil {
			fmt.Println(err)
			return
		}

		clients[c] = true
	}
}
