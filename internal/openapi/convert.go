package openapi

import (
	"fmt"
	"maps"
	"net/url"
	"slices"
	"strings"

	"github.com/Kwangseok-Seo/cli-maker/internal/manifest"
)

// 의존 방향은 openapi → manifest 한 쪽뿐이다. 임포터는 목적지 형식을 알아야 하지만,
// manifest 는 자기가 어디서 왔는지 몰라야 한다 — 알게 되면 런타임이 OpenAPI 를
// 아는 셈이 되고 ADR-0001 의 "매니페스트가 데이터의 전부"가 깨진다.

// supportedMajor 는 우리가 읽을 수 있는 OpenAPI 메이저 버전이다.
//
// 판정과 에러 문구가 이 하나를 함께 쓴다. 리터럴로 두면 지원 범위를 넓힐 때
// 판정만 고치고 문구는 "3.x 뿐"이라고 계속 광고하는 드리프트가 난다.
//
// 끝의 점이 일한다 — "3" 만 보면 "30.1" 같은 값도 통과한다.
const supportedMajor = "3."

// preferredContentType 은 후보가 여럿일 때 먼저 고르는 미디어 타입이다.
// 이 CLI 는 JSON 쪽으로 기울어 있다 — 응답 포맷(-o pretty/compact)이 JSON 을 안다.
const preferredContentType = "application/json"

// pickContentType 은 requestBody 의 content 후보 중 하나를 결정론적으로 고르고,
// 임의로 고른 것이면 그 사실을 함께 돌려준다(경고가 없으면 빈 문자열).
//
// content 는 map 이라 순서가 없다 — 11a 의 Paths 에 이은 두 번째 무순서 지점이다.
// 여기서도 같은 방식으로 없앤다: 사전순. 다만 그 앞에 선호를 하나 둔다.
func pickContentType(content map[string]struct{}) (string, string) {
	if len(content) == 0 {
		return "", ""
	}
	if _, ok := content[preferredContentType]; ok {
		return preferredContentType, ""
	}

	keys := slices.Sorted(maps.Keys(content))
	if len(keys) == 1 {
		return keys[0], ""
	}
	// json 이 없는데 후보가 여럿 — 고른 근거가 사전순뿐이므로 유저에게 알린다.
	return keys[0], fmt.Sprintf("content 후보 %v 중 %q 를 골랐다", keys, keys[0])
}

// toCommand 는 Op 하나를 Command 하나로 옮긴다.
//
// 돌려주는 값 셋이 서로 다른 것을 뜻한다:
//
//	err != nil       → 이 operation 은 옮길 수 없다. 통째로 뺀다.
//	warnings != nil  → 옮겼지만 일부를 잃었다. 명령은 살아 있다.
//	cmd              → 옮겨진 결과
//
// 에러와 경고를 가르는 기준은 하나다 — **잃은 것이 required 인가**.
// 필수 입력을 잃은 명령은 반드시 잘못된 요청을 보내므로 살려 두면 안 된다.
// 선택 입력을 잃은 명령은 나머지가 정확하므로 살린다.
func toCommand(op Op) (cmd *manifest.Command, warnings []string, err error) {
	// 이름이 없으면 명령을 지을 수 없다. 이 에러엔 operationId 를 쓸 수 없으므로
	// 유저가 스펙에서 찾아갈 수 있는 좌표(메서드+경로)를 대신 넣는다.
	if op.Operation.OperationID == "" {
		return nil, nil, fmt.Errorf("%s %s: operationId 가 없다 — 명령 이름을 지을 수 없다", op.Method, op.Path)
	}

	var params []manifest.Param
	for _, p := range op.Operation.Parameters {
		if !manifest.SupportsParamIn(p.In) {
			if p.Required {
				return nil, nil, fmt.Errorf("%s: 필수 param %q 의 in %q 를 옮길 수 없다",
					op.Operation.OperationID, p.Name, p.In)
			}
			warnings = append(warnings, fmt.Sprintf("%s: param %q 를 뺐다 — in %q 는 지원하지 않는다",
				op.Operation.OperationID, p.Name, p.In))
			continue
		}
		params = append(params, manifest.Param{
			Name:     p.Name,
			In:       p.In,
			Type:     p.Schema.Type, // Spec 에선 타입이 한 겹 안쪽(schema)에 있다
			Required: p.Required,
		})
	}

	// 본문은 있을 수도, 없을 수도 있다. 없으면 포인터를 nil 로 남긴다 —
	// 여기서 &manifest.Body{} 를 지어 버리면 --data 가 붙지 말아야 할 명령에 붙는다 (ADR-0010).
	var body *manifest.Body
	if rb := op.Operation.RequestBody; rb != nil {
		ct, warn := pickContentType(rb.Content)
		if warn != "" {
			warnings = append(warnings, op.Operation.OperationID+": "+warn)
		}
		body = &manifest.Body{Required: rb.Required, ContentType: ct}
	}

	return &manifest.Command{
		Name:   op.Operation.OperationID,
		Method: op.Method,
		Path:   op.Path,
		Params: params,
		Body:   body,
	}, warnings, nil
}

