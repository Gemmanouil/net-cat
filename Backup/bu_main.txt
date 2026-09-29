package main // define main package

import (
	"bufio"   // buffered I/O
	"fmt"     // formatted printing
	"net"     // networking (TCP)
	"os"      // OS args
	"strings" // string helpers
	"time"    // timestamps
)

// Client holds one user's name and connection
type Client struct {
	Name string   // username
	Conn net.Conn // TCP connection
}

// global slice of all connected clients
var clients []Client

// broadcast sends a message to every connected client
func broadcast(message string) {
	for _, client := range clients { // loop over all clients
		client.Conn.Write([]byte(message)) // write message to each connection
	}
}

// privateMessage sends a message only to one target user
func privateMessage(sender string, target string, message string) {
	for _, client := range clients { // loop over all clients
		if client.Name == target { // if name matches target
			client.Conn.Write([]byte("[PM from " + sender + "] " + message + "\n")) // send private message
			return                                                                  // stop after first match
		}
	}
}

// handleClient manages one client in a goroutine
func handleClient(conn net.Conn) {
	defer conn.Close() // ensure connection closes when function ends

	fmt.Println("A client connected!") // log connection
	conn.Write([]byte("Welcome!\n"))   // greet client

	reader := bufio.NewReader(conn) // create buffered reader for this connection

	conn.Write([]byte("Enter your name: ")) // ask for name
	name, _ := reader.ReadString('\n')      // read name line
	name = strings.TrimSpace(name)          // trim spaces/newline

	fmt.Println("Client name is:", name) // log name

	client := Client{Name: name, Conn: conn} // create Client struct
	clients = append(clients, client)        // add to global slice

	fmt.Println("Clients online:", len(clients)) // log count

	broadcast(name + " has joined the chat.\n") // announce join to everyone

	for { // main chat loop
		msg, err := reader.ReadString('\n') // read one line from client
		if err != nil {                     // if read fails (disconnect)
			fmt.Println(name, "disconnected.")        // log disconnect
			broadcast(name + " has left the chat.\n") // announce leave
			for i, c := range clients {               // loop over clients
				if c.Conn == conn { // find matching connection
					clients = append(clients[:i], clients[i+1:]...) // remove from slice
					break                                           // stop loop
				}
			}
			fmt.Println("Clients online:", len(clients)) // log new count
			return                                       // end goroutine
		}

		msg = strings.TrimSpace(msg) // trim spaces/newline from message

		if msg == "/quit" || msg == "/exit" { // if user typed quit command
			fmt.Println(name, "requested to leave.")  // log request
			broadcast(name + " has left the chat.\n") // announce leave
			for i, c := range clients {               // loop over clients
				if c.Conn == conn { // find matching connection
					clients = append(clients[:i], clients[i+1:]...) // remove from slice
					break                                           // stop loop
				}
			}
			fmt.Println("Clients online:", len(clients)) // log new count
			return                                       // end goroutine
		}

		if msg == "/users" { // if user wants list of users
			conn.Write([]byte("Online users:\n")) // header line
			for _, c := range clients {           // loop over clients
				conn.Write([]byte(" - " + c.Name + "\n")) // print each name
			}
			continue // skip normal broadcast for this command
		}

		if strings.HasPrefix(msg, "@") { // if message starts with '@' (private)
			parts := strings.SplitN(msg, " ", 2) // split into "@user" and rest
			if len(parts) == 2 {                 // ensure we have both parts
				target := parts[0][1:]                // remove '@' from username
				message := parts[1]                   // private message text
				privateMessage(name, target, message) // send private message
			} else { // invalid format
				conn.Write([]byte("Invalid private message format. Use: @username message\n")) // show help
			}
			continue // skip normal broadcast
		}

		timestamp := time.Now().Format("15:04") // get current time HH:MM

		broadcast("[" + timestamp + "] " + name + ": " + msg + "\n") // send formatted message to all
	}
}

func main() {
	port := "8989" // default port

	argsCount := len(os.Args) // number of CLI args

	if argsCount > 2 { // too many args
		fmt.Println("[USAGE]: ./TCPChat $port") // usage message
		return                                  // exit
	}

	if argsCount == 2 { // if user provided port
		port = os.Args[1] // override default port
	}

	fmt.Println("Server starting on port", port) // log startup

	listener, err := net.Listen("tcp", "localhost:"+port) // start TCP listener
	if err != nil {                                       // if listen fails
		fmt.Println("Error cannot open port:", err) // log error
		return                                      // exit
	}
	defer listener.Close() // ensure listener closes on exit

	for { // accept loop
		conn, err := listener.Accept() // wait for new connection
		if err != nil {                // if accept fails
			fmt.Println("Error accepting connection:", err) // log error
			continue                                        // try next
		}

		go handleClient(conn) // start goroutine for this client
	}
}
