package main

import (
	"context"
	"log"
	"testing"

	"github.com/ximplez/wxread/utils/json_tool"
)

func TestReadingRewardService_ExchangeReadingReward(t *testing.T) {
	s := NewReadingRewardService()
	type args struct {
		ctx        context.Context
		vid        string
		skey       string
		levelId    int
		choiceType int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				ctx:        context.Background(),
				vid:        "",
				skey:       "",
				levelId:    2,
				choiceType: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := s.ExchangeReadingReward(tt.args.ctx, tt.args.vid, tt.args.skey, tt.args.levelId, tt.args.choiceType); (err != nil) != tt.wantErr {
				t.Errorf("ExchangeReadingReward() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestReadingRewardService_GetReadingReward(t *testing.T) {
	s := NewReadingRewardService()
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
			got, err := s.GetReadingReward(tt.args.ctx, tt.args.vid, tt.args.skey)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetReadingReward() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			log.Printf("got: %v", json_tool.ToJson(got, true))
		})
	}
}
