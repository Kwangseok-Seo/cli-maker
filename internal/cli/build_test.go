package cli

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Kwangseok-Seo/cli-maker/clirun"
	"github.com/Kwangseok-Seo/cli-maker/internal/manifest"
	"github.com/spf13/cobra"
)

// TestBuildRegistersSharedFlags 는 그룹이 공유 flag 를 들고 있는지 본다.
//
// 기본값의 숫자는 단언하지 않는다. 그 값의 정의는 clirun 에 하나뿐이고, 여기서
// 다시 적으면 AddSharedFlags 가 지운 복제를 테스트가 되살린다. 확인하는 것은
// "등록됐는가"와 "유저가 실제로 치는 shorthand 가 살아 있는가" 둘이다.
func TestBuildRegistersSharedFlags(t *testing.T) {
	group := Build(probeManifest("api"))

	for _, name := range []string{clirun.OutputFlag, clirun.TimeoutFlag} {
		if group.PersistentFlags().Lookup(name) == nil {
			t.Errorf("그룹에 --%s 가 없다 — clirun.AddSharedFlags 를 부르지 않았다", name)
		}
	}

	if f := group.PersistentFlags().Lookup(clirun.OutputFlag); f != nil && f.Shorthand != "o" {
		t.Errorf("--%s 의 shorthand = %q, want %q", clirun.OutputFlag, f.Shorthand, "o")
	}
}

