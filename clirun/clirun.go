// Package clirun 은 cli-maker 가 생성한 CLI 가 쓰는 공개 표면이다.
//
// 생성된 CLI 는 자기 go.mod 를 가진 남의 모듈이라 우리 internal/ 을 import 할 수
// 없다 (Go 의 internal 규칙). 그래서 그쪽이 실제로 필요로 하는 것만 여기로 통과시킨다.
// 여기 있는 것만 외부 계약이고, internal/ 아래는 계속 자유롭게 고칠 수 있다.
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
