// Package openapi 는 OpenAPI Spec(외부 표준 명세)을 읽는다.
//
// CONTEXT.md 의 구분을 그대로 지킨다 — Spec 은 남이 정한 형식이고 Manifest 는 우리
// 형식이다. 이 패키지는 Spec 을 읽기만 하고, Manifest 로 옮기는 일은 옆 파일이 한다.
// 런타임은 Spec 을 절대 읽지 않는다(ADR-0001).
package openapi

import (
	"maps"
	"slices"
)

// Spec 은 OpenAPI 문서 전체가 아니라 **우리가 읽는 자리만** 담는다.
//
// 여기 칸이 없는 키(info·tags·components·responses…)는 언마샬이 조용히 버린다.
// 그게 이 struct 의 요점이자 함정이다 — 우리가 yaml 태그를 오타 내도 똑같이 조용히
// 버려지고, 해당 필드가 zero value 로 남는 것 말고는 알 방법이 없다. 그래서 이
// 패키지의 테스트는 "에러가 안 났다"가 아니라 "실제로 무엇이 담겼나"를 본다.
type Spec struct {
	OpenAPI string `yaml:"openapi"`
	// Swagger 2.0 문서는 이 칸에만 버전을 적는다. 우리는 2.0 을 지원하지 않지만
	// **거절하려고** 읽는다 — 안 읽으면 OpenAPI 가 빈 문자열인 것만 보이고,
	// "OpenAPI 문서가 아니다"라는 틀린 진단이 나간다.
	Swagger    string              `yaml:"swagger"`
	Servers    []Server            `yaml:"servers"`
	Paths      map[string]PathItem `yaml:"paths"`
	Components Components          `yaml:"components"`
}

type Server struct {
	URL string `yaml:"url"`
}

// Components 에서 우리가 보는 건 securitySchemes 하나다. 옮기려고가 아니라
// **못 옮긴다고 알리려고** 본다 — 인증이 걸린 API 를 인증 없는 매니페스트로 내면서
// 아무 말도 안 하면, 유저는 401 을 받고서야 알게 된다.
type Components struct {
	SecuritySchemes map[string]SecurityScheme `yaml:"securitySchemes"`
}

type SecurityScheme struct {
	Type string `yaml:"type"`
}

// PathItem 은 한 경로에 달린 메서드들이다.
//
// Spec 에선 이것도 map(키가 메서드 이름)이지만 struct 로 받는다. 키의 가짓수가 HTTP
// 메서드로 정해져 있어서 미리 칸을 팔 수 있고, 그러면 순회 순서가 **필드 선언 순서**로
// 고정돼 결정론이 공짜로 따라온다. 무순서 문제가 남는 곳은 Paths 하나뿐이 된다.
type PathItem struct {
	Get     *Operation `yaml:"get"`
	Post    *Operation `yaml:"post"`
	Put     *Operation `yaml:"put"`
	Patch   *Operation `yaml:"patch"`
	Delete  *Operation `yaml:"delete"`
	Head    *Operation `yaml:"head"`
	Options *Operation `yaml:"options"`
}

// Operation 은 한 (경로, 메서드) 쌍의 명세다.
//
// RequestBody 가 포인터인 이유는 Manifest 의 Body 와 같다 — "본문이 아예 없다"와
// "본문이 있는데 required 가 아니다"를 갈라야 하고, bool 하나로는 그게 안 된다
// (ADR-0010). petstore 의 uploadFile 이 정확히 후자다.
type Operation struct {
	OperationID string       `yaml:"operationId"`
	Parameters  []Parameter  `yaml:"parameters"`
	RequestBody *RequestBody `yaml:"requestBody"`
}

type Parameter struct {
	Name     string `yaml:"name"`
	In       string `yaml:"in"`
	Required bool   `yaml:"required"`
	Schema   Schema `yaml:"schema"`
}

// Schema 에서 우리가 쓰는 건 type 하나뿐이다 — Param.Type 으로 옮겨가 flag 의
// usage 문구가 된다. 그 외(format·items·$ref…)는 쓸 자리가 없으므로 칸을 파지 않는다.
type Schema struct {
	Type string `yaml:"type"`
}

// RequestBody 의 content 는 미디어 타입이 키인 map 이다. 값(스키마)은 쓰지 않고
// 키만 보므로 값 타입을 빈 struct 로 둔다 — "여기는 안 읽는다"를 타입으로 적은 것.
//
// 키가 여럿일 수 있다(petstore 의 addPet 은 json/xml/x-www-form-urlencoded 셋).
// 우리 Body.ContentType 은 하나뿐이라 고르는 일이 생기는데, 그건 11b 의 몫이다.
type RequestBody struct {
	Required bool                `yaml:"required"`
	Content  map[string]struct{} `yaml:"content"`
}

// Op 은 순회 결과 한 건이다.
//
// Operation 만으로는 어느 경로·어느 메서드였는지 알 수 없다 — Spec 에선 그 둘이
// 바깥 map 의 키였기 때문이다. 순회하면서 되붙여 준다.
type Op struct {
	Path      string
	Method    string // 대문자. Manifest 의 method 표기와 같은 자리로 간다.
	Operation *Operation
}

// methods 는 이 PathItem 에 실제로 있는 (메서드, Operation) 쌍을 선언 순서대로
// 돌려준다. nil 인 칸은 그 경로에 그 메서드가 없다는 뜻이라 건너뛴다.
//
// Path 는 채우지 않는다 — PathItem 은 자기가 어느 경로에 달렸는지 모른다.
// 채우는 건 바깥 키를 아는 Operations 의 몫이다.
func (p PathItem) methods() []Op {
	pairs := []struct {
		name string
		op   *Operation
	}{
		{"GET", p.Get}, {"POST", p.Post}, {"PUT", p.Put}, {"PATCH", p.Patch},
		{"DELETE", p.Delete}, {"HEAD", p.Head}, {"OPTIONS", p.Options},
	}

	var out []Op
	for _, pr := range pairs {
		if pr.op != nil {
			out = append(out, Op{Method: pr.name, Operation: pr.op})
		}
	}
	return out
}

// Operations 는 spec 의 모든 operation 을 **결정론적 순서로** 돌려준다.
//
// 같은 Spec 을 몇 번 넣든 같은 순서가 나와야 한다. 그러지 않으면 import 를 두 번
// 돌렸을 때 매니페스트의 명령 순서가 달라지고, diff 가 매번 통째로 뜬다.
//
// 순서: 경로를 사전순으로, 같은 경로 안에서는 methods 의 선언 순서로.
// (외부 Spec 의 저자 순서는 보존하지 않는다 — 11a 에서 내린 결정.)
func (s *Spec) Operations() []Op {
	var out []Op

	keys := slices.Sorted(maps.Keys(s.Paths))

	for _, key := range keys {
		item := s.Paths[key]

		for _, op := range item.methods() {
			// methods 는 Path 를 채우지 않으므로 늘 비어 있다 — 검사할 것이 없다.
			op.Path = key
			out = append(out, op)
		}
	}
	return out
}
