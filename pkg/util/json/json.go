package json

import (
	"bytes"
	"encoding/json"
)

func PrettyJsonStr(origin interface{}) (string, error) {
	originBytes, err := json.Marshal(origin)
	if err != nil {
		return "", err
	}

	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, originBytes, "", "    "); err != nil {
		return "", err
	}

	return prettyJSON.String(), nil
}
