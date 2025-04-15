package chat

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/PranavKharche24/chat-app/internal/encryption"
)

// Message represents a chat message.
type Message struct {
	Content string `json:"content"`
}

// ChatAPI stores messages and defines HTTP handlers.
type ChatAPI struct {
	messages []string
	mutex    sync.Mutex
}

// NewChatAPI returns a new ChatAPI instance.
func NewChatAPI() *ChatAPI {
	return &ChatAPI{
		messages: make([]string, 0),
	}
}

// SendMessageHandler encrypts and stores a received message.
func (api *ChatAPI) SendMessageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Unable to read body", http.StatusBadRequest)
		return
	}
	var msg Message
	if err := json.Unmarshal(body, &msg); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	encrypted, err := encryption.Encrypt([]byte(msg.Content))
	if err != nil {
		http.Error(w, "Encryption error", http.StatusInternalServerError)
		return
	}
	api.mutex.Lock()
	api.messages = append(api.messages, encrypted)
	api.mutex.Unlock()
	log.Printf("Message received via API: %s", msg.Content)
	w.WriteHeader(http.StatusOK)
}

// GetMessagesHandler decrypts and returns stored messages.
func (api *ChatAPI) GetMessagesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}
	api.mutex.Lock()
	defer api.mutex.Unlock()
	var decryptedMessages []string
	for _, encryptedMsg := range api.messages {
		decrypted, err := encryption.Decrypt(encryptedMsg)
		if err != nil {
			decryptedMessages = append(decryptedMessages, "DECRYPTION ERROR")
		} else {
			decryptedMessages = append(decryptedMessages, decrypted)
		}
	}
	jsonData, err := json.Marshal(decryptedMessages)
	if err != nil {
		http.Error(w, "Failed to marshal messages", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

