// Package format 은 API 응답 본문을 어떤 모양으로 낼지를 정한다.
//
// 포맷을 "고르는 곳"(플래그를 읽는 internal/cli)과 "쓰는 곳"(응답이 도착하는
// internal/executor)이 떨어져 있어서, 그 사이를 인터페이스 하나로 잇는다.
// executor 는 "pretty" 같은 문자열의 의미를 알 필요가 없다.
package format

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// Formatter 는 응답 본문 src 를 읽어 어떤 모양으로든 dst 에 쓴다.
//
// 구현은 src 를 정확히 한 번만 소비한다. 흘려보낼 수 있는 구현(Raw)은
// 스트리밍으로, 전체가 있어야 하는 구현(Pretty/Compact)은 자기 안에서
// io.ReadAll 로 버퍼링한다 — 비용을 필요한 구현만 낸다.
type Formatter interface {
	Format(dst io.Writer, src io.Reader) error
}

// Raw 는 받은 바이트를 그대로 흘려보낸다. M4 의 io.Copy 가 여기로 이사 온 것.
// 상태가 없으므로 빈 구조체다 (필드 0개, 크기 0바이트).
type Raw struct{}

func (r Raw) Format(dst io.Writer, src io.Reader) error {
	_, err := io.Copy(dst, src)
	return err
}

// Pretty 는 JSON 을 두 칸 들여쓰기로 편다.
// 본문이 JSON 이 아니면 원본을 그대로 낸다 (json.Indent 는 실패 시 dst 를
// 건드리지 않으므로, 부분 출력이 새어 나갈 걱정 없이 되돌릴 수 있다).
type Pretty struct {
	// Warn 이 nil 이 아니면 raw 로 되돌릴 때 여기에 한 줄 남긴다.
	// 유저가 -o pretty 를 명시했을 때만 채운다 — 우리가 알아서 고른 pretty 가
	// 실패한 것은 유저가 요청한 적 없는 일이라 알릴 이유가 없다.
	Warn io.Writer
}

func (p Pretty) Format(dst io.Writer, src io.Reader) error {
	return rewrite(dst, src, p.Warn, func(buf *bytes.Buffer, body []byte) error {
		return json.Indent(buf, body, "", "  ")
	})
}

// Compact 는 JSON 의 공백을 걷어내 한 줄로 만든다.
// 이미 한 줄로 오는 API(GitHub)엔 아무 일도 없고, 들여써서 주는 API 에서 일한다.
type Compact struct {
	Warn io.Writer
}

func (c Compact) Format(dst io.Writer, src io.Reader) error {
	return rewrite(dst, src, c.Warn, json.Compact)
}

// rewrite 는 본문 전체를 읽어 fn 으로 다시 쓴다. Pretty 와 Compact 가 공유한다.
//
// fn 이 실패하면 = 본문이 JSON 이 아니면 원본을 그대로 내보낸다. json.Indent/Compact 는
// 실패 시 buf 를 원래 길이로 되돌리므로(실측: 앞부분이 유효해도 0 바이트), 여기서
// 부분 출력이 새어 나갈 일이 없다.
func rewrite(dst io.Writer, src io.Reader, warn io.Writer, fn func(*bytes.Buffer, []byte) error) error {
	body, err := io.ReadAll(src)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := fn(&buf, body); err != nil {
		if warn != nil {
			fmt.Fprintln(warn, "warning: response is not JSON — printing raw")
		}
		_, err := dst.Write(body)
		return err
	}

	if _, err := dst.Write(buf.Bytes()); err != nil {
		return err
	}

	// 사람이 읽으라고 다시 쓴 출력이니 개행으로 끝맺는다 — 안 그러면 셸 프롬프트가
	// 마지막 줄에 들러붙는다. Raw 와 위의 폴백 경로는 바이트를 그대로 두는 게 계약이라
	// 여기(포맷에 성공한 경로)에서만 붙인다.
	_, err = io.WriteString(dst, "\n")
	return err
}
