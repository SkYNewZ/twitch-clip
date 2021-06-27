// +build linux

package player

import (
	"bytes"
	"io"
	"reflect"
	"strings"
	"testing"
)

func TestDefaultPlayer(t *testing.T) {
	tests := []struct {
		name    string
		want    Player
		wantErr bool
	}{
		{
			name:    "Expected MPV",
			want:    MPV,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DefaultPlayer()
			if (err != nil) != tt.wantErr {
				t.Errorf("DefaultPlayer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DefaultPlayer() got = %v, want %v", got, tt.want)
			}
		})
	}
}
func Test_player_Run(t *testing.T) {
	type fields struct {
		name       string
		command    []string
		registry   string
		registry32 string
	}

	// Read command output
	var buff bytes.Buffer

	type args struct {
		u      string
		title  string
		output io.Writer
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
		want    string // read command output
	}{
		{
			name: "Expected with $url",
			fields: fields{
				name:       "Echo",
				command:    []string{"echo", "$url"},
				registry:   "",
				registry32: "",
			},
			args: args{
				u:      "world",
				title:  "",
				output: &buff,
			},
			wantErr: false,
			want:    "world",
		},
		{
			name: "Expected with $title",
			fields: fields{
				name:       "Echo",
				command:    []string{"echo", "$title"},
				registry:   "",
				registry32: "",
			},
			args: args{
				u:      "",
				title:  "world",
				output: &buff,
			},
			wantErr: false,
			want:    "world",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer buff.Reset() //empty buffer after each tests
			p := &player{
				name:       tt.fields.name,
				command:    tt.fields.command,
				registry:   tt.fields.registry,
				registry32: tt.fields.registry32,
			}

			if err := p.Run(tt.args.u, tt.args.title, tt.args.output); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			got := strings.TrimSpace(buff.String())
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Run() got = %v, want %v", got, tt.want)
			}
		})
	}
}
