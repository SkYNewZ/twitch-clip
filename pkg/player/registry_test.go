// +build windows

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
			name: "VLC",
			fields: fields{
				name:       VLC.(*player).name,
				registry:   VLC.(*player).registry,
				registry32: VLC.(*player).registry32,
				command:    VLC.(*player).command,
			},
			want: true,
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
