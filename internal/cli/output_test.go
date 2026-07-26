package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Kwangseok-Seo/cli-maker/internal/format"
	"github.com/spf13/cobra"
)

// openDevNull 은 char device 를 하나 연다.
//
// 콘솔을 테스트에서 띄울 수는 없으므로 대리를 쓴다. 대리가 유효한지는 재 봤다 —
// NUL 의 mode 는 Dcrw-rw-rw- 로 M7 에서 실측한 콘솔과 같고, ModeCharDevice 비트가 섰다.
// isTerminal 이 보는 것이 정확히 그 비트라 이 대리로 참 분기를 밟을 수 있다.
func openDevNull(t *testing.T) *os.File {
	t.Helper()
	f, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("%s 를 열 수 없다: %v", os.DevNull, err)
	}
	t.Cleanup(func() { f.Close() })
	return f
}

// TestIsTerminal 은 "어디에 쓰는가"로 터미널 여부를 가르는 판정을 못 박는다.
func TestIsTerminal(t *testing.T) {
	tests := []struct {
		name string
		w    func(t *testing.T) io.Writer
		want bool
	}{
		{
			// 파이프·리다이렉트 대신 메모리 버퍼. *os.File 이 아니라 타입 단언에서 걸린다.
			name: "*os.File 이 아니면 터미널이 아니다",
			w:    func(*testing.T) io.Writer { return &bytes.Buffer{} },
			want: false,
		},
		{
			// *os.File 이긴 하지만 char device 가 아니다 — 리다이렉트가 이 모양이다.
			name: "일반 파일은 터미널이 아니다",
			w: func(t *testing.T) io.Writer {
				f, err := os.Create(filepath.Join(t.TempDir(), "out.txt"))
				if err != nil {
					t.Fatalf("임시 파일 생성 실패: %v", err)
				}
				t.Cleanup(func() { f.Close() })
				return f
			},
			want: false,
		},
		{
			name: "char device 는 터미널로 본다",
			w:    func(t *testing.T) io.Writer { return openDevNull(t) },
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTerminal(tt.w(t)); got != tt.want {
				t.Errorf("isTerminal() = %v, want %v", got, tt.want)
			}
		})
	}
}

// newTestCmd 는 루트가 하는 것과 같은 모양으로 --output flag 를 단 명령을 만든다.
func newTestCmd(out, errOut io.Writer) *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String(OutputFlag, outputAuto, "")
	cmd.SetOut(out)
	cmd.SetErr(errOut)
	return cmd
}

// TestResolveFormatter 는 --output 값이 어떤 포매터가 되는지를,
// 고른 포매터에게 실제로 본문을 흘려 보내 그 결과로 판정한다.
//
// 구조(반환된 struct 의 필드)가 아니라 행동을 보는 이유는, 유저가 겪는 것이
// 행동이기 때문이다. Warn 이 cmd 의 stderr 에 제대로 배선됐는지도 이렇게 해야 드러난다.
func TestResolveFormatter(t *testing.T) {
	tests := []struct {
		name     string
		flag     string // "" 면 flag 를 주지 않은 것 = auto
		in       string
		wantOut  string
		wantWarn string
		wantErr  bool
	}{
		{
			// out 이 버퍼라 터미널이 아니다 → auto 는 Raw 를 고른다.
			name:    "flag 를 안 주면 파이프에선 Raw",
			in:      `{"a":1}`,
			wantOut: `{"a":1}`,
		},
		{
			name:    "raw 는 손대지 않는다",
			flag:    "raw",
			in:      `{"a":1}`,
			wantOut: `{"a":1}`,
		},
		{
			name:    "pretty 는 들여쓴다",
			flag:    "pretty",
			in:      `{"a":1}`,
			wantOut: "{\n  \"a\": 1\n}\n",
		},
		{
			name:    "compact 는 한 줄로 만든다",
			flag:    "compact",
			in:      "{\n  \"a\": 1\n}",
			wantOut: "{\"a\":1}\n",
		},
		{
			// 명시한 pretty 가 평문을 만나면 경고가 cmd 의 stderr 로 가야 한다.
			// Warn 배선이 끊겨 있으면 이 행만 깨진다.
			name:     "명시한 pretty 의 경고는 cmd 의 stderr 로 간다",
			flag:     "pretty",
			in:       "Encourage flow.",
			wantOut:  "Encourage flow.",
			wantWarn: "warning: response is not JSON — printing raw\n",
		},
		{
			name:    "모르는 값은 에러다",
			flag:    "xml",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out, errOut bytes.Buffer
			cmd := newTestCmd(&out, &errOut)
			if tt.flag != "" {
				if err := cmd.Flags().Set(OutputFlag, tt.flag); err != nil {
					t.Fatalf("flag 설정 실패: %v", err)
				}
			}

			f, err := resolveFormatter(cmd)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("resolveFormatter() = %T, want 에러", f)
				}
				return
			}
			if err != nil {
				t.Fatalf("resolveFormatter() 에러: %v", err)
			}

			if err := f.Format(cmd.OutOrStdout(), strings.NewReader(tt.in)); err != nil {
				t.Fatalf("Format() 에러: %v", err)
			}

			if got := out.String(); got != tt.wantOut {
				t.Errorf("out = %q, want %q", got, tt.wantOut)
			}
			if got := errOut.String(); got != tt.wantWarn {
				t.Errorf("stderr = %q, want %q", got, tt.wantWarn)
			}
		})
	}
}

// TestResolveFormatterAutoOnTerminal 은 auto 의 나머지 절반 — 터미널이면 Pretty — 를 본다.
//
// 여기서는 출력이 NUL 로 사라져 행동을 볼 수 없으므로, 예외적으로 고른 타입을 확인한다.
func TestResolveFormatterAutoOnTerminal(t *testing.T) {
	cmd := newTestCmd(openDevNull(t), &bytes.Buffer{})

	f, err := resolveFormatter(cmd)
	if err != nil {
		t.Fatalf("resolveFormatter() 에러: %v", err)
	}

	p, ok := f.(format.Pretty)
	if !ok {
		t.Fatalf("auto + 터미널 = %T, want format.Pretty", f)
	}
	// auto 가 고른 Pretty 는 조용해야 한다 (ADR-0008).
	if p.Warn != nil {
		t.Errorf("auto 로 고른 Pretty 의 Warn = %v, want nil", p.Warn)
	}
}
