package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()

	messages := map[int]bool{}
	var messagesMu sync.RWMutex
	var neighbors []string

	pending := map[string]map[int]struct{}{}
	var pendingMu sync.RWMutex

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body map[string]any

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		message := int(body["message"].(float64))

		messagesMu.Lock()
		_, exists := messages[message]
		if !exists {
			messages[message] = true
		}
		messagesMu.Unlock()

		if !exists {
			pendingMu.Lock()
			for _, node := range n.NodeIDs() {
				if node != n.ID() && node != msg.Src {
					if pending[node] == nil {
						pending[node] = make(map[int]struct{})
					}
					pending[node][message] = struct{}{}
					n.Send(node, map[string]any{"type": "replicate", "message": message})
				}
			}
			pendingMu.Unlock()

		}

		return n.Reply(msg, map[string]any{"type": "broadcast_ok"})
	})

	n.Handle("replicate", func(msg maelstrom.Message) error {
		var body map[string]any

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		message := int(body["message"].(float64))

		messagesMu.Lock()
		_, exists := messages[message]
		if !exists {
			messages[message] = true
		}
		messagesMu.Unlock()

		body["type"] = "replicate_ok"
		n.Send(msg.Src, body)
		return nil
	})

	n.Handle("replicate_ok", func(msg maelstrom.Message) error {
		var body map[string]any

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		// Remove the message from the pending set for this node, if present
		message := int(body["message"].(float64))

		pendingMu.Lock()
		if nodePending, exists := pending[msg.Src]; exists {
			delete(nodePending, int(message))
			if len(nodePending) == 0 {
				delete(pending, msg.Src)
			}
		}
		pendingMu.Unlock()

		return nil
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		var body map[string]any

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// Extract all values from the messages map into a slice
		messagesMu.RLock()
		messageValues := make([]int, 0, len(messages))
		for key := range messages {
			messageValues = append(messageValues, key)
		}
		messagesMu.RUnlock()

		body["type"] = "read_ok"
		body["messages"] = messageValues
		return n.Reply(msg, body)
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		var body map[string]any

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// We can assume n.ID() is present in topology
		neighborList := body["topology"].(map[string]any)[n.ID()].([]any)
		neighbors = neighbors[:0]
		for _, v := range neighborList {
			neighbors = append(neighbors, v.(string))
		}
		log.Print("Neighbors: ", neighbors)

		return n.Reply(msg, map[string]any{"type": "topology_ok"})
	})

	go func() {
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			pendingMu.RLock()
			total := 0
			for node, messages := range pending {
				total += len(messages)
				for message := range messages {
					n.Send(node, map[string]any{"type": "replicate", "message": message})
				}
			}
			pendingMu.RUnlock()
			log.Printf("Current size of pending map: %d", total)
		}
	}()

	// Execute the node's message loop. This will run until STDIN is closed.
	if err := n.Run(); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}
