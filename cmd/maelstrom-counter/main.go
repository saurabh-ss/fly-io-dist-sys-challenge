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

	node.Handle("add", func(msg maelstrom.Message) error {
		var body map[string]any

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		delta := int(body["delta"].(float64))
		key := node.ID()

		ctx := context.Background()

		for {
			val, err := kv.ReadInt(ctx, key)
			if err != nil {
				val = 0
			}
			err = kv.CompareAndSwap(ctx, key, val, val+delta, true)
			if err != nil {
				log.Printf("CAS failed from %d to %d", val, val+delta)
			} else {
				return node.Reply(msg, map[string]any{"type": "add_ok"})
			}
		}
	})

	node.Handle("read", func(msg maelstrom.Message) error {
		var body map[string]any

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		ctx := context.Background()
		kv.Write(ctx, "last", int(body["msg_id"].(float64)))

		total := 0
		for _, n := range node.NodeIDs() {
			value, err := kv.ReadInt(ctx, n)
			if err != nil {
				value = 0
			}
			total += value
		}
		return node.Reply(msg, map[string]any{"type": "read_ok", "value": total})
	})

	// Execute the node's message loop. This will run until STDIN is closed.
	if err := node.Run(); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}
