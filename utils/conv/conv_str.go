package conv

import (
	"strconv"

	"github.com/ximplez/wxread/utils/json_tool"
)

func Str[T any](v T) string {
	return json_tool.ToJson(v, false)
}

func StrPtr[T any](v T) *string {
	str := Str(v)
	return &str
}

func Str2Int64(s string) int64 {
	parseInt, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return parseInt
}

func Str2Int64Ptr(s string) *int64 {
	parseInt, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return Ptr(int64(0))
	}
	return Ptr(parseInt)
}
