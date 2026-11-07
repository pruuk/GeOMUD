package buffs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetDurations(t *testing.T) {
	type args struct {
		buff *Buff
		spec *BuffSpec
	}
	tests := []struct {
		name       string
		args       args
		wantRounds int
		wantTotal  int
	}{
		{
			// Fresh buff, RoundCounter=0: all triggers remain, next trigger in RoundInterval rounds
			name: "Normal case, fresh buff",
			args: args{
				buff: &Buff{TriggersLeft: 5, TriggersInitial: 5, RoundCounter: 0},
				spec: &BuffSpec{TriggerCount: 5, RoundInterval: 3},
			},
			wantRounds: 15, // (5-1)*3 + (3-0) = 12 + 3 = 15
			wantTotal:  15,
		},
		{
			// Mid-interval: RoundCounter=2, RoundInterval=3 -> 1 round until next trigger
			name: "Mid-interval",
			args: args{
				buff: &Buff{TriggersLeft: 5, TriggersInitial: 5, RoundCounter: 2},
				spec: &BuffSpec{TriggerCount: 5, RoundInterval: 3},
			},
			wantRounds: 13, // (5-1)*3 + (3-2) = 12 + 1 = 13
			wantTotal:  15,
		},
		{
			// Just triggered: RoundCounter is a multiple of RoundInterval
			name: "Just triggered",
			args: args{
				buff: &Buff{TriggersLeft: 4, TriggersInitial: 5, RoundCounter: 3},
				spec: &BuffSpec{TriggerCount: 5, RoundInterval: 3},
			},
			wantRounds: 12, // (4-1)*3 + (3-0) = 9 + 3 = 12
			wantTotal:  15,
		},
		{
			name: "One trigger left",
			args: args{
				buff: &Buff{TriggersLeft: 1, TriggersInitial: 4, RoundCounter: 0},
				spec: &BuffSpec{TriggerCount: 4, RoundInterval: 2},
			},
			wantRounds: 2, // (1-1)*2 + (2-0) = 0 + 2 = 2
			wantTotal:  8,
		},
		{
			name: "Zero triggers left",
			args: args{
				buff: &Buff{TriggersLeft: 0, TriggersInitial: 5, RoundCounter: 0},
				spec: &BuffSpec{TriggerCount: 5, RoundInterval: 5},
			},
			wantRounds: 0,
			wantTotal:  25,
		},
		{
			name: "Zero round interval",
			args: args{
				buff: &Buff{TriggersLeft: 4, TriggersInitial: 4, RoundCounter: 0},
				spec: &BuffSpec{TriggerCount: 4, RoundInterval: 0},
			},
			wantRounds: 0,
			wantTotal:  0,
		},
		{
			// TriggersLeft overridden beyond TriggerCount: totalRounds uses TriggersInitial
			name: "Overridden trigger count",
			args: args{
				buff: &Buff{TriggersLeft: 80, TriggersInitial: 80, RoundCounter: 0},
				spec: &BuffSpec{TriggerCount: 20, RoundInterval: 1},
			},
			wantRounds: 80, // (80-1)*1 + (1-0) = 79 + 1 = 80
			wantTotal:  80, // uses TriggersInitial=80, not spec.TriggerCount=20
		},
		{
			// Old save with no TriggersInitial: falls back to spec.TriggerCount
			name: "Legacy buff, no TriggersInitial",
			args: args{
				buff: &Buff{TriggersLeft: 3, TriggersInitial: 0, RoundCounter: 0},
				spec: &BuffSpec{TriggerCount: 5, RoundInterval: 2},
			},
			wantRounds: 6,  // (3-1)*2 + (2-0) = 4 + 2 = 6
			wantTotal:  10, // falls back to spec.TriggerCount=5 * 2
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			gotRounds, gotTotal := GetDurations(tt.args.buff, tt.args.spec)
			assert.Equal(t, tt.wantRounds, gotRounds)
			assert.Equal(t, tt.wantTotal, gotTotal)
		})
	}
}
func TestBuffs_HasBuff(t *testing.T) {
	type fields struct {
		list    []*Buff
		buffIds map[int]int
	}
	tests := []struct {
		name   string
		fields fields
		arg    int
		want   bool
	}{
		{
			name: "Buff exists in buffIds",
			fields: fields{
				list: []*Buff{
					{BuffId: 1},
					{BuffId: 2},
				},
				buffIds: map[int]int{1: 0, 2: 1},
			},
			arg:  1,
			want: true,
		},
		{
			name: "Buff does not exist in buffIds",
			fields: fields{
				list: []*Buff{
					{BuffId: 1},
					{BuffId: 2},
				},
				buffIds: map[int]int{1: 0, 2: 1},
			},
			arg:  3,
			want: false,
		},
		{
			name: "Empty buffIds map",
			fields: fields{
				list:    []*Buff{},
				buffIds: map[int]int{},
			},
			arg:  1,
			want: false,
		},
		{
			name: "BuffIds map is nil",
			fields: fields{
				list:    []*Buff{},
				buffIds: nil,
			},
			arg:  1,
			want: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			bs := &Buffs{
				List:    tt.fields.list,
				buffIds: tt.fields.buffIds,
			}
			assert.Equal(t, tt.want, bs.HasBuff(tt.arg))
		})
	}
}
func TestBuffs_RemoveBuff(t *testing.T) {
	type fields struct {
		list    []*Buff
		buffIds map[int]int
	}
	tests := []struct {
		name         string
		fields       fields
		arg          int
		wantResult   bool
		wantTriggers int
		shouldModify bool
	}{
		{
			name: "Buff exists and is removed",
			fields: fields{
				list: []*Buff{
					{BuffId: 1, TriggersLeft: 5},
					{BuffId: 2, TriggersLeft: 3},
				},
				buffIds: map[int]int{1: 0, 2: 1},
			},
			arg:          2,
			wantResult:   true,
			wantTriggers: TriggersLeftExpired,
			shouldModify: true,
		},
		{
			name: "Buff does not exist",
			fields: fields{
				list: []*Buff{
					{BuffId: 1, TriggersLeft: 5},
				},
				buffIds: map[int]int{1: 0},
			},
			arg:          3,
			wantResult:   false,
			wantTriggers: 5,
			shouldModify: false,
		},
		{
			name: "buffIds is nil",
			fields: fields{
				list:    []*Buff{{BuffId: 1, TriggersLeft: 7}},
				buffIds: nil,
			},
			arg:          1,
			wantResult:   false,
			wantTriggers: 7,
			shouldModify: false,
		},
		{
			name: "buffIds is empty map",
			fields: fields{
				list:    []*Buff{{BuffId: 1, TriggersLeft: 7}},
				buffIds: map[int]int{},
			},
			arg:          1,
			wantResult:   false,
			wantTriggers: 7,
			shouldModify: false,
		},
		{
			name: "Multiple buffs, remove first",
			fields: fields{
				list: []*Buff{
					{BuffId: 10, TriggersLeft: 2},
					{BuffId: 20, TriggersLeft: 4},
				},
				buffIds: map[int]int{10: 0, 20: 1},
			},
			arg:          10,
			wantResult:   true,
			wantTriggers: TriggersLeftExpired,
			shouldModify: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			bs := &Buffs{
				List:    tt.fields.list,
				buffIds: tt.fields.buffIds,
			}
			result := bs.RemoveBuff(tt.arg)
			assert.Equal(t, tt.wantResult, result)
			if idx, ok := bs.buffIds[tt.arg]; ok && tt.shouldModify {
				assert.Equal(t, tt.wantTriggers, bs.List[idx].TriggersLeft)
			} else if len(bs.List) > 0 {
				// If not modified, original value should remain
				assert.Equal(t, tt.wantTriggers, bs.List[0].TriggersLeft)
			}
		})
	}
}
func TestBuff_Expired(t *testing.T) {
	tests := []struct {
		name         string
		triggersLeft int
		want         bool
	}{
		{
			name:         "TriggersLeft is zero (expired)",
			triggersLeft: 0,
			want:         true,
		},
		{
			name:         "TriggersLeft is negative (expired)",
			triggersLeft: -1,
			want:         true,
		},
		{
			name:         "TriggersLeft is positive (not expired)",
			triggersLeft: 3,
			want:         false,
		},
		{
			name:         "TriggersLeft is TriggersLeftUnlimited (not expired)",
			triggersLeft: TriggersLeftUnlimited,
			want:         false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			b := &Buff{TriggersLeft: tt.triggersLeft}
			assert.Equal(t, tt.want, b.Expired())
		})
	}
}
