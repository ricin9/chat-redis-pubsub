package utils

import (
	"encoding/json"
	"fmt"
)

func Dump(data interface{}) {
	b, _ := json.MarshalIndent(data, "", "  ")
	fmt.Print(string(b))
}
