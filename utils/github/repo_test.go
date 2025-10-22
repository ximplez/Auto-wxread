package github

import "testing"

func Test_getGithubRepoPubKey(t *testing.T) {
	type args struct {
		githubToken string
		repo        string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		want1   string
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				githubToken: "",
				repo:        "ximplez/Auto-wxread",
			},
			want:    "",
			want1:   "",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := getGithubRepoPubKey(tt.args.githubToken, tt.args.repo)
			if (err != nil) != tt.wantErr {
				t.Errorf("getGithubRepoPubKey() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getGithubRepoPubKey() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getGithubRepoPubKey() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}
