package env_utils

import "github.com/ximplez-go/gf/os/genv"

func GetEnv(key string) string {
	return genv.Get(key, "").String()
}
