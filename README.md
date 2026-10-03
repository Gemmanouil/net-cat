# TCPChat: Multi-Client TCP Chat Server (Go)

A group chat server written in Go that recreates the core idea of NetCat in a
server-client architecture. Clients connect with `nc` (Netcat) and chat in real time.

Built while learning Go networking, goroutines, and mutexes, using only the
standard library.

## Features

- **TCP server with multiple clients** (1 server to many clients), one goroutine per client.
- **Name required:** the client is greeted with a Linux logo and must enter a
  non-empty, unique name.
- **Connection limit:** maximum 10 simultaneous connections; extra clients are told the chat is full.
- **Timestamped messages:** every message is formatted as
  `[2026-09-29 18:34:23][name]:message`.
- **Empty messages are ignored:** they are never broadcast.
- **Message history:** a client that joins receives all previous chat messages.
- **Join and leave notices:** the other clients are informed when someone joins or leaves.
- **Clean disconnects:** when a client leaves, everyone else stays connected.
- **Safe concurrency:** all shared state (clients, history, connection count) is protected by a `sync.Mutex`.
- **Default port 8989**, or a custom port as the first argument.

## Requirements

- Go 1.18+
- `nc` (Netcat) to connect as a client

## Build and run

```bash
go mod init TCPChat        # only once, if there is no go.mod yet
go build -o TCPChat .

./TCPChat                  # listens on :8989
./TCPChat 2525             # listens on :2525
./TCPChat 2525 localhost   # prints: [USAGE]: ./TCPChat $port
```

For development you can also use `go run . 2525`, or `go run -race . 2525`
to detect data races.

## Connect as a client

```bash
nc localhost 8989          # or the custom port you chose
```

Each terminal running `nc` is a separate chat client. Press `Ctrl+C` to leave.

## Example

```
$ nc localhost 2525
Welcome to TCP-Chat!
         _nnnn_
        dGGGGMMb
        ...
[ENTER YOUR NAME]: Yenlik
[2026-09-29 16:03:43][Yenlik]:hello
Lee has joined our chat...
[2026-09-29 16:04:32][Lee]:Hi everyone!
Lee has left our chat...
```

## How it works

1. The server listens with `net.Listen` and accepts connections in a loop.
2. Before starting a goroutine, it reserves one of the 10 connection slots.
   If none is free, the client is told the chat is full and disconnected.
3. Each client goroutine sends the logo, asks for a name until it is valid,
   then registers the client.
4. On joining, the server replays the message history to the new client and
   announces the join to everyone else.
5. Each line a client sends is trimmed; empty lines are dropped, other lines are
   timestamped, stored in the history, and broadcast to the other clients.
6. When the connection closes, the client is removed and the others are notified.

Shared data is only touched while holding the server mutex, so joins, leaves
and broadcasts from different goroutines cannot corrupt each other.

## Allowed packages used

`bufio`, `errors`, `fmt`, `log`, `net`, `os`, `strings`, `sync`, `time`

## Possible extensions

- Clients can change their name, with a notice to the group
- Saving all logs to a file
- Multiple chat rooms
- Unit tests for the server and client
