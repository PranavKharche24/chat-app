package main

import (
	"log"

	"github.com/PranavKharche24/chat-app/internal/chat"
	"github.com/PranavKharche24/chat-app/internal/storage"
)

func main() {
	db, err := storage.NewDatabase("users.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	srv := chat.NewChatServer(db)
	log.Fatal(srv.Start(":9000"))
}
