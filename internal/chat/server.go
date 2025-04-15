package chat

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/PranavKharche24/chat-app/internal/encryption"
)

// ChatServer holds client connections and a broadcast channel.
type ChatServer struct {
	clients   map[net.Conn]bool
	mutex     sync.Mutex
	broadcast chan string
}

// NewChatServer creates a new ChatServer instance.
func NewChatServer() *ChatServer {
	return &ChatServer{
		clients:   make(map[net.Conn]bool),
		broadcast: make(chan string),
	}
}

// Start begins listening on the given TCP address.
func (cs *ChatServer) Start(tcpAddr string) {
	listener, err := net.Listen("tcp", tcpAddr)
	if err != nil {
		log.Fatalf("Error starting TCP server: %v", err)
	}
	defer listener.Close()

	// Start broadcasting messages.
	go cs.handleBroadcast()

	log.Printf("TCP Chat server started on %s", tcpAddr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		cs.mutex.Lock()
		cs.clients[conn] = true
		cs.mutex.Unlock()
		go cs.handleClient(conn)
	}
}

// handleBroadcast encrypts messages and sends them to all clients.
func (cs *ChatServer) handleBroadcast() {
	for msg := range cs.broadcast {
		encryptedMsg, err := encryption.Encrypt([]byte(msg))
		if err != nil {
			log.Printf("Encryption error: %v", err)
			continue
		}
		cs.mutex.Lock()
		// Append newline for client scanner separation.
		for conn := range cs.clients {
			_, err := conn.Write([]byte(encryptedMsg + "\n"))
			if err != nil {
				log.Printf("Error writing to client: %v", err)
				conn.Close()
				delete(cs.clients, conn)
			}
		}
		cs.mutex.Unlock()
	}
}

// handleClient reads messages from a TCP client and forwards them for broadcast.
func (cs *ChatServer) handleClient(conn net.Conn) {
	defer func() {
		cs.mutex.Lock()
		delete(cs.clients, conn)
		cs.mutex.Unlock()
		conn.Close()
	}()

	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Client disconnected: %v", err)
			break
		}
		fmt.Printf("Received message: %s", msg)
		cs.broadcast <- msg
	}
}

