package main

import (
	"context"
	"os"
	"testing"
)

func Test_all(t *testing.T) {
	vid = os.Getenv("vid")
	skey = os.Getenv("skey")
	feishuBotUrl = os.Getenv("feishuBotUrl")
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "test",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := stat(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("stat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
		t.Run(tt.name, func(t *testing.T) {
			if err := dailyLottery(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("dailyLottery() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
		t.Run(tt.name, func(t *testing.T) {
			if err := readingReward(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("readingReward() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_stat(t *testing.T) {
	vid = os.Getenv("vid")
	skey = os.Getenv("skey")
	feishuBotUrl = os.Getenv("feishuBotUrl")
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "test",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := stat(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("stat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_dailyLottery(t *testing.T) {
	vid = os.Getenv("vid")
	skey = os.Getenv("skey")
	feishuBotUrl = os.Getenv("feishuBotUrl")
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "test",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := dailyLottery(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("dailyLottery() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_readingReward(t *testing.T) {
	vid = os.Getenv("vid")
	skey = os.Getenv("skey")
	feishuBotUrl = os.Getenv("feishuBotUrl")
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "test",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := readingReward(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("readingReward() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
