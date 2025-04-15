package main

import (
	"log"

	"github.com/PranavKharche24/chat-app/internal/chat"
	"github.com/PranavKharche24/chat-app/internal/storage"
)

func main() {
	// Initialize database file ("chatapp.db" in the current directory)
	db, err := storage.NewDatabase("chatapp.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	server := chat.NewChatServer(db)
	// Start the TCP server on port 9000.
	server.Start(":9000")
}
