package openapi

import (
	"os"

	"gopkg.in/yaml.v3"
)

// Load 는 OpenAPI Spec 파일을 읽는다.
//
// .json 이든 .yaml 이든 같은 함수로 읽힌다 — YAML 1.2 가 JSON 의 상위집합이라
// yaml.Unmarshal 이 둘 다 받는다. 그래서 encoding/json 을 따로 들이지 않는다.
//
// 여기서 걸러지는 것은 "파일을 못 읽음"과 "문법이 깨짐"뿐이다. 우리가 읽을 수 있는
// 문서인지는 Convert 가 판정한다 — 부분 디코드는 모르는 키를 조용히 버리므로,
// 언마샬이 통과했다는 사실 자체는 아무것도 보장하지 않는다.
func Load(path string) (*Spec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var s Spec
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}
