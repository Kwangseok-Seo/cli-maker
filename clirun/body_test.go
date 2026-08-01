package clirun

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// TestAddBodyFlag 는 --data 가 본문을 받는 명령에만 붙는지 본다.
//
// ADR-0010 이 Body 를 포인터로 둔 값이 여기서 나온다 — nil 을 구별할 수 없으면
// 본문을 받지 않는 명령의 --help 에도 --data 가 보인다.
func TestAddBodyFlag(t *testing.T) {
	tests := []struct {
		name         string
		body         *Body
		wantFlag     bool
		wantRequired bool
	}{
		{name: "Body 가 nil 이면 안 붙는다", body: nil},
		{name: "Body 가 있으면 붙는다", body: &Body{}, wantFlag: true},
		{name: "required 면 필수로 표시된다", body: &Body{Required: true}, wantFlag: true, wantRequired: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "x"}
			AddBodyFlag(cmd, tt.body)

			f := cmd.Flags().Lookup(dataFlag)
			if (f != nil) != tt.wantFlag {
				t.Fatalf("--%s 존재 = %v, want %v", dataFlag, f != nil, tt.wantFlag)
			}
			if f == nil {
				return
			}

			// cobra 는 필수 표시를 flag 의 annotation 으로 저장한다. Execute 를 돌리지
			// 않고 확인하려면 그 자리를 봐야 한다.
			_, required := f.Annotations[cobra.BashCompOneRequiredFlag]
			if required != tt.wantRequired {
				t.Errorf("required = %v, want %v", required, tt.wantRequired)
			}
		})
	}
}

// TestResolveBody 는 --data 의 세 입력원과, "안 줬다" 와 "빈 값을 줬다" 의 구별을 못 박는다.
//
// 반환된 리더의 **구체 타입**도 함께 본다. *bytes.Reader / *strings.Reader 가 아니면
// http.NewRequest 가 길이를 알아내지 못해 요청이 chunked 로 나가고 리다이렉트 재전송이
// 조용히 실패한다 (실측: executor 의 TestExecuteSendsBody). 인터페이스만 맞으면 통과하는
// 테스트는 그 결정을 지키지 못한다.
func TestResolveBody(t *testing.T) {
	file := filepath.Join(t.TempDir(), "pet.json")
	if err := os.WriteFile(file, []byte(`{"from":"file"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		body *Body
		// args 는 명령줄. nil 이면 --data 를 주지 않은 것이다.
		args  []string
		stdin string
		// wantNil 이면 본문이 없어야 한다. 아니면 want 와 내용이 같아야 한다.
		wantNil bool
		want    string
		wantErr string
	}{
		{
			name:    "본문을 안 받는 명령이면 nil",
			body:    nil,
			wantNil: true,
		},
		{
			name:    "--data 를 안 주면 nil",
			body:    &Body{},
			wantNil: true,
		},
		{
			// 값이 비었어도 유저가 명시적으로 줬으면 빈 본문을 보낸다. 값만 보면
			// 이 케이스가 위 케이스와 구별되지 않는다 — Changed 가 가른다.
			name: `--data "" 는 빈 본문이지 본문 없음이 아니다`,
			body: &Body{},
			args: []string{"--data", ""},
			want: "",
		},
		{
			name: "리터럴",
			body: &Body{},
			args: []string{"--data", `{"a":1}`},
			want: `{"a":1}`,
		},
		{
			name: "@파일",
			body: &Body{},
			args: []string{"--data", "@" + file},
			want: `{"from":"file"}`,
		},
		{
			// 우리 문구가 아니라 OS 가 쓴 문구라 조각만 본다.
			name:    "없는 파일은 에러다",
			body:    &Body{},
			args:    []string{"--data", "@" + filepath.Join(t.TempDir(), "없다.json")},
			wantErr: "없다.json",
		},
		{
			// os.Stdin 이 아니라 cmd.InOrStdin() 을 읽기 때문에 갈아끼울 수 있다.
			name:  "- 는 stdin",
			body:  &Body{},
			args:  []string{"--data", "-"},
			stdin: `{"from":"stdin"}`,
			want:  `{"from":"stdin"}`,
		},
		{
			// @ 도 - 도 아니면 그냥 리터럴이다. curl 과 같은 관용이라
			// @ 로 시작하는 본문을 리터럴로 보내려면 파일로 우회해야 한다.
			name: "@ 가 뒤에 있으면 리터럴이다",
			body: &Body{},
			args: []string{"--data", "a@b"},
			want: "a@b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "x"}
			AddBodyFlag(cmd, tt.body)
			cmd.SetIn(strings.NewReader(tt.stdin))

			if tt.args != nil {
				if err := cmd.ParseFlags(tt.args); err != nil {
					t.Fatalf("ParseFlags: %v", err)
				}
			}

			got, err := resolveBody(cmd, &Command{Name: "x", Body: tt.body})

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("resolveBody() = %v, want 에러", got)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("에러에 %q 가 없다: %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveBody() 에러: %v", err)
			}

			if tt.wantNil {
				if got != nil {
					t.Errorf("resolveBody() = %#v, want nil", got)
				}
				return
			}
			if got == nil {
				t.Fatal("resolveBody() = nil, want 본문")
			}

			// 길이를 알 수 있는 구체 타입이어야 한다 — 이 단언이 chunked 로 새는 것을 막는다.
			switch got.(type) {
			case *bytes.Reader, *strings.Reader:
			default:
				t.Errorf("리더 타입 = %T, want *bytes.Reader 또는 *strings.Reader", got)
			}

			b, err := io.ReadAll(got)
			if err != nil {
				t.Fatal(err)
			}
			if string(b) != tt.want {
				t.Errorf("본문 = %q, want %q", b, tt.want)
			}
		})
	}
}
