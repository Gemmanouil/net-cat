package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	defaultPort = "8989"
	maxClients  = 10
	timeFormat  = "2006-01-02 15:04:05"
	usage       = "[USAGE]: ./TCPChat $port"
)

// The logo contains backticks, so those two lines are spliced in
// with interpreted strings.
const welcome = `Welcome to TCP-Chat!
         _nnnn_
        dGGGGMMb
       @p~qp~~qMb
       M|@||@) M|
       @,----.JM|
      JS^\__/  qKL
     dZP        qKRb
    dZP          qKKb
   fZP            SMMb
   HZM            MMMM
   FqM            MMMM
 __| ".        |\dS"qML
 |    ` + "`" + `.       | ` + "`" + `' \Zq
_)      \.___.,|     .'
\____   )MMMMMP|   .'
     ` + "`" + `-'       ` + "`" + `--'
`

var errNameTaken = errors.New("That name is already taken, choose another one.")

type client struct {
	name string
	conn net.Conn
}

// server holds all shared state. Every field below mu must only be
// touched while holding mu.
type server struct {
	mu      sync.Mutex
	active  int // connections in use (including ones still choosing a name)
	clients map[net.Conn]*client
	history []string
}

func newServer() *server {
	return &server{clients: make(map[net.Conn]*client)}
}

// acquire reserves one of the maxClients slots.
func (s *server) acquire() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.active >= maxClients {
		return false
	}
	s.active++
	return true
}

func (s *server) release() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.active--
}

// prompt prints "[time][name]:" for one client. Caller must hold s.mu.
func prompt(c *client) {
	fmt.Fprintf(c.conn, "[%s][%s]:", time.Now().Format(timeFormat), c.name)
}

// broadcastLocked sends text to everyone except from, then re-prompts them.
// Caller must hold s.mu. Write errors are ignored on purpose: a dead
// connection is cleaned up by its own read loop.
func (s *server) broadcastLocked(from *client, text string) {
	for _, other := range s.clients {
		if other == from {
			continue
		}
		fmt.Fprint(other.conn, "\n"+text)
		prompt(other)
	}
}

// join registers a client, replays history to it and announces it.
func (s *server) join(c *client) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, other := range s.clients {
		if other.name == c.name {
			return errNameTaken
		}
	}
	for _, line := range s.history {
		fmt.Fprint(c.conn, line)
	}
	s.clients[c.conn] = c
	s.broadcastLocked(c, c.name+" has joined the chat...\n")
	prompt(c)
	return nil
}

func (s *server) leave(c *client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, c.conn)
	s.broadcastLocked(c, c.name+" has left the chat...\n")
}

// post stores and broadcasts one chat message.
func (s *server) post(c *client, msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	line := fmt.Sprintf("[%s][%s]:%s\n", time.Now().Format(timeFormat), c.name, msg)
	s.history = append(s.history, line)
	s.broadcastLocked(c, line)
	prompt(c)
}

func (s *server) reprompt(c *client) {
	s.mu.Lock()
	defer s.mu.Unlock()
	prompt(c)
}

// readName keeps asking until the client sends a non-empty name.
func readName(conn net.Conn, r *bufio.Reader) (string, error) {
	for {
		fmt.Fprint(conn, "[ENTER YOUR NAME]: ")
		line, err := r.ReadString('\n')
		if err != nil {
			return "", err
		}
		name := strings.TrimSpace(line)
		if name != "" {
			return name, nil
		}
		fmt.Fprintln(conn, "Name cannot be empty.")
	}
}

func (s *server) handle(conn net.Conn) {
	defer conn.Close()
	defer s.release()

	fmt.Fprint(conn, welcome)
	reader := bufio.NewReader(conn)

	var c *client
	for c == nil {
		name, err := readName(conn, reader)
		if err != nil {
			return // disconnected before choosing a name
		}
		candidate := &client{name: name, conn: conn}
		if err := s.join(candidate); err != nil {
			fmt.Fprintln(conn, err)
			continue
		}
		c = candidate
	}
	defer s.leave(c)
	log.Printf("%s joined (%s)", c.name, conn.RemoteAddr())

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("%s left", c.name)
			return
		}
		msg := strings.TrimSpace(line)
		if msg == "" {
			s.reprompt(c) // never broadcast empty messages
			continue
		}
		s.post(c, msg)
	}
}

func main() {
	port := defaultPort
	switch len(os.Args) {
	case 1:
	case 2:
		port = os.Args[1]
	default:
		fmt.Println(usage)
		return
	}

	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("cannot listen on port %s: %v", port, err)
	}
	defer listener.Close()
	fmt.Println("Listening on the port :" + port)

	srv := newServer()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept error: %v", err)
			continue
		}
		if !srv.acquire() {
			fmt.Fprintln(conn, "Chat is full (max 10 connections), try again later.")
			conn.Close()
			continue
		}
		go srv.handle(conn)
	}
}
