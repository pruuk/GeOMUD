// Command smoke is the end-to-end smoke test. First an admin logs in and
// spawns one item per bot on the floor of the start room. Then the bots log
// in over telnet, create a character, pick up an item, drop it, pick it up
// again, walk, say, quit, log back in and check the item is still held.
//
// By default it builds and launches its own server on a temp copy of the
// empty world (run it from the repo root). With -addr it plays against a
// server that is already running; the admin account is then set with
// -admin-user and -admin-pass.
//
//	go run ./cmd/smoke
//	go run ./cmd/smoke -bots 50 -hold 2m
//	go run ./cmd/smoke -addr 127.0.0.1:33333 -bots 50 -admin-pass secret
package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"sync"
	"syscall"
	"time"
)

func main() {
	os.Exit(run())
}

func run() int {
	addr := flag.String("addr", "", "host:port of a running server; empty builds and launches one")
	bots := flag.Int("bots", 1, "number of bots to run at once")
	hold := flag.Duration("hold", 0, "how long bots stay logged in after the script, e.g. 2m")
	timeout := flag.Duration("timeout", 30*time.Second, "how long each step waits for the server's reply")
	serverLog := flag.String("server-log", "", "launch mode: copy the server's log to this file when the run ends")
	adminUser := flag.String("admin-user", "admin", "admin account that spawns the bots' items (launch mode: the temp world's stock admin)")
	adminPass := flag.String("admin-pass", "password", "admin's current password (a forced first-login change sets it to this same text)")
	transcript := flag.String("transcript", "", "directory to write each bot's transcript to (bot<N>.txt); empty writes none")
	flag.Parse()
	if *bots < 1 {
		fmt.Fprintln(os.Stderr, "smoke: -bots must be at least 1")
		return 2
	}

	adm := adminLogin{*adminUser, *adminPass}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if *addr == "" {
		start := time.Now()
		srv, err := launch(ctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, "smoke: launch:", err)
			return 2
		}
		srv.copyLogTo = *serverLog
		passed := false
		defer func() { srv.stop(!passed) }()
		fmt.Printf("smoke: server pid %d listening on %s (ready in %s)\n",
			srv.cmd.Process.Pid, srv.addr, time.Since(start).Round(time.Millisecond))
		*addr = srv.addr
		passed = runBots(ctx, *addr, *bots, *timeout, *hold, adm, *transcript)
		if !passed {
			return 1
		}
		return 0
	}

	if !runBots(ctx, *addr, *bots, *timeout, *hold, adm, *transcript) {
		return 1
	}
	return 0
}

type result struct {
	id   int
	bot  bot
	took time.Duration
	err  error
	gmcp map[string]int // GMCP packages the bot received
}

type adminLogin struct{ user, pass string }

// runBots has the admin spawn one item per bot, plays n bots at once and
// prints a summary. It reports whether every bot passed.
func runBots(ctx context.Context, addr string, n int, timeout, hold time.Duration, adm adminLogin, transcriptDir string) bool {
	if transcriptDir != "" {
		if err := os.MkdirAll(transcriptDir, 0o755); err != nil {
			fmt.Printf("smoke: FAIL transcript dir: %v\n", err)
			return false
		}
	}
	tag := runTag()
	fmt.Printf("smoke: %d bot(s) against %s, run tag %s\n", n, addr, tag)

	start := time.Now()
	if err := prepareItems(ctx, addr, adm.user, adm.pass, n, timeout); err != nil {
		fmt.Printf("smoke: FAIL admin could not spawn the items: %v\n", err)
		return false
	}
	fmt.Printf("smoke: admin spawned %d %s(s) in %s\n", n, smokeItem, time.Since(start).Round(time.Millisecond))
	results := make([]result, n)
	var wg sync.WaitGroup
	for i := range n {
		b := bot{
			user: fmt.Sprintf("smk%su%d", tag, i),
			pass: "smoke" + tag,
			char: fmt.Sprintf("Smk%sc%d", tag, i),
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			t := time.Now()
			var rec *recorder
			if transcriptDir != "" {
				rec = &recorder{}
			}
			pkgs, err := play(ctx, addr, b, timeout, hold, rec)
			if rec != nil {
				name := filepath.Join(transcriptDir, fmt.Sprintf("bot%d.txt", i))
				if werr := os.WriteFile(name, []byte(rec.String()), 0o644); werr != nil && err == nil {
					err = fmt.Errorf("write transcript: %w", werr)
				}
			}
			took := time.Since(t)
			if err == nil {
				took -= hold // a failed bot never reached the hold
			}
			results[i] = result{i, b, took, err, pkgs}
		}()
	}
	wg.Wait()
	total := time.Since(start)

	var failed []result
	var times []time.Duration
	allGMCP := map[string]int{}
	for _, r := range results {
		for n, k := range r.gmcp {
			allGMCP[n] += k
		}
		times = append(times, r.took)
		if r.err != nil {
			failed = append(failed, r)
		}
	}
	for _, r := range failed {
		fmt.Printf("FAIL bot %d (%s / %s) after %s: %v\n", r.id, r.bot.user, r.bot.char, r.took.Round(time.Millisecond), r.err)
	}
	fmt.Printf("smoke: GMCP packages received (all bots, message counts): %s\n", gmcpList(allGMCP))
	sort.Slice(times, func(i, j int) bool { return times[i] < times[j] })
	verdict := "PASS"
	if len(failed) > 0 {
		verdict = "FAIL"
	}
	fmt.Printf("smoke: %s %d/%d bots passed in %s (script time per bot: min %s, median %s, max %s; hold %s)\n",
		verdict, n-len(failed), n, total.Round(time.Millisecond),
		times[0].Round(time.Millisecond), times[len(times)/2].Round(time.Millisecond),
		times[len(times)-1].Round(time.Millisecond), hold)
	return len(failed) == 0
}

// runTag makes account names unique per run, so -addr runs against a
// long-lived server never collide with earlier runs.
func runTag() string {
	b := make([]byte, 3)
	rand.Read(b)
	return hex.EncodeToString(b)
}
