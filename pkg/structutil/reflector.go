package structutil

import "encoding/json"

func MapToStruct[T any](val map[string]any) (res T) {
	res = *new(T)
	b, _ := json.Marshal(val)
	json.Unmarshal(b, &res)

	return
}
