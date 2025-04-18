package chat

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	"github.com/PranavKharche24/chat-app/internal/chat/groups"
	"github.com/PranavKharche24/chat-app/internal/encryption"
	"github.com/PranavKharche24/chat-app/internal/storage"
)

type Client struct {
	conn   net.Conn
	userID string
	out    chan string
}

type ChatServer struct {
	db      *storage.Database
	gm      *groups.Manager
	clients map[string]*Client
	mutex   sync.RWMutex
}

func NewChatServer(db *storage.Database) *ChatServer {
	return &ChatServer{
		db:      db,
		gm:      groups.NewManager(),
		clients: make(map[string]*Client),
	}
}

func (s *ChatServer) Start(addr string) {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen error: %v", err)
	}
	log.Printf("Server listening on %s", addr)
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

	// Command list
	c.out <- "✨ Super‑Chat Server ✨"
	c.out <- "REGISTER username email dob full_name password"
	c.out <- "LOGIN userID password"
	c.out <- "SEND username message"
	c.out <- "SENDALL message"
	c.out <- "CREATEGROUP name"
	c.out <- "JOINGROUP name"
	c.out <- "LEAVEGROUP name"
	c.out <- "SENDGROUP name message"
	c.out <- "SETTINGS"
	c.out <- "QUIT"

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
				c.out <- "Usage: REGISTER username email dob full_name password"
				continue
			}
			id, err := s.db.RegisterUser(f[0], f[1], f[2], f[3], f[4])
			if err != nil {
				c.out <- "ERROR: " + err.Error()
			} else {
				c.out <- "REGISTERED:" + id
			}

		case "LOGIN":
			f := strings.SplitN(arg, " ", 2)
			if len(f) < 2 {
				c.out <- "Usage: LOGIN userID password"
				continue
			}
			if err := s.db.Authenticate(f[0], f[1]); err != nil {
				c.out <- "ERROR: " + err.Error()
				continue
			}
			c.userID = f[0]
			s.mutex.Lock()
			s.clients[c.userID] = c
			s.mutex.Unlock()
			c.out <- "LOGIN OK"

		case "SEND":
			if c.userID == "" {
				c.out <- "ERROR: login first"
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
			full := fmt.Sprintf("[%s] %s", c.userID, f[1])
			enc, _ := encryption.Encrypt([]byte(full))
			s.mutex.RLock()
			if dest, ok := s.clients[toID]; ok {
				dest.out <- "CHAT:" + enc
			}
			s.mutex.RUnlock()
			c.out <- "SENT"

		case "SENDALL":
			if c.userID == "" {
				c.out <- "ERROR: login first"
				continue
			}
			full := fmt.Sprintf("[ALL][%s] %s", c.userID, arg)
			enc, _ := encryption.Encrypt([]byte(full))
			s.mutex.RLock()
			for id, client := range s.clients {
				if id != c.userID {
					client.out <- "CHAT:" + enc
				}
			}
			s.mutex.RUnlock()
			c.out <- "SENT"

		case "CREATEGROUP":
			if err := s.gm.CreateGroup(arg); err != nil {
				c.out <- "ERROR: " + err.Error()
			} else {
				c.out <- "GROUPCREATED:" + arg
			}

		case "JOINGROUP":
			if err := s.gm.AddMember(arg, c.userID); err != nil {
				c.out <- "ERROR: " + err.Error()
			} else {
				c.out <- "JOINEDGROUP:" + arg
			}

		case "LEAVEGROUP":
			if err := s.gm.RemoveMember(arg, c.userID); err != nil {
				c.out <- "ERROR: " + err.Error()
			} else {
				c.out <- "LEFTGROUP:" + arg
			}

		case "SENDGROUP":
			f := strings.SplitN(arg, " ", 2)
			if len(f) < 2 {
				c.out <- "Usage: SENDGROUP name message"
				continue
			}
			members, err := s.gm.Members(f[0])
			if err != nil {
				c.out <- "ERROR: " + err.Error()
				continue
			}
			full := fmt.Sprintf("[Group:%s][%s] %s", f[0], c.userID, f[1])
			enc, _ := encryption.Encrypt([]byte(full))
			s.mutex.RLock()
			for _, uid := range members {
				if dest, ok := s.clients[uid]; ok && uid != c.userID {
					dest.out <- "CHAT:" + enc
				}
			}
			s.mutex.RUnlock()
			c.out <- "SENTGROUP"

		case "SETTINGS":
			u, e, d, fn, err := s.db.GetProfile(c.userID)
			if err != nil {
				c.out <- "ERROR: " + err.Error()
			} else {
				c.out <- fmt.Sprintf("PROFILE:%s:%s:%s:%s", u, e, d, fn)
			}

		case "QUIT":
			c.out <- "BYE"
			return

		default:
			c.out <- "Unknown command"
		}
	}
}
