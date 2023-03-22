package backend

import (
	"github.com/imroc/req/v3"
	"github.com/tidwall/gjson"
)

type Response struct {
	*req.Response
}

func (resp Response) GetJsonStringField(path string) string {
	result := gjson.Get(resp.String(), path)
	return result.String()
}
