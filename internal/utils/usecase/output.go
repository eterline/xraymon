package usecase

import (
	"encoding/json"
	"fmt"
)

func PrintJSON(v any, pretty bool) {
	var data []byte
	if !pretty {
		data, _ = json.Marshal(v)
	} else {
		data, _ = json.MarshalIndent(v, "", "    ")
	}
	fmt.Println(string(data))
}
