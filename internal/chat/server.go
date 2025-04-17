package chat

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/PranavKharche24/chat-app/internal/encryption"
	"github.com/PranavKharche24/chat-app/internal/storage"
)

type Client struct {
	conn   net.Conn
	userID int64
	out    chan string
}

type ChatServer struct {
	db      *storage.Database
	clients map[int64]*Client
	mutex   sync.RWMutex
}

func NewChatServer(db *storage.Database) *ChatServer {
	return &ChatServer{
		db:      db,
		clients: make(map[int64]*Client),
	}
}

func (s *ChatServer) Start(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	log.Printf("Listening on %s …", addr)
	for {
		conn, _ := ln.Accept()
		go s.handleConn(conn)
	}
}

func (s *ChatServer) handleConn(conn net.Conn) {
	c := &Client{conn: conn, out: make(chan string, 10)}
	defer conn.Close()
	go func() {
		for msg := range c.out {
			conn.Write([]byte(msg + "\n"))
		}
	}()

	// Welcome & commands
	c.out <- "🌟 Welcome to Super‑Chat! 🌟"
	c.out <- "Commands:"
	c.out <- "  REGISTER username email dob(YYYY-MM-DD) fullname password"
	c.out <- "  LOGIN userID password"
	c.out <- "  SEND username message…"
	c.out <- "  QUIT"

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, " ", 2)
		cmd := strings.ToUpper(parts[0])
		arg := ""
		if len(parts) > 1 {
			arg = parts[1]
		}

		switch cmd {
		case "REGISTER":
			f := strings.SplitN(arg, " ", 5)
			if len(f) < 5 {
				c.out <- "Usage: REGISTER username email dob fullname password"
				continue
			}
			uid, err := s.db.RegisterUser(f[0], f[1], f[2], f[3], f[4])
			if err != nil {
				c.out <- "ERROR: " + err.Error()
			} else {
				c.out <- fmt.Sprintf("REGISTERED userID=%d", uid)
			}

		case "LOGIN":
			f := strings.SplitN(arg, " ", 2)
			if len(f) < 2 {
				c.out <- "Usage: LOGIN userID password"
				continue
			}
			uid, err := strconv.ParseInt(f[0], 10, 64)
			if err != nil {
				c.out <- "ERROR: invalid userID"
				continue
			}
			if err := s.db.Authenticate(uid, f[1]); err != nil {
				c.out <- "ERROR: " + err.Error()
			} else {
				c.userID = uid
				s.mutex.Lock()
				s.clients[uid] = c
				s.mutex.Unlock()
				c.out <- "LOGIN OK"
			}

		case "SEND":
			if c.userID == 0 {
				c.out <- "ERROR: please LOGIN first"
				continue
			}
			f := strings.SplitN(arg, " ", 2)
			if len(f) < 2 {
				c.out <- "Usage: SEND username message"
				continue
			}
			toID, err := s.db.LookupByUsername(f[0])
			if err != nil {
				c.out <- "ERROR: user not found"
				continue
			}
			s.mutex.RLock()
			dest, online := s.clients[toID]
			s.mutex.RUnlock()
			// Format message with sender’s username
			sender, _ := s.db.LookupUsernameByID(c.userID)
			full := fmt.Sprintf("[%s] %s", sender, f[1])
			enc, _ := encryption.Encrypt([]byte(full))
			if online {
				dest.out <- "CHAT:" + enc
			}
			c.out <- "SENT"

		case "QUIT":
			c.out <- "BYE"
			return

		default:
			c.out <- "Unknown command"
		}
	}
}