// TestBuildEndToEnd 는 루트부터 Execute 로 돌려 배선 전체를 밟는다.
//
// 순수 함수 테스트가 덮지 못하는 것이 여기 있다 — param 이 flag 로 등록됐는가,
// 필수 표시가 실제로 막는가, 공유 flag 가 상속되는가, 그 값이 clirun.Run 을 지나
// 실행기까지 가는가. 하나라도 끊기면 어느 단위 테스트도 안 깨지고 CLI 만 죽는다.
//
// 목을 만들지 않고 httptest 로 진짜 서버를 띄운다 — 판정 방향이 "우리가 무엇을
// 했는가"가 아니라 "상대가 무엇을 받았는가"여야 하기 때문이다.
func TestBuildEndToEnd(t *testing.T) {
	var gotPath, gotQuery, gotBody, gotType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath, gotQuery = r.URL.Path, r.URL.RawQuery
		b, _ := io.ReadAll(r.Body)
		gotBody, gotType = string(b), r.Header.Get("Content-Type")
		io.WriteString(w, "{\"id\":10,\n \"name\":\"Rufus\"}")
	}))
	t.Cleanup(srv.Close)

	tests := []struct {
		name string
		// args 는 루트에 넘길 명령줄. 첫 항목이 그룹 이름이다.
		args []string
		// wantErr 는 Execute 의 에러에 있어야 할 조각. "" 면 성공해야 한다.
		wantErr string
		// wantPath·wantQuery 는 서버가 실제로 받은 것.
		wantPath  string
		wantQuery string
		// wantOut 은 stdout 에 그대로 있어야 할 조각.
		wantOut string
		// wantBody·wantType 은 서버가 실제로 받은 요청 본문과 Content-Type.
		wantBody string
		wantType string
	}{
		{
			// -o compact 가 그룹의 persistent flag 로 상속돼 명령 자리에서 먹는다.
			// 서버가 보낸 본문에 개행이 있으므로 compact 가 실제로 일했는지 보인다.
			name:     "path param 치환 + -o compact",
			args:     []string{"api", "get", "--petId", "10", "-o", "compact"},
			wantPath: "/pet/10",
			wantOut:  `{"id":10,"name":"Rufus"}`,
		},
		{
			// -o 를 그룹 앞자리에 둬도 같은 flag 다.
			name:     "-o 를 그룹 자리에 둬도 같다",
			args:     []string{"api", "-o", "compact", "get", "--petId", "10"},
			wantPath: "/pet/10",
			wantOut:  `{"id":10,"name":"Rufus"}`,
		},
		{
			// MarkFlagRequired 가 실제로 막는지. 막지 못하면 /pet/ 이 나간다.
			name:    "필수 param 을 빼면 요청 전에 막힌다",
			args:    []string{"api", "get"},
			wantErr: `required flag(s) "petId" not set`,
		},
		{
			name:      "query param 은 쿼리로 실린다",
			args:      []string{"api", "find", "--status", "available"},
			wantPath:  "/pet/find",
			wantQuery: "status=available",
		},
		{
			// -o 를 안 주면 auto 다. 여기서 stdout 은 버퍼(터미널 아님)라 raw 가 되고,
			// 서버가 보낸 개행이 그대로 남는다.
			name:     "auto 는 터미널이 아니면 raw 다",
			args:     []string{"api", "get", "--petId", "10"},
			wantPath: "/pet/10",
			wantOut:  "{\"id\":10,\n \"name\":\"Rufus\"}",
		},
		{
			// --data 값이 clirun.Run 의 resolveBody 를 지나 실행기까지 가서 선을 탄다.
			// Content-Type 은 매니페스트가 비워 뒀으므로 기본값이 붙어야 한다.
			name:     "본문은 --data 로 실려 서버까지 간다",
			args:     []string{"api", "add", "--data", `{"name":"rex"}`},
			wantPath: "/pet",
			wantBody: `{"name":"rex"}`,
			wantType: "application/json",
		},
		{
			// AddBodyFlag 가 MarkFlagRequired 를 부르는지. 안 부르면 빈 POST 가 나간다 —
			// M10 이전의 그 동작이다.
			name:    "required 본문을 빼면 요청 전에 막힌다",
			args:    []string{"api", "add"},
			wantErr: `required flag(s) "data" not set`,
		},
		{
			// Body 가 nil 인 명령에는 --data 가 아예 등록되지 않는다. 값 struct 였다면
			// 이 구별이 불가능해 모든 명령에 --data 가 붙었을 것이다 (ADR-0010).
			name:    "본문 없는 명령에 --data 를 주면 거절한다",
			args:    []string{"api", "get", "--petId", "10", "--data", "x"},
			wantErr: "unknown flag: --data",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPath, gotQuery, gotBody, gotType = "", "", "", ""

			m := probeManifest("api")
			m.BaseURL = srv.URL // 바꾸는 것은 목적지뿐이다

			root := &cobra.Command{Use: "cli-maker"}
			root.AddCommand(Build(m))

			var out, errOut bytes.Buffer
			root.SetOut(&out)
			root.SetErr(&errOut)
			root.SetArgs(tt.args)

			err := root.Execute()

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("Execute() = nil, want 에러 %q\nstdout: %q", tt.wantErr, out.String())
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("에러에 %q 가 없다: %v", tt.wantErr, err)
				}
				if gotPath != "" {
					t.Errorf("막혔어야 하는데 서버가 %q 를 받았다", gotPath)
				}
				return
			}

			if err != nil {
				t.Fatalf("Execute() 에러: %v\nstderr: %s", err, errOut.String())
			}
			if gotPath != tt.wantPath {
				t.Errorf("서버가 받은 경로 = %q, want %q", gotPath, tt.wantPath)
			}
			if gotQuery != tt.wantQuery {
				t.Errorf("서버가 받은 쿼리 = %q, want %q", gotQuery, tt.wantQuery)
			}
			if gotBody != tt.wantBody {
				t.Errorf("서버가 받은 본문 = %q, want %q", gotBody, tt.wantBody)
			}
			if gotType != tt.wantType {
				t.Errorf("서버가 받은 Content-Type = %q, want %q", gotType, tt.wantType)
			}
			if tt.wantOut != "" && !strings.Contains(out.String(), tt.wantOut) {
				t.Errorf("stdout = %q, want %q 를 포함", out.String(), tt.wantOut)
			}
		})
	}
}

// probeManifest 는 path param 과 query param 을 하나씩 가진 매니페스트를 낸다.
func probeManifest(name string) *manifest.Manifest {
	return &manifest.Manifest{
		Name:    name,
		BaseURL: "https://example.com",
		Commands: []manifest.Command{
			{
				Name: "get", Method: "GET", Path: "/pet/{petId}",
				Params: []manifest.Param{{Name: "petId", In: "path", Type: "int", Required: true}},
			},
			{
				Name: "find", Method: "GET", Path: "/pet/find",
				Params: []manifest.Param{{Name: "status", In: "query", Type: "string"}},
			},
			{
				// 본문을 받는 명령. Body 가 nil 인 위 둘과 나란히 둬서 --data 가
				// 이 명령에만 붙는지 같은 매니페스트 안에서 대조된다.
				Name: "add", Method: "POST", Path: "/pet",
				Body: &manifest.Body{Required: true},
			},
		},
	}
}
