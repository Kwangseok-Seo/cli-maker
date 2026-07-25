package executor

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/Kwangseok-Seo/cli-maker/internal/manifest"
)

// applyAuth 는 매니페스트의 auth 설정을 보고 요청에 인증 헤더를 붙인다.
//
// 토큰을 얻지 못하면 헤더를 붙이지 않고 errOut 에 경고만 남긴다 — 요청 자체는 그대로 나간다.
// (익명으로도 200 이 나오는 엔드포인트를 막지 않기 위해. 빈 토큰을 붙이면 그런 곳도 401 이 된다.)
func applyAuth(req *http.Request, a manifest.Auth, errOut io.Writer) {
	switch a.Type {
	case "":
		return
	case "bearer":
		v, ok := os.LookupEnv(a.Env)
		if !ok {
			fmt.Fprintf(errOut, "warning: %s is not set — sending unauthenticated request\n", a.Env)
			return
		}
		if v == "" {
			fmt.Fprintf(errOut, "warning: %s is set but empty — sending unauthenticated request\n", a.Env)
			return
		}
		req.Header.Set("Authorization", "Bearer "+v)
		return
	default:
		fmt.Fprintf(errOut, "warning: unknown auth type %q — sending unauthenticated request\n", a.Type)
	}
}
