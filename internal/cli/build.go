package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/Kwangseok-Seo/cli-maker/internal/executor"
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
			RunE: func(cmd *cobra.Command, args []string) error {
				cmd.SilenceUsage = true
				values := map[string]string{}
				for _, p := range c.Params {
					values[p.Name], _ = cmd.Flags().GetString(p.Name)
				}
				ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
				defer cancel()
				return executor.Execute(ctx, m, &c, values, cmd.OutOrStdout())
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
