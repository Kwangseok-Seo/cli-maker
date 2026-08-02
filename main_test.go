package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kwangseok-Seo/cli-maker/internal/manifest"
	"gopkg.in/yaml.v3"
)

// servers 가 절대 URL 이고 securityScheme 이 있는 최소 spec. 후자는 경고를 하나
// 만들기 위한 것이다 — stdout/stderr 가 갈리는지 보려면 stderr 에 나갈 것이 있어야 한다.
const specAbsolute = `
openapi: "3.0.0"
servers: [{url: "https://api.example.com/v1"}]
components:
  securitySchemes:
    tok: {type: http}
paths:
  /ping:
    get:
      operationId: ping
`

// petstore 와 같은 모양 — servers 가 상대 URL 이라 --base-url 이 없으면 못 옮긴다.
const specRelative = `
openapi: "3.0.0"
servers: [{url: "/v1"}]
paths:
  /ping:
    get:
      operationId: ping
`

// writeSpec 은 spec 을 임시 파일로 떨어뜨리고 경로를 준다.
// t.TempDir 은 테스트가 끝나면 통째로 지워지므로 뒷정리가 필요 없다.
func writeSpec(t *testing.T, body string) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "spec.yaml")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// runImport 는 import 명령을 한 번 실행하고 두 스트림을 갈라 돌려준다.
//
// newImportCmd 를 매번 새로 부르는 것이 요점이다 — flag 를 담는 변수가 그 안에서
// 태어나므로, 한 프로세스에서 여러 번 돌려도 앞 실행의 --out 값이 남지 않는다.
func runImport(t *testing.T, args ...string) (stdout, stderr string, err error) {
	t.Helper()

	cmd := newImportCmd()
	var out, errOut bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&errOut)
	// 에러 문구는 err 로 직접 보므로, cobra 가 stderr 에 또 찍지 않게 한다.
	cmd.SilenceErrors = true
	cmd.SetArgs(args)

	err = cmd.Execute()
	return out.String(), errOut.String(), err
}

// import 산출물은 stdout 으로만 나가야 한다.
// `cli-maker import spec.json --name x > apis/x.yaml` 이 성립하려면, 경고 한 줄이라도
// stdout 에 섞이는 순간 파일이 깨진다.
func TestImportKeepsWarningsOffStdout(t *testing.T) {
	stdout, stderr, err := runImport(t, writeSpec(t, specAbsolute), "--name", "svc")
	if err != nil {
		t.Fatalf("import: %v (stderr=%q)", err, stderr)
	}

	if strings.Contains(stdout, "securityScheme") {
		t.Errorf("경고가 stdout 에 섞였다:\n%s", stdout)
	}
	if !strings.Contains(stderr, "securityScheme") {
		t.Errorf("경고가 stderr 에 없다: %q", stderr)
	}

	// stdout 이 그대로 매니페스트로 읽혀야 한다 — 그게 이 명령의 계약이다.
	var m manifest.Manifest
	if err := yaml.Unmarshal([]byte(stdout), &m); err != nil {
		t.Fatalf("stdout 을 매니페스트로 못 읽었다: %v\n%s", err, stdout)
	}
	if m.Name != "svc" || m.BaseURL != "https://api.example.com/v1" || len(m.Commands) != 1 {
		t.Errorf("매니페스트 = %+v", m)
	}
}

