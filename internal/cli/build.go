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
