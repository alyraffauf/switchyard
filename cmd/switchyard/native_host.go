// SPDX-License-Identifier: GPL-3.0-or-later

// Native-messaging host: stdin/stdout JSON framing with browser extensions.

package main

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
)

// Config save state shared with ui_helpers.go's debounced writer.
var (
	isSaving  bool
	savingMux sync.Mutex
)

type clientRequest struct {
	Action string `json:"action"`
}

type browserSummary struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func runNativeMessagingHost() {
	var length uint32
	if err := binary.Read(os.Stdin, binary.NativeEndian, &length); err != nil {
		return
	}
	payload := make([]byte, length)
	if _, err := io.ReadFull(os.Stdin, payload); err != nil {
		return
	}

	var req clientRequest
	if err := json.Unmarshal(payload, &req); err != nil {
		writeResponse(map[string]string{"error": "invalid request"})
		return
	}

	switch req.Action {
	case "listBrowsers":
		writeResponse(map[string]any{"browsers": listInstalledBrowsers()})
	default:
		writeResponse(map[string]string{"error": "unknown action: " + req.Action})
	}
}

func listInstalledBrowsers() []browserSummary {
	browsers := detectBrowsers()
	out := make([]browserSummary, 0, len(browsers))
	for _, b := range browsers {
		// detectBrowsers skips the currently-running variant; drop the other
		// one too so Devel and Stable never list each other.
		if strings.HasPrefix(b.ID, "io.github.alyraffauf.Switchyard") {
			continue
		}
		out = append(out, browserSummary{b.ID, b.Name})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func writeResponse(v any) {
	out, _ := json.Marshal(v)
	var prefix [4]byte
	binary.NativeEndian.PutUint32(prefix[:], uint32(len(out)))
	os.Stdout.Write(prefix[:])
	os.Stdout.Write(out)
}
