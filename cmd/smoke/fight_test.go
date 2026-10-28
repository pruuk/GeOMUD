package main

import "testing"

func TestCheckObserverText(t *testing.T) {
	const f = "Smkabc0"
	cases := []struct {
		name, got string
		wantErr   bool
	}{
		{"observer lines", "Smkabc0 prepares to fight rat.\n[☀️ 4:01AM HP:6/6 MP:5/5]:Smkabc0 hits rat with their fists.\nrat has died.\n", false},
		{"swing after a blank prompt line", "[☀️ 4:01AM HP:6/6 MP:5/5]:\nSmkabc0 swings at rat, but misses.\nrat has died.\n", false},
		{"possessive name", "[☀️ 4:01AM HP:6/6 MP:5/5]:Smkabc0 prepares to fight rat.\n[☀️ 4:01AM HP:6/6 MP:5/5]:Smkabc0's fists connects with rat.\nrat has died.\n", false},
		{"only the prepare line", "Smkabc0 prepares to fight rat.\nrat has died.\n", true},
		{"attacker line", "Smkabc0 hits rat with their fists.\n[☀️ 4:01AM HP:6/6 MP:5/5]:You miss with your fists.\nYou swing your fists at rat.\n", true},
		{"attacker prepare", "Smkabc0 hits rat.\nYou prepare to fight rat!\n", true},
		{"xp line", "Smkabc0 hits rat.\nYou gained 44 experience points! (combat)\n", true},
		{"unrelated you line", "Smkabc0 hits rat.\nYou say, \"hi\"\n", false},
	}
	for _, c := range cases {
		err := checkObserverText(c.got, f)
		if (err != nil) != c.wantErr {
			t.Errorf("%s: err = %v, want error %v", c.name, err, c.wantErr)
		}
	}
}
