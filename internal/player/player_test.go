package player

import (
	"testing"
)

func Test_player_Name(t *testing.T) {
	type fields struct {
		name       string
		command    []string
		registry   string
		registry32 string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Expected",
			fields: fields{
				name:       "Foo",
				command:    []string{"hello", "foo"},
				registry:   "foo",
				registry32: "bar",
			},
			want: "Foo",
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
			if got := p.Name(); got != tt.want {
				t.Errorf("Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_player_checkIfExist(t *testing.T) {
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &player{
				name:       tt.fields.name,
				command:    tt.fields.command,
				registry:   tt.fields.registry,
				registry32: tt.fields.registry32,
			}
			if got := p.checkIfExist(); got != tt.want {
				t.Errorf("checkIfExist() = %v, want %v", got, tt.want)
			}
		})
	}
}
