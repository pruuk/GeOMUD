package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// server is a GeOMUD server this program built and started. It runs from
// the repo root (the server always reads _datafiles/config.yaml from its
// working directory) against a copy of the empty world in a temp dir, so
// the repo's _datafiles are never written.
type server struct {
	dir  string // temp dir: world copy, overrides, log
	addr string
	cmd  *exec.Cmd
	done chan struct{} // closed when the process exits

	children []int // server processes copyover started; stop kills them too

	copyLogTo string // if set, stop copies the server log here before deleting the temp dir
}

// overrides returns the CONFIG_PATH file contents: the world copy and only
// the loopback LocalPort listener. The public telnet ports, web, SSH and AI
// listeners all bind every interface, which on Windows raises a firewall
// prompt; LocalPort binds 127.0.0.1 only and has no connection cap.
func overrides(world string, localPort int) string {
	return fmt.Sprintf(`FilePaths.DataFiles: %s
Network.TelnetPort: []
Network.LocalPort: %d
Network.HttpPort: 0
Network.HttpsPort: 0
Network.SSHPort: 0
Network.AI.Port: 0
`, strconv.Quote(filepath.ToSlash(world)), localPort)
}

// fightRats sets up launch mode's world copy for the smoke fight (Phase 5
// slice 4.1, ruling 11). Town Square (startland room 1) gets exactly two
// rats: a "small rat" pinned at level 1 and a "large rat" pinned at level 5
// (launchRats), neither wandering (an idle list of one empty command) and
// neither respawning within a run. End of the Line (room 2), where the
// shipped world's three rats live and from where they wander, gets none.
// The admin step still spawns one plain "rat" in Town Square (npcSpeaks),
// which may wander; bot 0 attacks "small" and "large", which never match
// it, and the fight checks wait for "small rat has died." and "large rat
// has died.". The rodent race's damage is unchanged:
// either rat fights back, and the fight step checks whichever ending comes.
// Only the temporary copy changes.
func fightRats(world string) error {
	dir := filepath.Join(world, "rooms", "startland")
	if err := editSpawns(filepath.Join(dir, "2.yaml"), func(spawns []any) ([]any, error) {
		for _, sp := range spawns {
			if mobIdOf(sp) != ratMobId {
				return nil, fmt.Errorf("room 2 spawns more than rats: update fightRats")
			}
		}
		return nil, nil
	}); err != nil {
		return err
	}
	return editSpawns(filepath.Join(dir, "1.yaml"), func(spawns []any) ([]any, error) {
		for _, sp := range spawns {
			if mobIdOf(sp) == ratMobId {
				return nil, fmt.Errorf("room 1 already spawns a rat: update fightRats")
			}
		}
		for i, level := range []int{1, 5} {
			spawns = append(spawns, yaml.MapSlice{
				{Key: "mobid", Value: ratMobId},
				{Key: "name", Value: launchRats[i].name},
				{Key: "level", Value: level},
				{Key: "idlecommands", Value: []string{""}},
				{Key: "respawnrate", Value: "24 real hours"},
			})
		}
		return spawns, nil
	})
}

// ratMobId is the rat's mob template (world/empty/mobs/startland/1-rat.yaml).
const ratMobId = 1

// editSpawns rewrites a room file's spawninfo list through edit; an empty
// result removes the key.
func editSpawns(path string, edit func([]any) ([]any, error)) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("fight rats: %w", err)
	}
	var room yaml.MapSlice
	if err := yaml.Unmarshal(b, &room); err != nil {
		return fmt.Errorf("fight rats: %s: %w", path, err)
	}
	out := yaml.MapSlice{}
	var spawns []any
	for _, item := range room {
		if item.Key == "spawninfo" {
			list, _ := item.Value.([]any)
			spawns = list
			continue
		}
		out = append(out, item)
	}
	spawns, err = edit(spawns)
	if err != nil {
		return fmt.Errorf("fight rats: %s: %w", path, err)
	}
	if len(spawns) > 0 {
		out = append(out, yaml.MapItem{Key: "spawninfo", Value: spawns})
	}
	b, err = yaml.Marshal(out)
	if err != nil {
		return fmt.Errorf("fight rats: %s: %w", path, err)
	}
	return os.WriteFile(path, b, 0o644)
}

// mobIdOf is a spawn entry's mobid, or 0.
func mobIdOf(spawn any) int {
	if m, ok := spawn.(yaml.MapSlice); ok {
		for _, item := range m {
			if item.Key == "mobid" {
				if n, ok := item.Value.(int); ok {
					return n
				}
			}
		}
	}
	return 0
}

// serverBinary is where launch mode builds the server. The path is fixed so
// repeated runs reuse one file instead of leaving a new exe per run.
func serverBinary() (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	bin := filepath.Join(cache, "geomud-smoke", "geomud-server")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	return bin, nil
}

