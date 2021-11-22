//go:build !windows
// +build !windows

package player

import "testing"

func Test_player_checkRegistry(t *testing.T) {
	type fields struct {
		name       string
		command    []string
		registry   string
		registry32 string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "Expected",
			fields: fields{
				name:       "Foo",
				command:    []string{"foo", "bar"},
				registry:   "foo",
				registry32: "bar",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &player{
				name:       tt.fields.name,
				command:    tt.fields.command,
				registry:   tt.fields.registry,
				registry32: tt.fields.registry32,
			}
			if got := p.checkRegistry(); got != tt.want {
				t.Errorf("checkRegistry() = %v, want %v", got, tt.want)
			}
		})
	}
}
