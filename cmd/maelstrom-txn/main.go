package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type txnRequest struct {
	Type  string  `json:"type"`
	MsgID int     `json:"msg_id"`
	Txn   [][]any `json:"txn"`
}

type txnResponse struct {
	Type      string  `json:"type"`
	InReplyTo int     `json:"in_reply_to"`
	Txn       [][]any `json:"txn"`
}

type replicateRequest struct {
	Type  string  `json:"type"`
	MsgID int     `json:"msg_id"`
	Txn   [][]any `json:"txn"`
}

func main() {
	n := maelstrom.NewNode()

	store := make(map[string]int)
	var storeMu sync.RWMutex

	n.Handle("txn", func(msg maelstrom.Message) error {
		var req txnRequest
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return err
		}

		resp := txnResponse{
			Type:      "txn_ok",
			InReplyTo: req.MsgID,
			Txn:       make([][]any, 0, len(req.Txn)),
		}

		readSet := make(map[string]struct{})
		writeSet := make(map[string]struct{})

		// Collect read and write sets
		for _, op := range req.Txn {
			kind := op[0].(string)
			switch kind {
			case "r":
				readSet[fmt.Sprintf("%v", op[1])] = struct{}{}
			case "w":
				writeSet[fmt.Sprintf("%v", op[1])] = struct{}{}
			}
		}

		// Store read results to maintain original order in the response
		readResults := make([]any, len(req.Txn))

		// Perform all reads with read lock
		storeMu.RLock()
		for i, op := range req.Txn {
			kind := op[0].(string)
			if kind == "r" {
				key := fmt.Sprintf("%v", op[1])
				val, exists := store[key]
				if !exists {
					readResults[i] = nil
				} else {
					readResults[i] = val
				}
			}
		}
		storeMu.RUnlock()

		// Perform all writes with write lock
		storeMu.Lock()
		for _, op := range req.Txn {
			kind := op[0].(string)
			if kind == "w" {
				key := fmt.Sprintf("%v", op[1])
				val := int(op[2].(float64))
				store[key] = val
			}
		}
		storeMu.Unlock()

		go func() {
			ctx := context.Background()
			for _, node := range n.NodeIDs() {
				if node != n.ID() {
					n.SyncRPC(ctx, node, replicateRequest{
						Type:  "replicate",
						MsgID: req.MsgID,
						Txn:   req.Txn,
					})
				}
			}
		}()

		// Reconstruct response in the original order
		for i, op := range req.Txn {
			kind := op[0].(string)
			switch kind {
			case "r":
				resp.Txn = append(resp.Txn, []any{kind, op[1], readResults[i]})
			case "w":
				resp.Txn = append(resp.Txn, []any{kind, op[1], op[2]})
			}
		}

		return n.Reply(msg, resp)
	})

	n.Handle("replicate", func(msg maelstrom.Message) error {
		var req replicateRequest
		if err := json.Unmarshal(msg.Body, &req); err != nil {
			return err
		}

		storeMu.Lock()
		for _, op := range req.Txn {
			kind := op[0].(string)
			if kind == "w" {
				key := fmt.Sprintf("%v", op[1])
				val := int(op[2].(float64))
				store[key] = val
			}
		}
		storeMu.Unlock()

		return n.Reply(msg, map[string]any{
			"type": "replicate_ok",
		})
	})

	// Execute the node's message loop. This will run until STDIN is closed.
	if err := n.Run(); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}
