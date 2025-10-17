package json_tool

import (
	"testing"
)

func TestFromJsonToMap(t *testing.T) {
	type args struct {
		data string
	}
	type testCase[T any] struct {
		name string
		args args
		want *T
	}
	tests := []testCase[map[string]any]{
		{
			name: "test",
			args: args{
				data: `{"id":1111111111111111111,"time":2}`,
			},
			want: &map[string]any{
				"id":   int64(1111111111111111111),
				"time": int64(2),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

		})
	}
}
