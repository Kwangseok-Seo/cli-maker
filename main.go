package main

import (
	"fmt"
	"os"
	"time"

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

	// 모든 API 명령이 공유하는 설정. env(CLI_MAKER_TIMEOUT)로도 줄 수 있다 — 우선순위는 internal/cli/timeout.go.
	rootCmd.PersistentFlags().Duration(cli.TimeoutFlag, 30*time.Second, "요청 타임아웃 (env: CLI_MAKER_TIMEOUT)")
	rootCmd.PersistentFlags().StringP(cli.OutputFlag, "o", "auto", "출력 형식: auto|raw|pretty|compact")

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

	for _, err := range cli.LoadDir(rootCmd, "apis") {
		fmt.Fprintln(os.Stderr, "cli-maker:", err)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
