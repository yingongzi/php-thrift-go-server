package util

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	jsoniter "github.com/json-iterator/go"
)

var (
	Json = jsoniter.ConfigCompatibleWithStandardLibrary
)

func JsonString(v interface{}) string {
	Json, _ := Json.Marshal(v)
	return string(Json)
}

func JsonUnmarshalFromString(jsonStr string, v interface{}) error {
	err := Json.UnmarshalFromString(jsonStr, v)
	if err != nil {
		return json.Unmarshal([]byte(jsonStr), v)
	}
	return err
}

func Md5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

func InSlice(i string, s []string) bool {
	for _, v := range s {
		if v == i {
			return true
		}
	}

	return false
}