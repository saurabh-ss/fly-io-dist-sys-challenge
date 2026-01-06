# Fly.io Distributed Systems Challenge

Go implementation of the [Fly.io Distributed Systems Challenge](https://fly.io/dist-sys/).

## Project Structure

This project uses the standard Go project layout with multiple binaries:

```
fly/
├── go.mod                    # Root Go module
├── go.sum                    # Dependency checksums
├── cmd/                      # Command-line applications
│   ├── maelstrom-echo/       # Echo challenge
│   │   └── main.go
│   ├── maelstrom-unique-ids/ # Unique IDs challenge
│   │   └── main.go
│   ├── maelstrom-broadcast/  # Broadcast challenge
│   │   └── main.go
│   ├── maelstrom-counter/    # Counter challenge
│   │   └── main.go
│   ├── maelstrom-kafka/      # Kafka challenge
│   │   └── main.go
│   └── maelstrom-txn/        # Transaction challenge
│       └── main.go
├── bin/                      # Compiled binaries (gitignored)
├── internal/                 # Internal shared code (when needed)
└── maelstrom/                # Maelstrom test harness
```

## Building

### Using Makefile (Recommended)

```bash
# Build all binaries
make build

# Build and install to $GOPATH/bin
make install

# Clean build artifacts
make clean

# Run tests
make test

# Show all available targets
make help
```

### Using Go Commands Directly

```bash
# Build echo service
go build -o bin/maelstrom-echo ./cmd/maelstrom-echo

# Build unique IDs service
go build -o bin/maelstrom-unique-ids ./cmd/maelstrom-unique-ids

# Build broadcast service
go build -o bin/maelstrom-broadcast ./cmd/maelstrom-broadcast

# Build all binaries
go build -o bin/ ./cmd/...
```

## Running Tests

Use the Maelstrom test harness to test your implementations:

```bash
# Test echo service
./maelstrom/maelstrom test -w echo --bin bin/maelstrom-echo --node-count 1 --time-limit 10

# Test unique IDs service
./maelstrom/maelstrom test -w unique-ids --bin bin/maelstrom-unique-ids --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition
```

## Development

### Adding New Challenges

To add a new challenge:

1. Create a new directory under `cmd/`: `cmd/my-new-challenge/`
2. Add a `main.go` file with `package main`
3. Build it: `go build -o bin/my-new-challenge ./cmd/my-new-challenge`

### Shared Code

If you need to share code between challenges, create packages in:
- `internal/` for project-specific shared code (create this directory when needed)

## Dependencies

- [Maelstrom Go library](https://github.com/jepsen-io/maelstrom) - Distributed systems testing framework
- [google/uuid](https://github.com/google/uuid) - UUID generation

