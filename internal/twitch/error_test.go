package twitch

import "testing"

func TestError_Error(t *testing.T) {
	type fields struct {
		Err     string
		Status  int
		Message string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "expected",
			fields: fields{
				Err:     "foo",
				Status:  1234,
				Message: "bar",
			},
			want: "twitch error 1234 foo: bar",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := Error{
				Err:     tt.fields.Err,
				Status:  tt.fields.Status,
				Message: tt.fields.Message,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
