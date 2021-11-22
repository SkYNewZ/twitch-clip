package player

import (
	"reflect"
	"testing"
)

func TestDefaultPlayer(t *testing.T) {
	tests := []struct {
		name    string
		want    Player
		wantErr bool
	}{
		{
			name:    "Expected VLC",
			want:    VLC,
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
