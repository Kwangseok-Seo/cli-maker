package cli

import (
	"fmt"

	"github.com/Kwangseok-Seo/cli-maker/clirun"
	"github.com/Kwangseok-Seo/cli-maker/internal/manifest"
	"github.com/spf13/cobra"
)

func Build(m *manifest.Manifest) *cobra.Command {
	group := &cobra.Command{
		Use:   m.Name,
		Short: fmt.Sprintf("%s API", m.Name),
	}

	// 이 API 의 모든 명령이 공유하는 설정. 그룹에 두는 이유는 요청을 보내는 명령에만
	// 해당하기 때문이다 — 루트에 두면 generate·greet 까지 이 flag 를 받는다.
	// 등록을 clirun 에 맡기므로 생성된 CLI 와 기본값이 갈릴 수 없다.
	clirun.AddSharedFlags(group)

	for _, c := range m.Commands {
		sub := &cobra.Command{
			Use:   c.Name,
			Short: c.Method + " " + c.Path,
			// 실행 본체는 clirun 이 갖는다 — 생성된 CLI 도 같은 함수를 부르므로
			// 두 경로의 동작이 갈릴 수 없다.
			RunE: func(cmd *cobra.Command, args []string) error {
				return clirun.Run(cmd, m, &c)
			},
		}

		// 본문 flag 를 param 보다 먼저 단다. 등록을 clirun 에 맡기므로 생성된 CLI 와
		// 이름·usage·필수 여부가 갈릴 수 없다.
		clirun.AddBodyFlag(sub, c.Body)

		for _, p := range c.Params {
			// pflag 는 같은 이름을 두 번 받으면 패닉이다. 이름이 이미 잡혀 있으면 등록하지
			// 않는다 — 그 매니페스트는 checkGlobal 이 통째로 거절하므로(ADR-0007) 여기서
			// 조용히 빠진 flag 가 유저에게 도달하지 않는다. Build 는 패닉 없이 돌아가는
			// 것만 책임지고, 보고는 checkGlobal 이 한다.
			if sub.Flags().Lookup(p.Name) != nil {
				continue
			}
			sub.Flags().String(p.Name, "", p.In+" - "+p.Type)
			if p.Required {
				sub.MarkFlagRequired(p.Name)
			}
		}

		group.AddCommand(sub)
	}

	return group
}
