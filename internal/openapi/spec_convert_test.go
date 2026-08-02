package openapi

import (
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func parseSpec(t *testing.T, doc string) *Spec {
	t.Helper()

	var s Spec
	if err := yaml.Unmarshal([]byte(doc), &s); err != nil {
		t.Fatal(err)
	}
	return &s
}

const petstoreBase = "https://petstore3.swagger.io/api/v3"

func TestConvertPetstore(t *testing.T) {
	m, warnings, err := Convert(loadPetstore(t), "pstore", petstoreBase)
	if err != nil {
		t.Fatalf("Convert: %v", err)
	}

	if m.Name != "pstore" || m.BaseURL != petstoreBase {
		t.Errorf("name/baseURL = %q/%q, want %q/%q", m.Name, m.BaseURL, "pstore", petstoreBase)
	}
	if len(m.Commands) != 19 {
		t.Errorf("commands = %d, want 19", len(m.Commands))
	}
	// 인증은 통째로 못 옮기므로 비어 있어야 한다 — 반쪽으로 채우면 조용히 틀린다.
	if m.Auth.Type != "" || m.Auth.Env != "" {
		t.Errorf("auth = %+v, want 비어 있음", m.Auth)
	}

	// 경고 두 건: securityScheme 미이전 + deletePet 의 header param.
	if len(warnings) != 2 {
		t.Fatalf("경고 = %v, want 2건", warnings)
	}
	if !strings.Contains(warnings[0], "securityScheme") {
		t.Errorf("첫 경고 = %q, want securityScheme 언급", warnings[0])
	}
}

// 버전 게이트가 없으면 2.0 이 경고 없이 절반만 옮겨진다 — 그 실측이 이 테스트의 이유다.
func TestConvertRejectsSwagger2(t *testing.T) {
	doc := `
swagger: "2.0"
host: petstore.swagger.io
basePath: /v2
paths:
  /pet/{petId}:
    get:
      operationId: getPetById
      parameters: [{name: petId, in: path, required: true, type: integer}]
`
	_, _, err := Convert(parseSpec(t, doc), "p", "https://example.com")
	if err == nil {
		t.Fatal("2.0 문서가 통과했다 — 반쪽 매니페스트가 나간다")
	}
	// "OpenAPI 문서가 아니다"가 아니라 2.0 이라고 이름을 대야 유저가 다음 수를 안다.
	if !strings.Contains(err.Error(), "2.0") {
		t.Errorf("에러 = %q, want 2.0 을 이름으로 지목", err)
	}
}

func TestConvertRejectsUnknownVersion(t *testing.T) {
	// 문서 나머지는 멀쩡하다 — 버전 말고는 거절할 구실이 없어야 이 테스트가 뜻을 갖는다.
	// (처음엔 paths 를 비워 뒀다가, 버전 게이트가 아니라 "Command 가 비어 있다"로
	//  통과하는 공허한 테스트인 걸 발견하고 고쳤다.)
	const body = `
paths:
  /x:
    get:
      operationId: getX
`
	tests := []struct{ doc, wantIn string }{
		// openapi 도 swagger 도 없다 — spec 이 아닌 파일을 넣은 경우.
		// 빈 값을 그대로 보여 줘야 "필드가 아예 없다"가 드러난다.
		{"", `""`},
		{`openapi: "4.1"`, "4.1"}, // 우리가 모르는 미래 버전
		{`openapi: "2.0"`, "2.0"}, // openapi 칸에 2.0 을 적은 문서
	}

	for _, tt := range tests {
		_, _, err := Convert(parseSpec(t, tt.doc+body), "p", "https://example.com")
		if err == nil {
			t.Errorf("통과했다: %q", tt.doc)
			continue
		}
		// 유저가 무엇을 넣었는지 되짚을 수 있어야 한다. 값을 안 보여 주면
		// "4.1 을 넣었다"와 "spec 이 아닌 파일을 넣었다"가 같은 문장이 된다.
		if !strings.Contains(err.Error(), tt.wantIn) {
			t.Errorf("에러 = %q, want %s 포함", err, tt.wantIn)
		}
	}
}

func TestResolveBaseURL(t *testing.T) {
	const abs = "https://api.example.com/v1"

	tests := []struct {
		name     string
		servers  string
		override string
		want     string
		wantErr  string // 에러 메시지에 들어 있어야 할 조각
	}{
		{
			name:    "절대 URL 이면 그대로 쓴다",
			servers: "servers: [{url: " + abs + "}]",
			want:    abs,
		},
		{
			name:     "--base-url 이 절대 URL 도 이긴다",
			servers:  "servers: [{url: " + abs + "}]",
			override: "http://localhost:8080",
			want:     "http://localhost:8080",
		},
		{
			// petstore 가 정확히 이 모양이다.
			name:    "상대 URL 이면 거절하고 --base-url 을 안내한다",
			servers: `servers: [{url: /api/v3}]`,
			wantErr: "--base-url",
		},
		{
			name:    "servers 가 없어도 마찬가지",
			servers: "",
			wantErr: "--base-url",
		},
		{
			name:     "servers 가 없어도 --base-url 이 있으면 된다",
			servers:  "",
			override: abs,
			want:     abs,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveBaseURL(parseSpec(t, tt.servers), tt.override)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("baseURL = %q, want 에러", got)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("에러 = %q, want %q 포함", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("에러: %v", err)
			}
			if got != tt.want {
				t.Errorf("baseURL = %q, want %q", got, tt.want)
			}
		})
	}
}
