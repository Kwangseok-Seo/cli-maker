package manifest

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var allowedMethods = map[string]bool{
	"GET": true, "POST": true, "PUT": true, "PATCH": true,
	"DELETE": true, "HEAD": true, "OPTIONS": true,
}

// 실행기가 실제로 처리하는 자리만 허용한다. header/body 는 아직 없으므로 받아 두고 무시하지 않는다.
var allowedParamIn = map[string]bool{"path": true, "query": true}

// SupportsParamIn 은 실행기가 처리하는 param 자리인지 답한다.
//
// 임포터(internal/openapi)가 "이 param 을 옮길 수 있나"를 우리에게 묻기 위한 문이다.
// 목록을 저쪽에 복제하면, 여기가 header 를 받게 되는 날 임포터만 조용히 뒤처진다.
func SupportsParamIn(in string) bool {
	return allowedParamIn[in]
}

// Validate 는 매니페스트 하나만 보고 판정할 수 있는 문제를 모두 찾아 한꺼번에 돌려준다.
// 하나 찾을 때마다 반환하지 않는 이유는, 유저가 매니페스트를 고치러 여러 번 왕복하지 않게 하기 위해서다.
//
// 여기 없는 것: 다른 매니페스트나 CLI 전역 표면을 알아야 하는 검사
// (매니페스트 name 충돌, 예약된 flag 이름). 그건 internal/cli 의 몫 — 6b.
func Validate(m *Manifest) error {
	var errs []error

	if m.Name == "" {
		errs = append(errs, errors.New("name 이 비어 있다"))
	}

	if m.BaseURL == "" {
		errs = append(errs, errors.New("baseURL 이 비어 있다"))
	} else {
		u, err := url.Parse(m.BaseURL)
		if err != nil {
			errs = append(errs, err)
		} else {
			if u.Scheme != "https" && u.Scheme != "http" {
				errs = append(errs, fmt.Errorf("baseURL %q: scheme 이 http/https 가 아니다", m.BaseURL))
			}
		}
	}

	if len(m.Commands) == 0 {
		errs = append(errs, errors.New("Command 가 비어 있다"))
	}

	seen := map[string]bool{}
	for i, c := range m.Commands {
		if c.Name == "" {
			errs = append(errs, fmt.Errorf("commands[%d]: name 이 비어 있다", i))
			continue
		}
		if seen[c.Name] {
			errs = append(errs, fmt.Errorf("commands[%d] %q: 이름이 중복이다", i, c.Name))
		}
		seen[c.Name] = true

		if !allowedMethods[c.Method] {
			errs = append(errs, fmt.Errorf("commands[%d] %q: method %q 는 지원하지 않는다", i, c.Name, c.Method))
		}

		// param 이름은 command 마다 따로 유일하면 된다 — 그래서 seenParam 은 바깥 루프 몸통 안에서 태어난다.
		seenParam := map[string]bool{}
		for j, p := range c.Params {
			if p.Name == "" {
				errs = append(errs, fmt.Errorf("commands[%d] %q: params[%d] 의 name 이 비어 있다", i, c.Name, j))
				continue
			}
			if seenParam[p.Name] {
				errs = append(errs, fmt.Errorf("commands[%d] %q: param %q 가 중복이다", i, c.Name, p.Name))
			}
			seenParam[p.Name] = true

			if !allowedParamIn[p.In] {
				errs = append(errs, fmt.Errorf("commands[%d] %q: param %q 의 in %q 는 지원하지 않는다", i, c.Name, p.Name, p.In))
			}
			// path param 이 required 가 아니면, 유저가 안 줬을 때 빈 문자열이 치환돼 엉뚱한 URL 이 나간다.
			if p.In == "path" && !p.Required {
				errs = append(errs, fmt.Errorf("commands[%d] %q: path param %q 는 required 여야 한다", i, c.Name, p.Name))
			}
		}

		// BuildURL 이 실제로 하는 치환을 그대로 흉내 낸다 — 아는 자리를 다 지우고 {가 남으면 대응이 없다.
		left := c.Path
		for _, p := range c.Params {
			if p.In == "path" {
				left = strings.ReplaceAll(left, "{"+p.Name+"}", "")
			}
		}
		if strings.Contains(left, "{") {
			errs = append(errs, fmt.Errorf("commands[%d] %q: path %q 에 대응 param 이 없는 자리표시자가 있다", i, c.Name, c.Path))
		}
	}

	return errors.Join(errs...)
}
