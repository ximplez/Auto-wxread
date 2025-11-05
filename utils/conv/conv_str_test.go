package conv

import (
	"fmt"
	"reflect"
	"testing"
)

type testData struct {
	Id   int64 `json:"id,omitempty"`
	Time int   `json:"time,omitempty"`
}

func TestStr(t *testing.T) {
	type args[T any] struct {
		v T
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want string
	}
	tests := []testCase[testData]{
		{
			name: "test",
			args: args[testData]{
				v: testData{
					Id:   1111111111111111111,
					Time: 2,
				},
			},
			want: "{\"id\":\"1111111111111111111\",\"time\":2}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := fmt.Sprintf("%v", tt.args.v); got != tt.want {
				t.Errorf("Str() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStrPtr(t *testing.T) {
	type args[T any] struct {
		v T
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want *string
	}
	tests := []testCase[testData]{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StrPtr(tt.args.v); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StrPtr() = %v, want %v", got, tt.want)
			}
		})
	}
}