// checkVersion 은 이 문서를 우리가 읽을 수 있는지 판정한다.
//
//	통과 — openapi 가 "3." 으로 시작한다
//	거절 — swagger 칸이 차 있다 (2.0). 이름을 대고 거절한다.
//	거절 — 그 외 (OpenAPI 문서가 아니거나 우리가 모르는 버전)
//
// 이 게이트가 없으면 2.0 문서가 **경고 한 줄 없이 절반만** 옮겨진다. paths·
// operationId·in:path 는 3.0 과 이름이 같아 그대로 넘어오고, servers 와 param
// 타입만 조용히 빈다 — 실측으로 확인했다.
func checkVersion(s *Spec) error {
	if strings.HasPrefix(s.OpenAPI, supportedMajor) {
		return nil
	}

	// 값이 틀린 게 아니라 맞는 문서를 우리가 못 읽는 것이다 — 그렇게 말해야
	// 유저가 스펙을 고치러 가지 않는다.
	if s.Swagger != "" {
		return fmt.Errorf("Swagger %s 는 아직 지원하지 않는다 — 2.0 은 servers 자리가 host+basePath+schemes 로 흩어져 있어 별개 매핑이 필요하다", s.Swagger)
	}

	// 값을 그대로 보여 준다. 빈 문자열이면 "이 파일엔 openapi 필드가 없다"는 뜻이라,
	// spec 이 아닌 파일을 넣은 경우와 모르는 버전을 넣은 경우가 갈린다.
	return fmt.Errorf("openapi 필드가 %q 다 — 읽을 수 있는 것은 %sx 뿐이다", s.OpenAPI, supportedMajor)
}

// resolveBaseURL 은 매니페스트에 적을 baseUrl 을 정한다.
//
// 우선순위는 M5 에서 배운 결과 같다 — 좁은 수명이 넓은 것을 이긴다.
// 여기서 묻는 것은 "유효한가"가 아니라 "절대 URL 인가" 하나뿐이다.
// http/https 인지는 manifest.Validate 의 질문이므로 여기서 다시 묻지 않는다.
func resolveBaseURL(s *Spec, override string) (string, error) {
	if override != "" {
		return override, nil
	}

	if len(s.Servers) == 0 || s.Servers[0].URL == "" {
		return "", fmt.Errorf("spec 에 servers 가 없다 — --base-url 로 절대 URL 을 주어야 한다")
	}

	raw := s.Servers[0].URL
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("servers[0].url %q 를 읽을 수 없다: %w", raw, err)
	}
	if !u.IsAbs() {
		return "", fmt.Errorf("servers[0].url %q 는 상대 URL 이다 — --base-url 로 절대 URL 을 주어야 한다", raw)
	}
	return raw, nil
}

// Convert 는 Spec 하나를 Manifest 하나로 옮긴다.
//
// name 과 baseURL 을 바깥에서 받는 이유는 Spec 에서 못 얻기 때문이다 — info.title 은
// 정규화해도 명령 이름이 못 되고("swagger-petstore-openapi-3-0"), servers 는 상대
// URL 일 수 있다(petstore 가 그렇다).
//
// 옮기지 못한 operation·param 은 warnings 로 나가고, 매니페스트 자체가 성립하지
// 않을 때만 error 다. warnings 는 error 와 함께 돌아올 수도 있다 — 실패해도 무엇을
// 잃었는지는 알려 주는 편이 유저에게 낫다.
func Convert(s *Spec, name, baseURL string) (*manifest.Manifest, []string, error) {
	if err := checkVersion(s); err != nil {
		return nil, nil, err
	}

	base, err := resolveBaseURL(s, baseURL)
	if err != nil {
		return nil, nil, err
	}

	var warnings []string

	// 인증은 옮기지 않는다 — securitySchemes 의 어느 타입도 {type: bearer, env} 에
	// 대응되지 않는다. 그래도 있었다는 사실은 알린다. 아무 말 없이 인증 없는
	// 매니페스트를 내면 유저는 401 을 받고서야 알게 된다.
	if len(s.Components.SecuritySchemes) > 0 {
		// 세 번째 무순서 지점이다 — 여기도 같은 방식으로 없앤다.
		names := slices.Sorted(maps.Keys(s.Components.SecuritySchemes))
		warnings = append(warnings, fmt.Sprintf(
			"securityScheme %v 를 옮기지 않았다 — auth 는 손으로 적어야 한다", names))
	}

	var cmds []manifest.Command
	for _, op := range s.Operations() {
		cmd, w, err := toCommand(op)
		if err != nil {
			warnings = append(warnings, "건너뜀: "+err.Error())
			continue
		}
		warnings = append(warnings, w...)
		cmds = append(cmds, *cmd)
	}

	m := &manifest.Manifest{Name: name, BaseURL: base, Commands: cmds}

	// 런타임 등록과 같은 검증을 통과해야 한다. 여기서 막지 않으면 apis/ 에 떨어진
	// 뒤에야 거절당하고, 그때는 무엇이 잘못됐는지 spec 까지 되짚어야 한다.
	if err := manifest.Validate(m); err != nil {
		return nil, warnings, err
	}
	return m, warnings, nil
}
