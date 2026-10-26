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
	"strconv"
	"time"
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
