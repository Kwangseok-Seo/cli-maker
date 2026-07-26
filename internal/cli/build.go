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

		for _, p := range c.Params {
			sub.Flags().String(p.Name, "", p.In+" - "+p.Type)
			if p.Required {
				sub.MarkFlagRequired(p.Name)
			}
		}

		group.AddCommand(sub)
	}

	return group
}
