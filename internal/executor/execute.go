package executor

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/Kwangseok-Seo/cli-maker/internal/manifest"
)

// Execute 는 Command 를 실제 HTTP 요청으로 보내고, 응답 본문을 out 에 흘려보낸다.
// 모든 매니페스트·모든 Command 이 이 하나를 공유한다 (CONTEXT.md 의 Executor).
//
// 순서:
//  1. BuildURL 로 최종 URL 을 만든다
//  2. http.NewRequestWithContext(ctx, c.Method, url, nil) 로 요청을 만든다 — 본문 없는 GET 이라 body 는 nil
//  3. http.DefaultClient.Do(req) 로 보낸다
//  4. err 를 검사한 "뒤에" defer resp.Body.Close()
//  5. io.Copy(out, resp.Body) 로 본문을 흘려보낸다
//
// 상태코드(4xx/5xx) 면 본문을 내보낸 뒤 에러를 반환한다.
func Execute(ctx context.Context, m *manifest.Manifest, c *manifest.Command, values map[string]string, out, errOut io.Writer) error {
	reqURL := BuildURL(m, c, values)
	req, err := http.NewRequestWithContext(ctx, c.Method, reqURL, nil)
	if err != nil {
		return err
	}

	applyAuth(req, m.Auth, errOut)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %s", resp.Status)
	}

	return nil
}
