package executor

import (
	"net/url"
	"strings"

	"github.com/Kwangseok-Seo/cli-maker/internal/manifest"
)

// BuildURL 은 Command 와 유저가 flag 로 준 값(values)으로 최종 요청 URL 을 만든다.
//
//	m.BaseURL = "https://petstore3.swagger.io/api/v3"
//	c.Path    = "/pet/{petId}"
//	values    = map[string]string{"petId": "10"}
//	→ "https://petstore3.swagger.io/api/v3/pet/10"
//
// param 의 In 이 "path" 면 경로의 {이름} 자리에 치환하고, "query" 면 ?a=b 로 붙인다.
// 값이 빈 문자열인 param 은 유저가 주지 않은 것이므로 query 에 싣지 않는다.
func BuildURL(m *manifest.Manifest, c *manifest.Command, values map[string]string) string {
	path := c.Path
	q := url.Values{}
	for _, p := range c.Params {
		v := values[p.Name]
		switch p.In {
		case "path":
			path = strings.ReplaceAll(path, "{"+p.Name+"}", url.PathEscape(v))
		case "query":
			if v != "" {
				q.Set(p.Name, v)
			}
		}
	}

	if qs := q.Encode(); qs != "" {
		path += "?" + qs
	}

	return m.BaseURL + path
}
