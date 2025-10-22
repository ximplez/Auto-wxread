package github

import "testing"

func Test_encryptSecret(t *testing.T) {
	type args struct {
		publicKey   string
		publicKeyId string
		secretValue string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test",
			args: args{
				publicKey:   "",
				publicKeyId: "",
				secretValue: "{}",
			},
			want: "R/B0qD8UjDnlIduMxjJ+cbohfI14ENTi5nKXGAVHaTH4r1Z28+G//JCtFl0RZFC5RAo=",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := encryptSecretWithPublicKey(tt.args.secretValue, tt.args.publicKey, tt.args.publicKeyId); err != nil {
				t.Errorf("encryptSecret() error = %v", err)
				return
			} else if got != tt.want {
				t.Errorf("encryptSecret() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_createOrUpdateGithubSecret(t *testing.T) {
	type args struct {
		githubToken string
		repo        string
		secretName  string
		secretValue string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "test",
			args: args{
				githubToken: "",
				repo:        "ximplez/Auto-wxread",
				secretName:  "test",
				secretValue: "{}",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CreateOrUpdateGithubSecret(tt.args.githubToken, tt.args.repo, tt.args.secretName, tt.args.secretValue); (err != nil) != tt.wantErr {
				t.Errorf("createOrUpdateGithubSecret() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
