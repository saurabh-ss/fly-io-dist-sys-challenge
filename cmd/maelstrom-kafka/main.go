package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"slices"
	"sort"
	"sync"

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

type PollOkBody struct {
	Type string             `json:"type"`
	Msgs map[string][][]int `json:"msgs"`
}

type ListCommittedOffsetsBody struct {
	Type string   `json:"type"`
	Keys []string `json:"keys"`
}

func main() {

	n := maelstrom.NewNode()

	// Storing commit offsets
	commitOffsets := maelstrom.NewLinKV(n)

	// Storing global offset
	kv := maelstrom.NewLinKV(n)

	// Main append-only log
	appendLog := make(map[string][]Pair)
	var appendLogMu sync.RWMutex

	getGlobalOffset := func() int {
		ctx := context.Background()
		for {
			val, err := kv.ReadInt(ctx, "global_offset")
			if err != nil {
				val = 0
			}
			err = kv.CompareAndSwap(ctx, "global_offset", val, val+1, true)
			if err == nil {
				return val + 1
			}
		}
	}

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

		offsetVal := getGlobalOffset()

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

		found := slices.Contains(n.NodeIDs(), m.Src)
		if !found {
			// Message is coming from a client
			var wg sync.WaitGroup
			var mu sync.Mutex

			for _, nei := range n.NodeIDs() {
				if nei != n.ID() {
					wg.Add(1)
					if err := n.RPC(nei, body, func(reply maelstrom.Message) error {
						defer wg.Done()
						var replyBody PollOkBody
						if err := json.Unmarshal(reply.Body, &replyBody); err != nil {
							return err
						}

						mu.Lock()
						defer mu.Unlock()
						for k, v := range replyBody.Msgs {
							msgs[k] = append(msgs[k], v...)
						}
						return nil
					}); err != nil {
						wg.Done()
					}
				}
			}
			wg.Wait()

			for k, v := range msgs {
				slices.SortFunc(v, func(a, b []int) int {
					return a[0] - b[0]
				})
				msgs[k] = v
			}
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
		ctx := context.Background()

		for key, offset := range body.Offsets {
			commitOffsets.Write(ctx, key, offset)
		}

		return n.Reply(m, map[string]any{"type": "commit_offsets_ok"})
	})

	n.Handle("list_committed_offsets", func(m maelstrom.Message) error {
		var body ListCommittedOffsetsBody
		if err := json.Unmarshal(m.Body, &body); err != nil {
			return err
		}

		msgs := make(map[string]int)
		ctx := context.Background()
		for _, key := range body.Keys {
			val, _ := commitOffsets.ReadInt(ctx, key)
			msgs[key] = val
		}

		return n.Reply(m, map[string]any{"type": "list_committed_offsets_ok", "offsets": msgs})

	})

	// Execute the node's message loop. This will run until STDIN is closed.
	if err := n.Run(); err != nil {
		log.Printf("ERROR: %s", err)
		os.Exit(1)
	}
}
