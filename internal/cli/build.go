package cli

import (
	"fmt"

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
			Run: func(cmd *cobra.Command, args []string) {
				fmt.Println(c.Method, m.BaseURL+c.Path)
			},
		}
		group.AddCommand(sub)
	}

	return group
}
