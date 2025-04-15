package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/PranavKharche24/chat-app/internal/encryption"
)

const (
	ANSI_RESET = "\033[0m"
	ANSI_GREEN = "\033[32m"
	ANSI_CYAN  = "\033[36m"
	ANSI_RED   = "\033[31m"
)

func main() {
	// By default, connect to localhost:9000; override by passing an address as the first argument.
	serverAddr := "localhost:9000"
	if len(os.Args) > 1 {
		serverAddr = os.Args[1]
	}
	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		log.Fatalf("Could not connect to server: %v", err)
	}
	defer conn.Close()
	fmt.Println(ANSI_GREEN + "Connected to ChatApp server at " + serverAddr + ANSI_RESET)

	// Goroutine to listen for messages from the server.
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "CHAT: ") {
				// Remove the prefix and decrypt the chat message.
				encryptedMsg := strings.TrimPrefix(line, "CHAT: ")
				decrypted, err := encryption.Decrypt(encryptedMsg)
				if err != nil {
					fmt.Println(ANSI_RED + "Failed to decrypt message" + ANSI_RESET)
				} else {
					fmt.Println(ANSI_CYAN + decrypted + ANSI_RESET)
				}
			} else {
				// Regular protocol messages (welcome, errors, prompts, etc.)
				fmt.Println(line)
			}
		}
	}()

	// Read user input from the terminal and forward it to the server.
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">> ")
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("Error reading from stdin: %v", err)
		}
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		_, err = fmt.Fprintf(conn, "%s\n", text)
		if err != nil {
			log.Printf("Error sending message: %v", err)
		}
	}
}
