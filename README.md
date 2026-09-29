TCPChat — Multi‑Client TCP Chat Server (Go)

A simple but fully functional multi‑client chat server written in Go.
Clients connect using Netcat (nc) or Telnet and can chat in real time.

This project was built step‑by‑step while learning Go networking, concurrency, and TCP communication.

FEATURES IMPLEMENTED (Up to Step 10)

✔ Multi‑Client Support
The server accepts unlimited clients using net.Listen and goroutines.

✔ Username System
Each client is asked to enter a name when connecting.

✔ Broadcasting
Messages are sent to all connected clients.

✔ Join & Leave Announcements
When a user joins or leaves, all clients are notified.

✔ Clean Disconnect
Users can leave using:
/quit
/exit

The server removes them from the active client list.

✔ Online User List
Users can see who is online using:
/users

✔ Private Messaging
Send a private message to a specific user:
@username message

✔ Timestamps
All messages include a timestamp in HH:MM format.

✔ Ghost‑User Prevention
Disconnected users are removed from the clients slice.

COMMANDS

/quit      Disconnect from the chat
/exit      Same as /quit
/users     Show all online users
@name msg  Send a private message to a specific user

HOW TO RUN THE SERVER

Build the program:
go build -o TCPChat

Run the server (default port 8989):
./TCPChat

Run the server on a custom port:
./TCPChat 8080

HOW TO CONNECT AS A CLIENT

Open a terminal and run:

nc localhost 8989

Or if using a custom port:

nc localhost 8080

Each terminal acts as a separate chat client.

HOW IT WORKS (Short Explanation)

The server listens for TCP connections.

Each new client is handled in a separate goroutine.

Clients are stored in a global slice.

Messages are read using bufio.Reader.

Broadcasting loops through all connected clients.

Disconnects remove the client from the slice.

Commands are parsed before broadcasting.

Private messages are sent only to the target user.

Timestamps are added using time.Now().Format("15:04").

REQUIREMENTS

Go 1.18+

Netcat (nc) or Telnet

Username colors

Message history

Admin commands

Max client limit

Rate limiting

Logging system

