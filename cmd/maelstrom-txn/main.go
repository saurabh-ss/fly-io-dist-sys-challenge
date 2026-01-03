package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"

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

type txnError struct {
	Type      string `json:"type"`
	InReplyTo int    `json:"in_reply_to"`
	Code      int    `json:"code"`
	Text      string `json:"text"`
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

		releaseLocks := func(ctx context.Context, lockList []string) {
			for i := len(lockList) - 1; i >= 0; i-- {
				key := lockList[i]
				err := kv.Write(ctx, key, 0)
				if err != nil {
					log.Printf("unlock error on key %s: %v", key, err)
				}
			}
		}

		ctx := context.Background()

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

		// Union of read and write sets
		unionSet := make(map[string]struct{})
		for k := range readSet {
			unionSet[k] = struct{}{}
		}
		for k := range writeSet {
			unionSet[k] = struct{}{}
		}

		var keysList []string
		for k := range unionSet {
			keysList = append(keysList, k)
		}
		sort.Strings(keysList)

		var lockList []string

		// Acquire locks for all keys in the union set
	lockLoop:
		for _, key := range keysList {
			lockKey := fmt.Sprintf("lock-%s", key)
			err := kv.CompareAndSwap(ctx, lockKey, 0, 1, true)
			if err == nil {
				lockList = append(lockList, lockKey)
			}
			if err != nil {
				log.Printf("lock error on key %s: %v", lockKey, err)
				log.Printf("Aborting transaction due to conflict with another transaction")
				// Release locks in reverse order
				releaseLocks(ctx, lockList)
				goto lockLoop
				// return n.Reply(
				// 	msg,
				// 	map[string]any{
				// 		"type": "error",
				// 		"code": maelstrom.TxnConflict,
				// 		"text": "txn abort",
				// 	},
				// )
			}
		}

		// Log all acquired locks
		log.Printf("Acquired locks: %v", lockList)

		// Store read results to maintain original order in the response
		readResults := make([]any, len(req.Txn))

		// Perform all reads
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

		// Perform all writes
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

		// Release locks for all keys
		releaseLocks(ctx, lockList)

		// Log all released locks
		log.Printf("Released locks: %v", lockList)

		return n.Reply(msg, resp)
	})

	// Execute the node's message loop. This will run until STDIN is closed.
	if err := n.Run(); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}
