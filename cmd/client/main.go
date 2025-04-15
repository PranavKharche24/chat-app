package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/PranavKharche24/chat-app/internal/encryption"
)

func main() {
	// Connect to the TCP chat server.
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		log.Fatalf("Error connecting to server: %v", err)
	}
	defer conn.Close()

	// Goroutine to listen for messages from the server.
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			encryptedMsg := scanner.Text()
			decryptedMsg, err := encryption.Decrypt(encryptedMsg)
			if err != nil {
				fmt.Println("Error decrypting message:", err)
			} else {
				fmt.Println("Message:", decryptedMsg)
			}
		}
	}()

	// Read input from the terminal and send it to the server.
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Enter message: ")
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Error reading input: %v", err)
		}
		_, err = fmt.Fprintf(conn, text)
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}

