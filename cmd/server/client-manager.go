package main

import (
	"encoding/binary"
	"net"
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
			for client := range cm.Clients {
				if err := binary.Write(client, binary.BigEndian, uint32(len(data))); err != nil {
					client.Close()
					delete(cm.Clients, client)
				}

				_, err := client.Write(data)

				if err != nil {
					client.Close()
					delete(cm.Clients, client)
				}
			}
		}
	}
}
