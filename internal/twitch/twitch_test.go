package twitch

import (
	"net/http"
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	type args struct {
		config *Config
	}
	tests := []struct {
		name    string
		args    args
		want    *Client
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.args.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() got = %v, want %v", got, tt.want)
			}
		})
	}
}

var _ http.RoundTripper = (*roundTripper)(nil)

type roundTripper struct{}

func (s roundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, nil
}

func Test_transport_RoundTrip(t *testing.T) {
	transport := &transport{
		clientID: "foo",
		Original: new(roundTripper),
	}

	tests := []struct {
		name string
		want string
	}{
		{
			name: "Client-Id",
			want: "foo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, "http://foo.bar", nil)
			_, _ = transport.RoundTrip(req) //nolint:bodyclose

			if req.Header.Get(tt.name) != tt.want {
				t.Errorf("RoundTrip() headers[%q] = %q, want %q", tt.name, req.Header.Get(tt.name), tt.want)
			}
		})
	}
}
