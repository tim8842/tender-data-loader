package pkg

import "encoding/json"

func UnmarshalVars[T any](vars map[string]interface{}) (T, error) {
	var result T
	bytes, err := json.Marshal(vars)
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(bytes, &result)
	return result, err
}
