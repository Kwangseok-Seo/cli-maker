// Package clirun 은 cli-maker 가 생성한 CLI 가 쓰는 공개 표면이다.
//
// 생성된 CLI 는 자기 go.mod 를 가진 남의 모듈이라 우리 internal/ 을 import 할 수
// 없다 (Go 의 internal 규칙). 그래서 그쪽이 실제로 필요로 하는 것만 여기로 통과시킨다.
// 여기 있는 것만 외부 계약이고, internal/ 아래는 계속 자유롭게 고칠 수 있다.
//
// # 대상 소비자
//
// 지원하는 소비자는 "cli-maker generate" 가 낸 코드다. 매니페스트를 Go 코드로
// 손수 조립하는 경로는 지원 범위가 아니다 — 유효값 검사가 여기 없기 때문이다.
// generate 는 소스를 내기 전에 검증을 통과시키지만, 직접 조립한 값은 아무도 보지
// 않는다. 예를 들어 Param.In 에 "header" 를 적으면 거부되지 않고 조용히 무시된다
// (실행기는 path 와 query 만 본다).
//
// # cobra 가 새는 것은 의도적이다
//
// Run 이 *cobra.Command 를 받으므로, 이 패키지를 쓰는 모듈은 cli-maker 와 같은
// spf13/cobra 메이저 버전을 써야 한다. flag 값을 미리 읽어 넘기는 시그니처로 바꾸면
// 이 결합은 사라지지만, 그러면 생성된 코드가 flag 를 직접 읽어야 해서 등록과 해석의
// 복제가 늘어난다 — AddSharedFlags 가 지우려는 바로 그 복제다. 결합을 감수하는
// 쪽을 골랐다.
//
// # 안정성
//
// v0.x 다. 표면을 넓히는 것은 하위 호환이지만, 좁히거나 시그니처를 바꾸는 것은
// 이미 나간 생성물을 재생성해야 한다는 뜻이므로 릴리스 노트에 적는다.
package clirun

import (
	"context"

	"github.com/Kwangseok-Seo/cli-maker/internal/executor"
	"github.com/Kwangseok-Seo/cli-maker/internal/manifest"
	"github.com/spf13/cobra"
)

// 생성된 코드가 매니페스트를 Go 값으로 적으려면 이 타입들의 이름이 필요하다.
//
// 정의(type X Y)가 아니라 별칭(type X = Y)이다. 별칭은 새 타입을 만들지 않고 같은
// 타입에 이름을 하나 더 붙일 뿐이라, 여기서 만든 값이 internal 쪽 함수에 변환 없이
// 그대로 들어간다. 정의로 두면 executor.Execute 가 받지 못한다.
//
// 대가가 하나 있다: 원본 타입이 internal/ 에 있어서 go doc 이 필드를 펼쳐 주지
// 않는다 (실측: "go doc ./clirun Manifest" 는 별칭 한 줄만 낸다). 밖에서 필드를
// 알 방법이 없으므로 모양을 여기 적어 둔다. TestAliasFieldsAreDocumented 가
// 이 목록이 실제 필드와 갈리는 순간 깨진다.
//
//	Manifest — Name, BaseURL string · Auth Auth · Commands []Command
//	Command  — Name, Method, Path string · Params []Param
//	Param    — Name, In, Type string · Required bool
//	Auth     — Type, Env string
//
// Method 는 대문자 HTTP 메서드, In 은 "path" 또는 "query" 다. Path 의 자리표시자
// {이름} 은 같은 이름의 path param 이 있어야 채워지고, 그 param 은 Required 여야
// 한다 — 안 그러면 빈 문자열이 치환된 URL 이 나간다.
type (
	Manifest = manifest.Manifest
	Command  = manifest.Command
	Param    = manifest.Param
	Auth     = manifest.Auth
)

// Run 은 API 명령 하나를 실행한다.
//
// 런타임 경로(internal/cli.Build)와 생성 경로가 이 함수 하나를 공유한다. 두 CLI 가
// 같은 코드를 부르므로 동작이 갈릴 수 없다 — 그 사실이 표면 일치의 근거가 된다.
//
// 받는 것은 셋뿐이다: flag 값을 어디서 읽을지(cmd), 무엇을 어디로 보낼지(m, c).
// 타임아웃·출력 포맷·실행은 전부 여기서 정한다.
func Run(cmd *cobra.Command, m *Manifest, c *Command) error {
	// 여기까지 왔으면 인자 파싱은 이미 통과한 것이다. 이후의 실패는 사용법 문제가
	// 아니므로 usage 를 덧붙이지 않는다.
	cmd.SilenceUsage = true

	values := map[string]string{}
	for _, p := range c.Params {
		values[p.Name], _ = cmd.Flags().GetString(p.Name)
	}

	timeout, err := resolveTimeout(cmd)
	if err != nil {
		return err
	}

	f, err := resolveFormatter(cmd)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), timeout)
	defer cancel()

	return executor.Execute(ctx, m, c, values, f, cmd.OutOrStdout(), cmd.ErrOrStderr())
}
