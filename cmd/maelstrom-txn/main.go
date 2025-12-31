package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

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

func main() {
	n := maelstrom.NewNode()

	// Linearizable key-value store
	// A sequential store seems to work too
	kv := maelstrom.NewLinKV(n)

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

		ctx := context.Background()

		// Store read results to maintain original order in the response
		readResults := make([]any, len(req.Txn))

		// Pass 1: Perform all reads
		for i, op := range req.Txn {
			kind := op[0].(string)
			if kind == "r" {
				key := fmt.Sprintf("%v", op[1])
				val, err := kv.ReadInt(ctx, key)
				if err != nil {
					readResults[i] = nil
				} else {
					readResults[i] = val
				}
			}
		}

		// Pass 2: Perform all writes
		for _, op := range req.Txn {
			kind := op[0].(string)
			if kind == "w" {
				key := fmt.Sprintf("%v", op[1])
				err := kv.Write(ctx, key, op[2])
				if err != nil {
					log.Printf("write error on key %s: %v", key, err)
				}
			}
		}

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

	// Execute the node's message loop. This will run until STDIN is closed.
	if err := n.Run(); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}
