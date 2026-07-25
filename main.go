package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Kwangseok-Seo/cli-maker/internal/cli"
	"github.com/Kwangseok-Seo/cli-maker/internal/manifest"
	"github.com/spf13/cobra"
)

// cli-maker: 여러 web API 를 하나의 CLI 로 다루기 위한 런타임 인터프리터.
// apis/ 의 매니페스트를 런타임에 읽어 <api> <command> 형태의 명령 트리를 만든다 (ADR-0003).
func main() {
	// 매니페스트에 적힌 순서를 --help 에 그대로 보이기 위해 cobra 의 자동 알파벳 정렬을 끈다 (ADR-0002).
	cobra.EnableCommandSorting = false

	rootCmd := &cobra.Command{
		Use:   "cli-maker",
		Short: "여러 web API 를 하나의 CLI 로 다루는 도구",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("안녕, cli-maker! (cobra 루트 명령)")
		},
	}

	greetCmd := &cobra.Command{
		Use:   "greet <이름>",
		Short: "이름을 받아 인사한다",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("안녕, " + args[0] + "!")
		},
	}

	rootCmd.AddCommand(greetCmd)

	parseCmd := &cobra.Command{
		Use:   "parse <파일>",
		Short: "파일을 받아 parse 한다",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			m, err := manifest.Load(args[0])
			if err != nil {
				fmt.Fprintln(os.Stderr, "parse:", err)
				os.Exit(1)
			}
			fmt.Printf("%+v\n", *m)
		},
	}

	rootCmd.AddCommand(parseCmd)

	entries, err := os.ReadDir("apis")
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Fprintln(os.Stderr, "cli-maker:", err)
	}

	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".yaml" {
			continue
		}
		path := filepath.Join("apis", e.Name())

		m, err := manifest.Load(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cli-maker: %s 생략 (%v)\n", path, err)
			continue
		}

		rootCmd.AddCommand(cli.Build(m))
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
