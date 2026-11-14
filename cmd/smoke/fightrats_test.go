package main

import (
	"os"
	"path/filepath"
	"testing"

	"gopkg.in/yaml.v2"
)

// fightRats (ruling 11) gives the world copy's Town Square exactly two rats,
// small at level 1 and large at level 5, neither wandering, and takes the
// shipped rats out of End of the Line; the shipped files are not touched.
func TestFightRatsPinsTwoRatsInTownSquare(t *testing.T) {
	shipped := filepath.Join("..", "..", "_datafiles", "world", "empty", "rooms", "startland")
	world := t.TempDir()
	dir := filepath.Join(world, "rooms", "startland")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	originals := map[string][]byte{}
	for _, f := range []string{"1.yaml", "2.yaml"} {
		b, err := os.ReadFile(filepath.Join(shipped, f))
		if err != nil {
			t.Fatal(err)
		}
		originals[f] = b
		if err := os.WriteFile(filepath.Join(dir, f), b, 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := fightRats(world); err != nil {
		t.Fatal(err)
	}

	type spawn struct {
		MobId        int      `yaml:"mobid"`
		Name         string   `yaml:"name"`
		Level        int      `yaml:"level"`
		IdleCommands []string `yaml:"idlecommands"`
	}
	type room struct {
		RoomId    int     `yaml:"roomid"`
		Title     string  `yaml:"title"`
		SpawnInfo []spawn `yaml:"spawninfo"`
	}
	read := func(f string) room {
		var r room
		b, err := os.ReadFile(filepath.Join(dir, f))
		if err != nil {
			t.Fatal(err)
		}
		if err := yaml.Unmarshal(b, &r); err != nil {
			t.Fatal(err)
		}
		return r
	}

	square := read("1.yaml")
	if square.RoomId != 1 || square.Title != startRoomTitle {
		t.Errorf("room 1 lost its fields: %+v", square)
	}
	rats := []spawn{}
	for _, sp := range square.SpawnInfo {
		if sp.MobId == ratMobId {
			rats = append(rats, sp)
		}
	}
	if len(rats) != 2 {
		t.Fatalf("Town Square spawns %d rats, want 2: %+v", len(rats), square.SpawnInfo)
	}
	for i, want := range []struct {
		name  string
		level int
	}{{"small rat", 1}, {"large rat", 5}} {
		got := rats[i]
		if got.Name != want.name || got.Level != want.level || len(got.IdleCommands) != 1 || got.IdleCommands[0] != "" {
			t.Errorf("rat %d is %+v, want %s at level %d with one empty idle command", i, got, want.name, want.level)
		}
		if launchRats[i].name != want.name {
			t.Errorf("launchRats[%d] is %q, want %q (the fight step attacks what fightRats spawns)", i, launchRats[i].name, want.name)
		}
	}
	if n := len(read("2.yaml").SpawnInfo); n != 0 {
		t.Errorf("End of the Line still spawns %d mobs, want 0", n)
	}
	for f, b := range originals {
		again, _ := os.ReadFile(filepath.Join(shipped, f))
		if string(again) != string(b) {
			t.Errorf("the shipped %s changed", f)
		}
	}
}
