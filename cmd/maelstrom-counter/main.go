package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	node := maelstrom.NewNode()
	kv := maelstrom.NewSeqKV(node)
	key := "counter"
	counter := 0

	node.Handle("add", func(msg maelstrom.Message) error {
		var body map[string]any

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		delta := int(body["delta"].(float64))
		counter += delta

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		err := kv.Write(ctx, key, counter)
		if err != nil {
			return err
		}

		return node.Reply(msg, map[string]any{"type": "add_ok"})
	})

	node.Handle("read", func(msg maelstrom.Message) error {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		value, err := kv.ReadInt(ctx, key)

		if err != nil {
			return err
		}
		return node.Reply(msg, map[string]any{"type": "read_ok", "value": value})
	})

	// Execute the node's message loop. This will run until STDIN is closed.
	if err := node.Run(); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}
