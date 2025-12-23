package main

import (
	"encoding/json"
	"log"
	"os"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()

	messages := map[int]bool{}
	var messagesMu sync.RWMutex
	var neighbors []string

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
			for _, node := range n.NodeIDs() {
				// for _, node := range neighbors {
				log.Println("Node: ", node)
				if node != n.ID() && node != msg.Src {
					n.Send(node, map[string]any{"type": "replicate", "message": message})
				}
			}
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

	// Execute the node's message loop. This will run until STDIN is closed.
	if err := n.Run(); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}
