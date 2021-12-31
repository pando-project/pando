package api

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"pando/pkg/api/types"
	"path"
	"strings"
)

var Client *resty.Client
var PandoAPIBaseURL string

func NewClient(apiBaseURL string) {
	Client = resty.New().SetBaseURL(apiBaseURL).SetDebug(false)
}

func PrintResponseData(res *resty.Response) error {
	resJson := types.ResponseJson{}
	err := json.Unmarshal(res.Body(), &resJson)
	if err != nil {
		return err
	}
	prettyJson, err := json.MarshalIndent(resJson, "", " ")
	if err != nil {
		return err
	}
	fmt.Printf("%s\n", prettyJson)
	return nil
}

func JoinPathFuncFactory(groupPath string) func(subPath ...string) string {
	return func(subPath ...string) string {
		fullPath := append([]string{groupPath}, subPath...)

		strBuilder := strings.Builder{}
		strBuilder.WriteString(PandoAPIBaseURL)
		strBuilder.WriteString(path.Join(fullPath...))

		return strBuilder.String()
	}
}
