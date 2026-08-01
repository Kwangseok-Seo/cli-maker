package clirun

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// dataFlag 는 요청 본문을 받는 flag 이름이다. 등록(AddBodyFlag)과 해석(resolveBody)이
// 같은 상수를 보므로 한쪽만 바뀔 수 없다.
const dataFlag = "data"

// AddBodyFlag 는 본문을 받는 명령에 --data 를 단다. b 가 nil 이면 아무것도 하지 않는다.
//
// nil 일 때 안 다는 것이 요점이다 — 본문을 받지 않는 명령의 --help 에 --data 가 보이면
// 표면이 거짓말을 한다. ADR-0010 이 Body 를 포인터로 둔 이유가 여기서 쓰인다.
//
// 런타임 경로(internal/cli.Build)와 생성된 main.go 가 이 함수를 공유한다. 각자 등록하면
// flag 이름·usage·필수 여부가 두 벌이 되어 갈린다 — AddSharedFlags 와 같은 이유다.
func AddBodyFlag(cmd *cobra.Command, b *Body) {
	if b == nil {
		return
	}
	cmd.Flags().String(dataFlag, "", "요청 본문 (@파일 이면 파일, - 이면 stdin)")
	if b.Required {
		cmd.MarkFlagRequired(dataFlag)
	}
}

// resolveBody 는 --data 값을 실제 요청 본문으로 바꾼다. 본문이 없으면 nil 을 돌려준다.
//
// 세 입력원을 받는다 (curl 계보):
//
//	--data '{"a":1}'    리터럴
//	--data @pet.json    파일
//	--data -            stdin
//
// 어느 경로든 내용을 다 읽어 *bytes.Reader 나 *strings.Reader 로 감싼다. 스트리밍으로
// 흘리면 http.NewRequest 가 길이를 알아내지 못해 요청이 chunked 로 나가고, 되감기용
// GetBody 가 없어 리다이렉트 재전송이 조용히 실패한다 (실측: executor 의
// TestExecuteSendsBody). 대가는 본문이 통째로 메모리에 오르는 것이다.
//
// "flag 를 안 줬다" 와 "빈 값을 줬다" 는 Changed 로 가른다. 값만 보면 --data "" 가
// 본문 없음과 구별되지 않는다 — yaml 의 "키 없음" 과 "빈 것" 을 포인터로 갈랐던 것과
// 같은 문제가 flag 층에도 있고, pflag 의 답이 Changed 다.
func resolveBody(cmd *cobra.Command, c *Command) (io.Reader, error) {
	if c.Body == nil || !cmd.Flags().Changed(dataFlag) {
		return nil, nil
	}

	raw, err := cmd.Flags().GetString(dataFlag)
	if err != nil {
		return nil, err
	}

	switch {
	case raw == "-":
		// os.Stdin 이 아니라 cmd.InOrStdin() 이다 — 테스트가 갈아끼울 수 있어야 한다.
		b, err := io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return nil, fmt.Errorf("stdin 을 읽을 수 없다: %w", err)
		}
		return bytes.NewReader(b), nil
	case strings.HasPrefix(raw, "@"):
		b, err := os.ReadFile(raw[1:])
		if err != nil {
			return nil, err
		}
		return bytes.NewReader(b), nil
	default:
		return strings.NewReader(raw), nil
	}
}
