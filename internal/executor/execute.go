package executor

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/Kwangseok-Seo/cli-maker/internal/format"
	"github.com/Kwangseok-Seo/cli-maker/internal/manifest"
)

// Call 은 "무엇을 어디로 보내는가" 한 벌이다.
//
// 넷을 묶은 이유는 개수가 아니라 짝이다 — Values 는 Command.Params 를, Body 는
// Command.Body 를 유저 입력으로 채운 것이라 Command 없이는 뜻이 없다. 위치 인자로
// 늘어놓으면 nil 을 받을 수 있는 자리가 나란히 붙어(Values, Body) 뒤바뀌어도 컴파일이
// 통과한다 — 필드 이름이 그 자리를 없앤다.
type Call struct {
	Manifest *manifest.Manifest
	Command  *manifest.Command
	// Values 는 param 이름 → 유저가 준 값. 안 준 param 은 빈 문자열이거나 없다.
	Values map[string]string
	// Body 는 요청 본문. nil 이면 본문 없는 요청이다.
	Body io.Reader
}

// Execute 는 Command 를 실제 HTTP 요청으로 보내고, 응답 본문을 f 를 거쳐 out 에 낸다.
// 모든 매니페스트·모든 Command 이 이 하나를 공유한다 (CONTEXT.md 의 Executor).
//
// 순서:
//  1. BuildURL 로 최종 URL 을 만든다
//  2. http.NewRequestWithContext(ctx, c.Method, url, body) 로 요청을 만든다
//  3. http.DefaultClient.Do(req) 로 보낸다
//  4. err 를 검사한 "뒤에" defer resp.Body.Close()
//  5. f.Format(out, resp.Body) 로 본문을 내보낸다
//
// body 가 nil 이면 본문 없는 요청이다. nil 이 아니면 Content-Type 이 함께 붙는다.
//
// body 의 "구체 타입"이 선을 타는 모양을 바꾼다. http.NewRequest 는 *bytes.Buffer ·
// *bytes.Reader · *strings.Reader 셋일 때만 길이를 알아내 ContentLength 를 채우고
// 되감기용 GetBody 를 심는다. 그 밖의 io.Reader 는 길이를 모르므로 요청이 chunked 로
// 나가고, 리다이렉트를 만나면 본문을 다시 읽을 수 없어 재전송이 되지 않는다.
// 그래서 호출자가 무엇을 넘기는지가 중요하다 — 인터페이스가 같아도 결과가 다르다.
//
// f 가 어떤 포맷인지는 여기서 알지 못한다 — resp.Body 를 넘길 뿐이다. Raw 면
// 스트리밍으로 흘러가고, Pretty/Compact 면 그쪽이 알아서 버퍼링한다.
//
// 상태코드(4xx/5xx) 면 본문을 내보낸 뒤 에러를 반환한다 — 에러 본문도 같은 포맷을 탄다.
func Execute(ctx context.Context, call Call, f format.Formatter, out, errOut io.Writer) error {
	reqURL := BuildURL(call.Manifest, call.Command, call.Values)
	req, err := http.NewRequestWithContext(ctx, call.Command.Method, reqURL, call.Body)
	if err != nil {
		return err
	}

	// 본문이 있을 때만 붙인다 — 본문 없는 요청에 Content-Type 을 다는 것은 거짓말이다.
	if call.Body != nil {
		req.Header.Set("Content-Type", call.Command.Body.ContentTypeOrDefault())
	}

	applyAuth(req, call.Manifest.Auth, errOut)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if err := f.Format(out, resp.Body); err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %s", resp.Status)
	}

	return nil
}
