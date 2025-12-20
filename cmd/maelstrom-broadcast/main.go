package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

func main() {
	n := maelstrom.NewNode()
	var ts uint64 = 0

	// This is a slice of int64 values, named 'messages'.
	// messages := []int64{}

	messages := map[string]int64{}

	var top map[string][]string

	n.Handle("broadcast", func(msg maelstrom.Message) error {
		var body map[string]any

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// messages = append(messages, int64(body["message"].(float64)))
		ts += 1
		key := fmt.Sprintf("%s-%d", n.ID(), ts)
		val := int64(body["message"].(float64))
		messages[key] = val

		// fmt.Println("Source of message: ", msg.Src)
		repbody := map[string]any{
			"type": "replicate",
			"key":  key,
			"val":  val,
		}
		if top != nil {
			for _, neighbor := range top[n.ID()] {
				n.Send(neighbor, repbody)
			}
		}

		return n.Reply(msg, map[string]any{"type": "broadcast_ok"})
	})

	n.Handle("replicate", func(msg maelstrom.Message) error {
		var body map[string]any

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}
		key := body["key"].(string)
		val := int64(body["val"].(float64))
		if _, exists := messages[key]; !exists {
			messages[key] = val

			repbody := map[string]any{
				"type": "replicate",
				"key":  key,
				"val":  val,
			}
			if top != nil {
				for _, neighbor := range top[n.ID()] {
					n.Send(neighbor, repbody)
				}
			}
		}
		return nil
		// return n.Reply(msg, map[string]any{"type": "replicate_ok"})
	})

	n.Handle("read", func(msg maelstrom.Message) error {
		var body map[string]any

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		// Extract all values from the messages map into a slice
		messageValues := make([]int64, 0, len(messages))
		for _, val := range messages {
			messageValues = append(messageValues, val)
		}

		body["type"] = "read_ok"
		body["messages"] = messageValues
		return n.Reply(msg, body)
	})

	n.Handle("topology", func(msg maelstrom.Message) error {
		var body struct {
			Topology map[string][]string `json:"topology"`
		}

		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		top = body.Topology

		return n.Reply(msg, map[string]any{"type": "topology_ok"})
	})

	// _ = top // Will be used for gossip protocol later

	// Execute the node's message loop. This will run until STDIN is closed.
	if err := n.Run(); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}
