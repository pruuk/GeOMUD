package main

import (
	"strings"
	"sync"
)

// recorder keeps one bot's whole transcript across its connections: the
// server's text as the waits see it (ANSI and carriage returns stripped) and
// each line the bot sent, marked "> ". A nil recorder records nothing.
type recorder struct {
	mu sync.Mutex
	b  strings.Builder
}

func (r *recorder) text(s string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.b.WriteString(s)
	r.mu.Unlock()
}

func (r *recorder) sent(line string) {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.b.WriteString("\n> " + line + "\n")
	r.mu.Unlock()
}

func (r *recorder) String() string {
	if r == nil {
		return ""
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.b.String()
}