func TestImportName(t *testing.T) {
	tests := []struct {
		name    string
		args    []string // spec 경로 뒤에 붙는 것들. {{out}} 은 임시 파일 경로로 바뀐다.
		want    string
		wantErr string
	}{
		{
			name: "--name 을 주면 그것",
			args: []string{"--name", "chosen", "--out", "{{out}}"},
			want: "chosen",
		},
		{
			// --out 이 apis/pstore.yaml 이면 CLI 에서 `cli-maker pstore ...` 로 부르게 된다.
			name: "--name 이 없으면 --out 파일명에서 가져온다",
			args: []string{"--out", "{{out}}"},
			want: "petstore",
		},
		{
			name: "--name 이 --out 을 이긴다",
			args: []string{"--name", "chosen", "--out", "{{out}}"},
			want: "chosen",
		},
		{
			// stdout 으로 낼 때는 파일명이 없으므로 물어보는 수밖에 없다.
			name:    "둘 다 없으면 에러",
			args:    nil,
			wantErr: "--name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outPath := filepath.Join(t.TempDir(), "petstore.yaml")

			args := []string{writeSpec(t, specAbsolute)}
			for _, a := range tt.args {
				args = append(args, strings.ReplaceAll(a, "{{out}}", outPath))
			}

			_, _, err := runImport(t, args...)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("에러를 기대했는데 통과했다")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("에러 = %q, want %q 포함", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("import: %v", err)
			}

			// 쓰인 파일을 런타임과 같은 경로로 다시 읽는다 — 산출물이 실제로
			// 등록 가능한지는 Load+Validate 가 통과해야 아는 것이다.
			m, err := manifest.Load(outPath)
			if err != nil {
				t.Fatal(err)
			}
			if err := manifest.Validate(m); err != nil {
				t.Fatalf("산출물이 검증을 통과하지 못한다: %v", err)
			}
			if m.Name != tt.want {
				t.Errorf("name = %q, want %q", m.Name, tt.want)
			}
		})
	}
}

// 이 파일은 유저가 손으로 이어 쓰는 초안이다. 두 번째 import 가 편집분을 조용히
// 지우면 되돌릴 방법이 없다.
func TestImportRefusesToOverwrite(t *testing.T) {
	spec := writeSpec(t, specAbsolute)
	outPath := filepath.Join(t.TempDir(), "svc.yaml")

	if _, _, err := runImport(t, spec, "--out", outPath); err != nil {
		t.Fatalf("첫 실행: %v", err)
	}

	// 유저가 손으로 고친 상태를 흉내 낸다.
	const edited = "name: svc\n# 손으로 적은 줄\n"
	if err := os.WriteFile(outPath, []byte(edited), 0o644); err != nil {
		t.Fatal(err)
	}

	_, _, err := runImport(t, spec, "--out", outPath)
	if err == nil {
		t.Fatal("덮어썼다")
	}
	if !strings.Contains(err.Error(), "이미 있다") {
		t.Errorf("에러 = %q", err)
	}

	// 거절만으로는 부족하다 — 파일이 그대로 남아 있어야 한다.
	got, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != edited {
		t.Errorf("파일이 바뀌었다:\n%s", got)
	}
}

func TestImportRejects(t *testing.T) {
	tests := []struct {
		name    string
		spec    string
		args    []string
		wantErr string
	}{
		{
			name:    "servers 가 상대 URL 인데 --base-url 이 없다",
			spec:    specRelative,
			args:    []string{"--name", "svc"},
			wantErr: "--base-url",
		},
		{
			name:    "Swagger 2.0",
			spec:    "swagger: \"2.0\"\npaths:\n  /ping:\n    get:\n      operationId: ping\n",
			args:    []string{"--name", "svc", "--base-url", "https://x.example.com"},
			wantErr: "2.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := append([]string{writeSpec(t, tt.spec)}, tt.args...)
			if _, _, err := runImport(t, args...); err == nil {
				t.Fatal("통과했다")
			} else if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("에러 = %q, want %q 포함", err, tt.wantErr)
			}
		})
	}
}

// --base-url 이 상대 URL 문제를 실제로 푸는지. 위 테스트는 없을 때만 봤다.
func TestImportBaseURLOverride(t *testing.T) {
	stdout, _, err := runImport(t,
		writeSpec(t, specRelative), "--name", "svc", "--base-url", "https://api.example.com/v1")
	if err != nil {
		t.Fatalf("import: %v", err)
	}
	if !strings.Contains(stdout, "baseUrl: https://api.example.com/v1") {
		t.Errorf("baseUrl 이 안 들어갔다:\n%s", stdout)
	}
}

func TestImportMissingFile(t *testing.T) {
	if _, _, err := runImport(t, filepath.Join(t.TempDir(), "없다.json"), "--name", "svc"); err == nil {
		t.Fatal("없는 파일이 통과했다")
	}
}
