package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// cli-maker: 여러 web API 를 하나의 CLI 로 다루기 위한 런타임 인터프리터.
// M0 골격 — 다음 단계(M1)에서 cobra 루트 명령으로 대체한다.
func main() {
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

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
