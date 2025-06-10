package main

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"io"
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

	// each split has a seperate tcp server for stream,
	// will be one server in the future with commands to change stream
	for i, split := range splits {
		go createSplitServer(fmt.Sprintf(":%d", 3000+i), split.Output)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
}

func createSplitServer(address string, frames chan []byte) {
	cm := NewClientManager()

	go cm.Run(frames)

	l, err := net.Listen("tcp", address)

	if err != nil {
		log.Fatal(err)
	}

	defer l.Close()

	for {
		c, err := l.Accept()

		if err != nil {
			fmt.Println(err)
			return
		}

		go handleConnection(c, cm)
	}
}

func handleConnection(c net.Conn, cm *ClientManager) {
	cm.Register <- c

	defer func() {
		cm.UnRegister <- c
	}()

	for {
		var cmd uint8

		if err := binary.Read(c, binary.BigEndian, &cmd); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, net.ErrClosed) {
				break
			}

			log.Printf("error reading command from %s: %v\n", c.RemoteAddr(), err)
			continue
		}

		log.Printf("received command %d from %s", cmd, c.RemoteAddr())
	}
}