func launch(ctx context.Context) (_ *server, err error) {
	if _, err := os.Stat(filepath.Join("_datafiles", "config.yaml")); err != nil {
		return nil, errors.New("launch mode must run from the repo root (no _datafiles/config.yaml here)")
	}
	dir, err := os.MkdirTemp("", "geomud-smoke-")
	if err != nil {
		return nil, err
	}
	s := &server{dir: dir}
	defer func() {
		if err != nil {
			s.stop(true)
		}
	}()

	world := filepath.Join(dir, "world")
	if err := os.CopyFS(world, os.DirFS(filepath.Join("_datafiles", "world", "empty"))); err != nil {
		return s, fmt.Errorf("copy world: %w", err)
	}
	if err := fightRats(world); err != nil {
		return s, err
	}

	// Build and run the binary directly, so the PID we hold is the server
	// itself (go run would put a parent process in between).
	bin, err := serverBinary()
	if err != nil {
		return s, err
	}
	build := exec.CommandContext(ctx, "go", "build", "-o", bin, ".")
	build.Stdout, build.Stderr = os.Stderr, os.Stderr
	if err := build.Run(); err != nil {
		return s, fmt.Errorf("build server: %w", err)
	}

	port, err := freePort()
	if err != nil {
		return s, err
	}
	s.addr = net.JoinHostPort("127.0.0.1", strconv.Itoa(port))
	ovr := filepath.Join(dir, "config-overrides.yaml")
	if err := os.WriteFile(ovr, []byte(overrides(world, port)), 0o644); err != nil {
		return s, err
	}

	log, err := os.Create(s.logPath())
	if err != nil {
		return s, err
	}
	defer log.Close() // the child holds its own handle

	s.cmd = exec.Command(bin)
	s.cmd.Env = append(os.Environ(), "CONFIG_PATH="+ovr, "LOG_NOCOLOR=1")
	s.cmd.Stdout, s.cmd.Stderr = log, log
	if err := s.cmd.Start(); err != nil {
		return s, fmt.Errorf("start server: %w", err)
	}
	s.done = make(chan struct{})
	go func() { s.cmd.Wait(); close(s.done) }()

	return s, s.waitReady(ctx, time.Minute)
}

func (s *server) logPath() string { return filepath.Join(s.dir, "server.log") }

// adoptCopyoverChild waits for the server copyover started to log its start
// (it inherits the log file) and records its PID so stop kills it: the
// process this program started exits during copyover.
func (s *server) adoptCopyoverChild(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		b, _ := os.ReadFile(s.logPath())
		if pid := lastStartPid(string(b)); pid != 0 && pid != s.cmd.Process.Pid {
			s.children = append(s.children, pid)
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("copyover: no new server start in the log after %s", timeout)
		}
		select {
		case <-time.After(200 * time.Millisecond):
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// lastStartPid returns the pid of the last "Starting server" log line, or 0.
func lastStartPid(log string) int {
	pid := 0
	for _, line := range strings.Split(log, "\n") {
		if !strings.Contains(line, `msg="Starting server`) {
			continue
		}
		if i := strings.Index(line, " pid="); i >= 0 {
			if f := strings.Fields(line[i+len(" pid="):]); len(f) > 0 {
				if n, err := strconv.Atoi(f[0]); err == nil {
					pid = n
				}
			}
		}
	}
	return pid
}

// waitReady dials the telnet port until it answers.
func (s *server) waitReady(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		conn, err := net.DialTimeout("tcp", s.addr, time.Second)
		if err == nil {
			return conn.Close()
		}
		select {
		case <-s.done:
			return fmt.Errorf("server exited before listening on %s", s.addr)
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("server not listening on %s after %s", s.addr, timeout)
		}
	}
}

// stop kills the server this program started (by its PID, never by name or
// port), optionally prints the end of its log, and deletes the temp dir.
func (s *server) stop(showLog bool) {
	if s.cmd != nil && s.cmd.Process != nil {
		s.cmd.Process.Kill()
		<-s.done
		// A copyover child started before a failed step is not in children yet.
		if b, err := os.ReadFile(s.logPath()); err == nil {
			if pid := lastStartPid(string(b)); pid != 0 && pid != s.cmd.Process.Pid && !slices.Contains(s.children, pid) {
				s.children = append(s.children, pid)
			}
		}
	}
	for _, pid := range s.children {
		if p, err := os.FindProcess(pid); err == nil {
			p.Kill()
		}
	}
	if s.copyLogTo != "" {
		if b, err := os.ReadFile(s.logPath()); err == nil {
			if err := os.WriteFile(s.copyLogTo, b, 0o644); err != nil {
				fmt.Fprintln(os.Stderr, "smoke: could not save the server log:", err)
			}
		}
	}
	if showLog {
		if b, err := os.ReadFile(s.logPath()); err == nil {
			fmt.Fprintf(os.Stderr, "--- end of server log ---\n%s\n---\n", lastLines(string(b), 30))
		}
	}
	// Windows can hold a just-killed process's files open for a moment.
	for i := 0; i < 20; i++ {
		if err := os.RemoveAll(s.dir); err == nil {
			return
		}
		time.Sleep(250 * time.Millisecond)
	}
	fmt.Fprintf(os.Stderr, "smoke: could not delete %s\n", s.dir)
}

func freePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}
