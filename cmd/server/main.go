package main

import (
	"log"
	"net/http"
	"sync"

	"github.com/PranavKharche24/chat-app/internal/chat"
)

func main() {
	var wg sync.WaitGroup
	wg.Add(2)

	// Start the TCP chat server.
	go func() {
		defer wg.Done()
		cs := chat.NewChatServer()
		cs.Start(":9000")
	}()

	// Start the HTTP API server.
	go func() {
		defer wg.Done()
		chatAPI := chat.NewChatAPI()
		http.HandleFunc("/api/send", chatAPI.SendMessageHandler)
		http.HandleFunc("/api/messages", chatAPI.GetMessagesHandler)

		// Serve the basic UI.
		http.Handle("/", http.FileServer(http.Dir("./web")))
		log.Println("HTTP API server started on :8080")
		log.Fatal(http.ListenAndServe(":8080", nil))
	}()

	wg.Wait()
}

