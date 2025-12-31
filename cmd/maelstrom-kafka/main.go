package main

import (
	"encoding/json"
	"log"
	"os"
	"sort"
	"sync"
	"sync/atomic"

	maelstrom "github.com/jepsen-io/maelstrom/demo/go"
)

type Pair struct {
	msg    int
	offset int
}

type CommitOffsetsBody struct {
	Type    string         `json:"type"`
	Offsets map[string]int `json:"offsets"`
}

type SendBody struct {
	Type string `json:"type"`
	Key  string `json:"key"`
	Msg  int    `json:"msg"`
}

type PollBody struct {
	Type    string         `json:"type"`
	Offsets map[string]int `json:"offsets"`
}

type ListCommittedOffsetsBody struct {
	Type string   `json:"type"`
	Keys []string `json:"keys"`
}

func main() {

	n := maelstrom.NewNode()

	// Main append-only log
	appendLog := make(map[string][]Pair)
	var appendLogMu sync.RWMutex

	// Offset vendor
	var globalOffset atomic.Int64

	// Storing commit offsets
	commitOffsets := make(map[string]int)
	var commitOffsetsMu sync.RWMutex

	getMessages := func(key string, startOffset int) []Pair {
		appendLogMu.RLock()
		defer appendLogMu.RUnlock()

		keyLog, found := appendLog[key]
		if found {
			idx := sort.Search(len(keyLog), func(i int) bool {
				return keyLog[i].offset >= startOffset
			})
			if idx < len(keyLog) {
				// Return a copy of the slice to avoid racing with future appends
				res := make([]Pair, len(keyLog[idx:]))
				copy(res, keyLog[idx:])
				return res
			}
		}
		return nil
	}

	n.Handle("send", func(m maelstrom.Message) error {
		var body SendBody
		if err := json.Unmarshal(m.Body, &body); err != nil {
			return err
		}

		offsetVal := globalOffset.Add(1)

		appendLogMu.Lock()
		keyLog, found := appendLog[body.Key]
		if !found {
			keyLog = []Pair{}
		}
		keyLog = append(keyLog, Pair{msg: body.Msg, offset: int(offsetVal)})
		appendLog[body.Key] = keyLog
		appendLogMu.Unlock()

		return n.Reply(m, map[string]any{"type": "send_ok", "offset": offsetVal})
	})

	n.Handle("poll", func(m maelstrom.Message) error {
		var body PollBody
		if err := json.Unmarshal(m.Body, &body); err != nil {
			return err
		}

		msgs := make(map[string][][]int)
		for key, startOffset := range body.Offsets {
			pairs := getMessages(key, startOffset)
			res := make([][]int, 0, len(pairs))
			for _, p := range pairs {
				res = append(res, []int{p.offset, p.msg})
			}
			msgs[key] = res
		}

		return n.Reply(m, map[string]any{
			"type": "poll_ok",
			"msgs": msgs,
		})
	})

	n.Handle("commit_offsets", func(m maelstrom.Message) error {
		var body CommitOffsetsBody
		if err := json.Unmarshal(m.Body, &body); err != nil {
			return err
		}

		commitOffsetsMu.Lock()
		for key, offset := range body.Offsets {
			commitOffsets[key] = offset
		}
		commitOffsetsMu.Unlock()

		return n.Reply(m, map[string]any{"type": "commit_offsets_ok"})
	})

	n.Handle("list_committed_offsets", func(m maelstrom.Message) error {
		var body ListCommittedOffsetsBody
		if err := json.Unmarshal(m.Body, &body); err != nil {
			return err
		}

		msgs := make(map[string]int)

		commitOffsetsMu.RLock()
		for _, key := range body.Keys {
			if offset, ok := commitOffsets[key]; ok {
				msgs[key] = offset
			}
		}
		commitOffsetsMu.RUnlock()

		return n.Reply(m, map[string]any{"type": "list_committed_offsets_ok", "offsets": msgs})

	})

	// Execute the node's message loop. This will run until STDIN is closed.
	if err := n.Run(); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}
