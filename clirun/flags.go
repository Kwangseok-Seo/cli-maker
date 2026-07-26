package clirun

import (
	"github.com/spf13/cobra"
)

// AddSharedFlags 는 한 API 의 모든 명령이 공유하는 flag 를 cmd 에 등록한다.
//
// 붙이는 자리는 "그 API 를 덮는 명령"이다 — 런타임 인터프리터에선 매니페스트 그룹,
// 생성된 CLI 에선 그 루트다(둘이 같은 것이다). 요청을 보내지 않는 명령에는 붙지
// 않으므로 --help 가 무영향 flag 를 광고하지 않는다.
//
// 등록과 해석이 한 패키지에 있어야 기본값이 갈리지 않는다. 전에는 등록이 호출자에
// 있고 해석만 여기 있었고, 그래서 기본값이 두 벌이었다 — resolveTimeout 은
// flag 가 주어지지 않으면 등록된 기본값을 무시하고 defaultTimeout 을 쓰므로,
// clirun 쪽만 고치면 --help 는 옛 숫자를 계속 광고했다(실측: defaultTimeout 을
// 1ns 로 바꿔도 --help 는 "default 30s").
func AddSharedFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().Duration(TimeoutFlag, defaultTimeout, "요청 타임아웃 (env: "+timeoutEnv+")")
	cmd.PersistentFlags().StringP(OutputFlag, "o", outputAuto, "출력 형식: auto|raw|pretty|compact")
}
