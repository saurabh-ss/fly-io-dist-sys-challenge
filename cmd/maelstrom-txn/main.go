package main

import (
	"encoding/json"
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

	n.Handle("txn", func(msg maelstrom.Message) error {
		var body txnRequest
		if err := json.Unmarshal(msg.Body, &body); err != nil {
			return err
		}

		resp := txnResponse{
			Type:      "txn_ok",
			InReplyTo: body.MsgID,
			Txn:       body.Txn,
		}

		return n.Reply(msg, resp)
	})

	// Execute the node's message loop. This will run until STDIN is closed.
	if err := n.Run(); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}
