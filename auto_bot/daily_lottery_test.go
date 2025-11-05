package main

import (
	"context"
	"log"
	"testing"

	"github.com/ximplez/wxread/utils/json_tool"
)

func TestDailyLotteryService_ExecuteDailyLottery(t *testing.T) {
	s := NewDailyLotteryService()
	type args struct {
		ctx   context.Context
		vid   string
		skey  string
		issue string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				ctx:   context.Background(),
				vid:   "",
				skey:  "",
				issue: "",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := s.ExecuteDailyLottery(tt.args.ctx, tt.args.vid, tt.args.skey, tt.args.issue)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteDailyLottery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			log.Printf("got: %v", json_tool.ToJson(got, true))
		})
	}
}

func TestDailyLotteryService_GetDailyLottery(t *testing.T) {
	s := NewDailyLotteryService()
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

			got, err := s.GetDailyLottery(tt.args.ctx, tt.args.vid, tt.args.skey)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDailyLottery() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			log.Printf("got: %v", json_tool.ToJson(got, true))
		})
	}
}
