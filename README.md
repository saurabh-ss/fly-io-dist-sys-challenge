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
# Challenge #2: Unique ID Generation
./maelstrom/maelstrom test -w unique-ids --bin bin/maelstrom-unique-ids --time-limit 30 --rate 1000 --node-count 3 --availability total --nemesis partition

# Challenge #3a: Single-Node Broadcast
./maelstrom/maelstrom test -w broadcast --bin bin/maelstrom-broadcast --node-count 1 --time-limit 20 --rate 10

# Challenge #3b: Multi-Node Broadcast
./maelstrom/maelstrom test -w broadcast --bin bin/maelstrom-broadcast --node-count 5 --time-limit 20 --rate 10

# Challenge #3c: Fault Tolerant Broadcast
./maelstrom/maelstrom test -w broadcast --bin bin/maelstrom-broadcast --node-count 5 --time-limit 20 --rate 10 --nemesis partition

# Challenge #3d: Efficient Broadcast, Part I
# Target messages-per-operation < 30, median latency < 400 ms, maximum latency < 600 ms
./maelstrom/maelstrom test -w broadcast --bin bin/maelstrom-broadcast --node-count 25 --time-limit 20 --rate 100 --latency 100

# Challenge #3e: Efficient Broadcast, Part II
# Target messages-per-operation < 20, median latency < 1000 ms, maximum latency < 2000 ms
./maelstrom/maelstrom test -w broadcast --bin bin/maelstrom-broadcast --node-count 25 --time-limit 20 --rate 100 --latency 100

# Challenge #4: Grow-Only Counter
./maelstrom/maelstrom test -w g-counter --bin bin/maelstrom-counter --node-count 3 --rate 100 --time-limit 20 --nemesis partition

# Challenge #5a: Single-Node Kafka-Style Log
./maelstrom/maelstrom test -w kafka --bin bin/maelstrom-kafka --node-count 1 --concurrency 2n --time-limit 20 --rate 1000

# Challenge #5b: Multi-Node Kafka-Style Log
./maelstrom/maelstrom test -w kafka --bin bin/maelstrom-kafka --node-count 2 --concurrency 2n --time-limit 20 --rate 1000

# Challenge #5c: Efficient Kafka-Style Log
./maelstrom/maelstrom test -w kafka --bin bin/maelstrom-kafka --node-count 2 --concurrency 2n --time-limit 20 --rate 1000

# Challenge #6a: Single-Node, Totally-Available Transactions
./maelstrom/maelstrom test -w txn-rw-register --bin bin/maelstrom-txn --node-count 1 --time-limit 20 --rate 1000 --concurrency 2n --consistency-models read-uncommitted --availability total

# Challenge #6b: Totally-Available, Read Uncommitted Transactions
./maelstrom/maelstrom test -w txn-rw-register --bin bin/maelstrom-txn --node-count 2 --concurrency 2n --time-limit 20 --rate 1000 --consistency-models read-uncommitted
# Ensure total availability in face of network partition
./maelstrom/maelstrom test -w txn-rw-register --bin bin/maelstrom-txn --node-count 2 --concurrency 2n --time-limit 20 --rate 1000 --consistency-models read-uncommitted --availability total --nemesis partition

# Challenge #6c: Totally-Available, Read Committed Transactions
./maelstrom/maelstrom test -w txn-rw-register --bin bin/maelstrom-txn --node-count 2 --concurrency 2n --time-limit 20 --rate 1000 --consistency-models read-committed --availability total --nemesis partition
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

