package chat

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/PranavKharche24/chat-app/internal/encryption"
	"github.com/PranavKharche24/chat-app/internal/storage"
)

// Client holds data for a connected client.
type Client struct {
	conn     net.Conn
	username string
	out      chan string
}

// ChatServer holds the storage and connected clients.
type ChatServer struct {
	storage *storage.Database
	clients map[net.Conn]*Client
	mutex   sync.Mutex
}

// NewChatServer creates a new ChatServer with an initialized storage backend.
func NewChatServer(db *storage.Database) *ChatServer {
	return &ChatServer{
		storage: db,
		clients: make(map[net.Conn]*Client),
	}
}

// Start begins listening on a TCP address and accepts incoming clients.
func (cs *ChatServer) Start(addr string) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("Error starting TCP server: %v", err)
	}
	defer listener.Close()
	log.Printf("Chat server listening on %s", addr)
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		client := &Client{
			conn: conn,
			out:  make(chan string),
		}
		cs.mutex.Lock()
		cs.clients[conn] = client
		cs.mutex.Unlock()
		go cs.handleClient(client)
		go cs.writeToClient(client)
	}
}

// broadcastMessage encrypts and sends a chat message to all clients.
func (cs *ChatServer) broadcastMessage(sender string, message string) {
	fullMsg := fmt.Sprintf("[%s]: %s", sender, message)
	encrypted, err := encryption.Encrypt([]byte(fullMsg))
	if err != nil {
		log.Printf("Encryption error: %v", err)
		return
	}
	broadcast := "CHAT: " + encrypted
	cs.mutex.Lock()
	defer cs.mutex.Unlock()
	for _, client := range cs.clients {
		client.out <- broadcast
	}
}

// handleClient processes commands from a single client.
func (cs *ChatServer) handleClient(client *Client) {
	defer func() {
		cs.mutex.Lock()
		delete(cs.clients, client.conn)
		cs.mutex.Unlock()
		client.conn.Close()
	}()
	// Send a welcome message.
	client.out <- "Welcome to ChatApp CLI.\nAvailable commands:\n  REGISTER username password\n  LOGIN username password\n  MSG message\n  HISTORY\n  QUIT"
	scanner := bufio.NewScanner(client.conn)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 2)
		command := strings.ToUpper(strings.TrimSpace(parts[0]))
		var args string
		if len(parts) > 1 {
			args = strings.TrimSpace(parts[1])
		}
		switch command {
		case "REGISTER":
			// Expect: REGISTER username password
			argsParts := strings.SplitN(args, " ", 2)
			if len(argsParts) < 2 {
				client.out <- "Usage: REGISTER username password"
				continue
			}
			username := strings.TrimSpace(argsParts[0])
			password := strings.TrimSpace(argsParts[1])
			if err := cs.storage.RegisterUser(username, password); err != nil {
				client.out <- fmt.Sprintf("Registration failed: %v", err)
			} else {
				client.out <- "Registration successful. Please LOGIN with your credentials."
			}
		case "LOGIN":
			// Expect: LOGIN username password
			argsParts := strings.SplitN(args, " ", 2)
			if len(argsParts) < 2 {
				client.out <- "Usage: LOGIN username password"
				continue
			}
			username := strings.TrimSpace(argsParts[0])
			password := strings.TrimSpace(argsParts[1])
			if err := cs.storage.AuthenticateUser(username, password); err != nil {
				client.out <- fmt.Sprintf("Login failed: %v", err)
			} else {
				client.username = username
				client.out <- "Login successful. You can now send messages using the MSG command."
			}
		case "MSG":
			if client.username == "" {
				client.out <- "You must login first."
				continue
			}
			if args == "" {
				client.out <- "Usage: MSG your message here"
				continue
			}
			// Save message to DB.
			if err := cs.storage.StoreMessage(client.username, args); err != nil {
				client.out <- fmt.Sprintf("Failed to store message: %v", err)
			}
			cs.broadcastMessage(client.username, args)
		case "HISTORY":
			messages, err := cs.storage.GetMessages(10)
			if err != nil {
				client.out <- fmt.Sprintf("Failed to fetch history: %v", err)
			} else {
				// Display from oldest to newest.
				for i := len(messages) - 1; i >= 0; i-- {
					msg := messages[i]
					client.out <- fmt.Sprintf("[%s @ %s]: %s", msg.Username, msg.Timestamp, msg.Content)
				}
			}
		case "QUIT":
			client.out <- "Goodbye!"
			return
		default:
			client.out <- "Unknown command. Options: REGISTER, LOGIN, MSG, HISTORY, QUIT"
		}
	}
}

// writeToClient continuously writes queued messages to the client connection.
func (cs *ChatServer) writeToClient(client *Client) {
	writer := bufio.NewWriter(client.conn)
	for msg := range client.out {
		_, err := writer.WriteString(msg + "\n")
		if err != nil {
			log.Printf("Error writing to client: %v", err)
			return
		}
		writer.Flush()
	}
}
