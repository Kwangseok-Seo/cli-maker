package cli

import (
	"context"
	"fmt"

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
				return executor.Execute(ctx, m, &c, values, f, cmd.OutOrStdout(), cmd.ErrOrStderr())
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
