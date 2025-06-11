package main

import (
	"encoding/binary"
	"errors"
	"log"
	"net"
	"os"
	"time"
)

type Client net.Conn

type ClientManager struct {
	Clients    map[Client]bool
	Register   chan Client
	UnRegister chan Client
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		Clients:    make(map[Client]bool),
		Register:   make(chan Client),
		UnRegister: make(chan Client),
	}
}

func (cm *ClientManager) Run(frames chan []byte) {
	for {
		select {
		case client := <-cm.Register:
			cm.Clients[client] = true
		case client := <-cm.UnRegister:
			delete(cm.Clients, client)
		case data := <-frames:
			t1 := time.Now()
			for client := range cm.Clients {
				client.SetWriteDeadline(time.Now().Add(time.Millisecond * 10))

				if err := binary.Write(client, binary.BigEndian, uint32(len(data))); err != nil {
					if errors.Is(err, os.ErrDeadlineExceeded) {
						log.Printf("%s took too long to write frame length\n", client.RemoteAddr().String())
						continue
					}

					client.Close()
					delete(cm.Clients, client)
				}

				_, err := client.Write(data)

				if err != nil {
					if errors.Is(err, os.ErrDeadlineExceeded) {
						log.Printf("%s took too long to write frame data\n", client.RemoteAddr().String())
						continue
					}

					client.Close()
					delete(cm.Clients, client)
				}

			}

			tt1 := time.Since(t1)

			if tt1 > time.Millisecond {
				log.Printf("ALL %s\n", tt1)
			}
		}
	}
}
