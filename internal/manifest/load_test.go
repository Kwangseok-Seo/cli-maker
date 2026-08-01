package manifest

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLoad 는 파일 → struct 경로만 판정한다.
//
// 의미 검증은 Validate 의 몫이므로(ADR-0007 의 두 층) 여기서 묻는 것은 셋뿐이다 —
// 읽혔는가, 어떻게 실패하는가, 그리고 yaml 태그가 실제로 먹었는가.
//
// 입력 파일은 서브테스트마다 t.TempDir() 로 새로 만든다. testdata/ 로 빼지 않는
// 이유는 매니페스트가 몇 줄뿐이라 케이스 옆에 두는 편이 읽히고, 앞 케이스의 파일이
// 다음 케이스로 새지 않기 때문이다 — validate_test.go 가 매번 새 struct 를 만드는
// 것과 같은 이유다.
func TestLoad(t *testing.T) {
	tests := []struct {
		name string
		// body 는 파일에 써 넣을 YAML.
		body string
		// missing 이면 파일을 아예 만들지 않는다.
		missing bool
		// wantIs 는 errors.Is 로 물어야 하는 sentinel.
		// OS 가 쓴 문구("The system cannot find the file specified.")는 플랫폼마다
		// 달라 단언할 수 없으므로, 문자열 대신 이걸로 묻는다.
		wantIs error
		// wantErr 는 에러 메시지에 있어야 할 조각. OS 가 쓴 문구는 넣지 않는다.
		wantErr string
		// check 는 성공했을 때 struct 가 어떻게 채워졌는지 본다.
		check func(*testing.T, *Manifest)
	}{
		{
			name: "정상 매니페스트가 struct 로 채워진다",
			body: `
name: gh
baseUrl: https://api.github.com
auth:
  type: bearer
  env: GITHUB_TOKEN
commands:
  - name: repo
    method: GET
    path: /repos/{owner}/{repo}
    params:
      - name: owner
        in: path
        type: string
        required: true
`,
			check: func(t *testing.T, m *Manifest) {
				if m.Name != "gh" {
					t.Errorf("Name = %q, want %q", m.Name, "gh")
				}
				// 이 한 줄이 struct 태그를 지킨다 — 필드는 BaseURL, 파일 키는 baseUrl 이다.
				if m.BaseURL != "https://api.github.com" {
					t.Errorf("BaseURL = %q — yaml:\"baseUrl\" 태그가 안 먹었다", m.BaseURL)
				}
				if m.Auth.Env != "GITHUB_TOKEN" {
					t.Errorf("Auth.Env = %q, want %q", m.Auth.Env, "GITHUB_TOKEN")
				}
				if len(m.Commands) != 1 {
					t.Fatalf("Commands %d개, want 1개", len(m.Commands))
				}
				if p := m.Commands[0].Params; len(p) != 1 || !p[0].Required {
					t.Errorf("Params = %+v — required: true 가 bool 로 안 왔다", p)
				}
			},
		},
		{
			// body: 는 M10 에서 생긴 자리다. 이 케이스가 잠그는 것은 필드가 exported 라는 것 —
			// 소문자 body 였다면 리플렉션이 쓸 수 없어 yaml 이 조용히 건너뛰고, 파싱은 성공한
			// 채로 세 단언이 모두 깨진다. 태그를 붙여도 권한이 생기지는 않는다.
			name: "body 가 포인터로 채워진다",
			body: `
name: pstore
baseUrl: https://petstore3.swagger.io/api/v3
commands:
  - name: addPet
    method: POST
    path: /pet
    body:
      required: true
      contentType: application/json
`,
			check: func(t *testing.T, m *Manifest) {
				if len(m.Commands) != 1 {
					t.Fatalf("Commands %d개, want 1개", len(m.Commands))
				}
				b := m.Commands[0].Body
				// 뒤 두 줄이 이 줄의 성립을 전제하므로 Errorf 가 아니라 Fatal 이다.
				// nil 인 채로 b.Required 를 읽으면 실패가 아니라 panic 이고,
				// 그러면 이 서브테스트만이 아니라 나머지 결과까지 함께 날아간다.
				if b == nil {
					t.Fatal("Body = nil — body: 를 적었는데 안 채워졌다")
				}
				if !b.Required {
					t.Errorf("Body.Required = false, want true")
				}
				if b.ContentType != "application/json" {
					t.Errorf("Body.ContentType = %q, want %q", b.ContentType, "application/json")
				}
			},
		},
		{
			// nil 이 "이 명령은 본문을 받지 않는다"를 뜻한다. 값 struct 였다면 body: 를 안 적은
			// 명령과 body: {} 를 적은 명령이 파싱 후에 구별되지 않고, 그러면 --data 를 모든
			// 명령에 등록할 수밖에 없어 GET 명령이 본문 flag 를 달게 된다.
			name: "body 를 안 적으면 nil 이다",
			body: `
name: pstore
baseUrl: https://petstore3.swagger.io/api/v3
commands:
  - name: getPetById
    method: GET
    path: /pet/{petId}
`,
			check: func(t *testing.T, m *Manifest) {
				if len(m.Commands) != 1 {
					t.Fatalf("Commands %d개, want 1개", len(m.Commands))
				}
				if b := m.Commands[0].Body; b != nil {
					t.Errorf("Body = &%+v, want nil", *b)
				}
			},
		},
		{
			// LoadDir 이 이 sentinel 로 "apis/ 가 없는 건 잘못이 아니다"를 가르므로,
			// Load 가 이걸 감싸서 돌려주는 성질이 위층 분기의 전제다.
			name:    "없는 파일은 fs.ErrNotExist 로 온다",
			missing: true,
			wantIs:  fs.ErrNotExist,
		},
		{
			name:    "닫히지 않은 리스트는 파싱 에러다",
			body:    "name: [unclosed",
			wantErr: "yaml:",
		},
		{
			// 문법은 맞는데 자리가 틀린 경우. yaml 이 목표 타입을 알기 때문에 잡힌다.
			name:    "commands 자리에 스칼라가 오면 잡힌다",
			body:    "commands: notalist",
			wantErr: "cannot unmarshal",
		},
		{
			// 태그는 baseUrl 인데 파일 키가 baseurl 이라 매칭되지 않는다. yaml.v3 은
			// 모르는 키를 조용히 버리므로 Load 는 통과하고, 빈 BaseURL 이 Validate 까지
			// 내려가 "baseURL 이 비어 있다"로 보고된다 — 오타 자리를 짚어 주지는 않는다.
			name: "대소문자가 다른 키는 조용히 버려진다",
			body: "name: gh\nbaseurl: https://api.github.com\n",
			check: func(t *testing.T, m *Manifest) {
				if m.Name != "gh" {
					t.Errorf("Name = %q, want %q", m.Name, "gh")
				}
				if m.BaseURL != "" {
					t.Errorf(`BaseURL = %q, want "" — 태그 매칭 규칙이 바뀌었다`, m.BaseURL)
				}
			},
		},
		{
			// 빈 파일은 Load 의 실패가 아니다 — 읽을 것이 없었을 뿐이다.
			// 판정은 Validate 로 넘어간다(name·baseURL·Command 셋을 한꺼번에 잡는다).
			name: "빈 파일은 에러 없이 zero value 가 된다",
			body: "",
			check: func(t *testing.T, m *Manifest) {
				if m.Name != "" || m.BaseURL != "" || len(m.Commands) != 0 {
					t.Errorf("zero value 가 아니다: %+v", *m)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 서브테스트마다 새 디렉토리. 지우는 코드는 우리가 쓰지 않는다 —
			// t.TempDir 이 t.Cleanup 에 등록해 두고 테스트가 끝날 때 지운다.
			path := filepath.Join(t.TempDir(), "m.yaml")

			if !tt.missing {
				// 준비가 실패한 것은 판정이 아니므로 Errorf 가 아니라 Fatal 이다.
				if err := os.WriteFile(path, []byte(tt.body), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			m, err := Load(path)

			if tt.wantIs != nil || tt.wantErr != "" {
				if err == nil {
					t.Fatalf("Load() = %+v, want 에러", m)
				}
				if m != nil {
					t.Errorf("에러인데 매니페스트가 왔다: %+v", m)
				}
				if tt.wantIs != nil && !errors.Is(err, tt.wantIs) {
					t.Errorf("errors.Is(err, %v) 가 아니다: %v", tt.wantIs, err)
				}
				if tt.wantErr != "" && !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("에러에 %q 가 없다: %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("Load() 에러: %v", err)
			}
			if tt.check != nil {
				tt.check(t, m)
			}
		})
	}
}
