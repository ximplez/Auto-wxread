package json_tool

import (
	"log"

	"github.com/ximplez-go/gf/encoding/gjson"
)

func JsonPrintln(data interface{}, pretty bool) {
	log.Println(ToJson(data, pretty))
}

func ToJson(data interface{}, pretty bool) string {
	if pretty {
		return gjson.ToJsonSilentPretty(data)
	} else {
		return gjson.ToJsonSilent(data)
	}
}

func PhaseJson[T any](data []byte) *T {
	return gjson.PhaseJsonSilent[T](data)
}

func PhaseJsonFromString[T any](data string) *T {
	return PhaseJson[T]([]byte(data))
}
