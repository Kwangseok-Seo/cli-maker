package executor

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Kwangseok-Seo/cli-maker/internal/format"
	"github.com/Kwangseok-Seo/cli-maker/internal/manifest"
)

// testCommand 는 path param 과 query param 을 하나씩 가진 명령이다.
// 이걸로 BuildURL 이 만든 URL 이 실제로 선을 타고 서버까지 가는지 본다.
func testCommand() *manifest.Command {
	return &manifest.Command{
		Name:   "repo",
		Method: "GET",
		Path:   "/repos/{owner}",
		Params: []manifest.Param{
			{Name: "owner", In: "path", Required: true},
			{Name: "sort", In: "query"},
		},
	}
}

// TestExecute 는 응답이 어떻게 나가고 실패가 어떻게 알려지는지를 못 박는다.
//
// 목을 만들지 않는다. httptest 가 진짜 HTTP 서버를 127.0.0.1 의 빈 포트에 띄우고,
// 매니페스트의 BaseURL 을 그 주소로 바꾸면 우리 코드는 평소 경로(http.DefaultClient)
// 그대로 돈다. 바꾸는 것은 목적지뿐이다.
func TestExecute(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		f       format.Formatter
		wantOut string
		// wantErr 는 반환 에러에 들어 있어야 할 조각. "" 면 에러가 없어야 한다.
		wantErr string
	}{
		{
			name:    "200 이면 본문을 그대로 낸다",
			status:  http.StatusOK,
			body:    `{"a":1}`,
			f:       format.Raw{},
			wantOut: `{"a":1}`,
		},
		{
			// Execute 는 f 가 무엇인지 모르는 채 resp.Body 를 넘긴다.
			// 그런데도 결과가 들여써져 나오면 배선이 살아 있다는 뜻.
			name:    "본문은 넘겨받은 포매터를 거쳐 나간다",
			status:  http.StatusOK,
			body:    `{"a":1}`,
			f:       format.Pretty{},
			wantOut: "{\n  \"a\": 1\n}\n",
		},
		{
			// ADR-0005. 순서가 계약이다 — 본문을 버리고 에러만 내면 유저는
			// API 가 뭐라고 했는지 영영 못 본다.
			name:    "404 여도 본문을 먼저 내보낸 뒤에 실패한다",
			status:  http.StatusNotFound,
			body:    `{"message":"Not Found"}`,
			f:       format.Raw{},
			wantOut: `{"message":"Not Found"}`,
			wantErr: "HTTP 404",
		},
		{
			name:    "500 도 같은 규칙을 따른다",
			status:  http.StatusInternalServerError,
			body:    "boom",
			f:       format.Raw{},
			wantOut: "boom",
			wantErr: "HTTP 500",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 서버가 받은 요청을 클로저로 바깥에 붙잡아 둔다 — 이게 목 프레임워크의 자리다.
			var gotReq *http.Request

			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotReq = r
				w.WriteHeader(tt.status)
				io.WriteString(w, tt.body)
			}))
			t.Cleanup(srv.Close)

			m := &manifest.Manifest{Name: "gh", BaseURL: srv.URL}
			values := map[string]string{"owner": "spf13", "sort": "stars"}

			var out, errOut bytes.Buffer
			err := Execute(t.Context(), m, testCommand(), values, tt.f, &out, &errOut)

			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Execute() 에러: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("Execute() = nil, want %q 를 담은 에러", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("에러 = %v, want %q 를 담은 것", err, tt.wantErr)
				}
			}

			if got := out.String(); got != tt.wantOut {
				t.Errorf("out = %q, want %q", got, tt.wantOut)
			}

			// BuildURL 이 만든 것이 그대로 도착했는가.
			if gotReq == nil {
				t.Fatal("서버가 요청을 받지 못했다")
			}
			if gotReq.URL.Path != "/repos/spf13" {
				t.Errorf("path = %q, want %q", gotReq.URL.Path, "/repos/spf13")
			}
			if gotReq.URL.RawQuery != "sort=stars" {
				t.Errorf("query = %q, want %q", gotReq.URL.RawQuery, "sort=stars")
			}
		})
	}
}

// tokenEnv 는 이 테스트만 쓰는 환경변수 이름이다.
const tokenEnv = "CLI_MAKER_TEST_TOKEN"

