package main

import (
	"log"

	"github.com/PranavKharche24/chat-app/internal/chat"
	"github.com/PranavKharche24/chat-app/internal/storage"
)

func main() {
	// Initialize server-side user database
	db, err := storage.NewDatabase("users.sqlite")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	// Start chat server on TCP port 9000
	srv := chat.NewChatServer(db)
	srv.Start(":9000")
}
