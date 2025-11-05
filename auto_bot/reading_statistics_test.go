package main

import (
	"context"
	"log"
	"testing"

	"github.com/ximplez/wxread/utils/json_tool"
)

func TestReadingStatisticService_GetReadingStatistic(t *testing.T) {
	s := NewReadingStatisticService()
	type args struct {
		ctx  context.Context
		vid  string
		skey string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				ctx:  context.Background(),
				vid:  "",
				skey: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.GetReadingStatistic(tt.args.ctx, tt.args.vid, tt.args.skey)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetReadingStatistic() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			log.Printf("got: %v", json_tool.ToJson(got, true))
		})
	}
}
