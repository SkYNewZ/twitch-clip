package twitch

import (
	"reflect"
	"testing"
)

func Test_streamsClient_GetStreams(t *testing.T) {
	type fields struct {
		c *Client
	}
	type args struct {
		userLogin []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*Stream
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &streamsClient{
				c: tt.fields.c,
			}
			got, err := s.GetStream(tt.args.userLogin...)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetStream() got = %v, want %v", got, tt.want)
			}
		})
	}
}
