package format

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

// TestFormatters 는 세 포매터가 무엇을 내고 무엇을 경고하는지 못 박는다.
//
// 확인하는 것이 둘이라 판정도 둘이다.
//   - dst 에 나간 바이트 (개행이 붙는지까지 — 그래서 %q 로 본다)
//   - Warn 에 나간 줄 (auto 로 고른 포맷은 조용해야 한다)
func TestFormatters(t *testing.T) {
	tests := []struct {
		name string
		// newf 는 경고를 받을 곳을 넘겨받아 포매터를 만든다.
		// Warn 을 비운 채 만드는 케이스는 이 인자를 그냥 무시한다.
		newf     func(warn io.Writer) Formatter
		in       string
		wantOut  string
		wantWarn string
	}{
		{
			name:    "Raw 는 JSON 을 손대지 않는다 — 개행도 안 붙인다",
			newf:    func(io.Writer) Formatter { return Raw{} },
			in:      `{"a":1}`,
			wantOut: `{"a":1}`,
		},
		{
			name:    "Raw 는 평문도 그대로 흘린다",
			newf:    func(io.Writer) Formatter { return Raw{} },
			in:      "Encourage flow.\n",
			wantOut: "Encourage flow.\n",
		},
		{
			name:    "Pretty 는 두 칸 들여쓰고 개행으로 끝맺는다",
			newf:    func(w io.Writer) Formatter { return Pretty{Warn: w} },
			in:      `{"a":1,"b":[2,3]}`,
			wantOut: "{\n  \"a\": 1,\n  \"b\": [\n    2,\n    3\n  ]\n}\n",
		},
		{
			name:    "Compact 는 공백을 걷어내고 개행으로 끝맺는다",
			newf:    func(w io.Writer) Formatter { return Compact{Warn: w} },
			in:      "{\n  \"a\": 1\n}",
			wantOut: "{\"a\":1}\n",
		},
		{
			// 유저가 -o pretty 를 명시했는데 JSON 이 아니면, 왜 안 이뻐졌는지 알려 준다.
			// 폴백 경로는 바이트를 그대로 두는 게 계약이라 개행을 붙이지 않는다.
			name:     "명시한 Pretty 가 평문을 만나면 원본을 내고 경고한다",
			newf:     func(w io.Writer) Formatter { return Pretty{Warn: w} },
			in:       "Encourage flow.",
			wantOut:  "Encourage flow.",
			wantWarn: "warning: response is not JSON — printing raw\n",
		},
		{
			// ADR-0008: auto 가 고른 Pretty 는 유저가 요청한 적 없으므로 조용히 물러난다.
			// Warn 을 채우지 않는 것이 그 결정을 코드로 표현한 방식이다.
			name:    "auto 로 고른 Pretty(Warn 없음)는 평문에 조용하다",
			newf:    func(io.Writer) Formatter { return Pretty{} },
			in:      "Encourage flow.",
			wantOut: "Encourage flow.",
		},
		{
			// 앞부분이 유효한 JSON 이라 파서가 거기까진 성공한다. 그래도 dst 에는
			// 반쪽짜리가 새면 안 된다 — 원본과 한 바이트도 다르지 않아야 한다.
			name:     "앞부분만 유효한 JSON 이어도 부분 출력이 새지 않는다",
			newf:     func(w io.Writer) Formatter { return Pretty{Warn: w} },
			in:       `{"a":1} 그리고 쓰레기`,
			wantOut:  `{"a":1} 그리고 쓰레기`,
			wantWarn: "warning: response is not JSON — printing raw\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// bytes.Buffer 는 zero value 가 바로 쓸 수 있는 상태다 — 만들 게 없다.
			var out, warn bytes.Buffer

			f := tt.newf(&warn)
			if err := f.Format(&out, strings.NewReader(tt.in)); err != nil {
				t.Fatalf("Format() 에러: %v", err)
			}

			// 개행이 붙었는지 안 붙었는지가 계약이므로 %q 로 본다 — %s 면 안 보인다.
			if got := out.String(); got != tt.wantOut {
				t.Errorf("out = %q, want %q", got, tt.wantOut)
			}
			if got := warn.String(); got != tt.wantWarn {
				t.Errorf("warn = %q, want %q", got, tt.wantWarn)
			}
		})
	}
}