// TestExecuteAuth 는 ADR-0006 — 토큰을 얻지 못해도 요청은 익명으로 나간다 — 을 못 박는다.
//
// 서버가 받은 Authorization 헤더로 판정한다. 우리 코드를 들여다보지 않고
// 상대편이 무엇을 받았는지로 확인하는 것이 이 테스트의 요점이다.
func TestExecuteAuth(t *testing.T) {
	tests := []struct {
		name       string
		auth       manifest.Auth
		setEnv     bool
		envValue   string
		wantHeader string
		// wantWarn 은 stderr 에 들어 있어야 할 조각. "" 면 조용해야 한다.
		wantWarn string
	}{
		{
			name: "auth 설정이 없으면 헤더도 경고도 없다",
			auth: manifest.Auth{},
		},
		{
			name:       "토큰이 있으면 Bearer 로 붙는다",
			auth:       manifest.Auth{Type: "bearer", Env: tokenEnv},
			setEnv:     true,
			envValue:   "sekret",
			wantHeader: "Bearer sekret",
		},
		{
			name:     "토큰이 없으면 헤더 없이 보내고 경고한다",
			auth:     manifest.Auth{Type: "bearer", Env: tokenEnv},
			setEnv:   false,
			wantWarn: "is not set",
		},
		{
			// 빈 토큰을 붙이면 익명이면 200 을 주던 곳도 401 이 된다 — 그래서 안 붙인다.
			name:     "토큰이 빈 문자열이면 헤더 없이 보내고 경고한다",
			auth:     manifest.Auth{Type: "bearer", Env: tokenEnv},
			setEnv:   true,
			envValue: "",
			wantWarn: "is set but empty",
		},
		{
			name:     "모르는 auth 타입이면 경고하고 익명으로 보낸다",
			auth:     manifest.Auth{Type: "basic", Env: tokenEnv},
			wantWarn: "unknown auth type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tokenEnv, tt.envValue)
			} else {
				// t.Setenv 로 "테스트가 끝나면 원래대로" 를 먼저 등록해 두고 지운다.
				// 개발자 셸에 이 변수가 이미 있어도 이 케이스가 흔들리지 않는다.
				t.Setenv(tokenEnv, "지워질 값")
				os.Unsetenv(tokenEnv)
			}

			var gotAuth string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotAuth = r.Header.Get("Authorization")
				io.WriteString(w, "ok")
			}))
			t.Cleanup(srv.Close)

			m := &manifest.Manifest{Name: "gh", BaseURL: srv.URL, Auth: tt.auth}
			values := map[string]string{"owner": "spf13"}

			var out, errOut bytes.Buffer
			if err := Execute(t.Context(), m, testCommand(), values, format.Raw{}, &out, &errOut); err != nil {
				t.Fatalf("Execute() 에러: %v", err)
			}

			if gotAuth != tt.wantHeader {
				t.Errorf("서버가 받은 Authorization = %q, want %q", gotAuth, tt.wantHeader)
			}

			warn := errOut.String()
			if tt.wantWarn == "" {
				if warn != "" {
					t.Errorf("stderr = %q, want 빈 값", warn)
				}
			} else if !strings.Contains(warn, tt.wantWarn) {
				t.Errorf("stderr = %q, want %q 를 담은 것", warn, tt.wantWarn)
			}
		})
	}
}

// TestExecuteHonorsContextTimeout 은 타임아웃이 "코드에 적혀 있다" 가 아니라
// 실제로 요청을 끊는지 본다. 서버는 일부러 늦게 답한다.
//
// 이 테스트는 0.3초쯤 걸린다고 보고되는데 끊는 데 걸린 시간이 아니다 —
// Execute 는 50ms 에 돌아온다(실측 50.3ms). 나머지는 t.Cleanup 의 srv.Close() 가
// 아직 자고 있는 핸들러가 끝나기를 기다리는 시간이다.
func TestExecuteHonorsContextTimeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
		io.WriteString(w, "늦게 온 답")
	}))
	t.Cleanup(srv.Close)

	ctx, cancel := context.WithTimeout(t.Context(), 50*time.Millisecond)
	defer cancel()

	m := &manifest.Manifest{Name: "gh", BaseURL: srv.URL}
	c := &manifest.Command{Name: "slow", Method: "GET", Path: "/"}

	var out, errOut bytes.Buffer
	err := Execute(ctx, m, c, nil, format.Raw{}, &out, &errOut)

	if err == nil {
		t.Fatal("Execute() = nil, want 타임아웃 에러")
	}
	// 문자열이 아니라 센티널로 판정한다 — 표준 라이브러리가 감싸도 errors.Is 는 뚫고 찾는다.
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("Execute() = %v, want context.DeadlineExceeded 를 감싼 에러", err)
	}
	if out.Len() != 0 {
		t.Errorf("out = %q, want 빈 값 (본문이 오기 전에 끊겼어야 한다)", out.String())
	}
}
