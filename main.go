package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Kwangseok-Seo/cli-maker/clirun"
	"github.com/Kwangseok-Seo/cli-maker/internal/cli"
	"github.com/Kwangseok-Seo/cli-maker/internal/generate"
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
	rootCmd.PersistentFlags().Duration(clirun.TimeoutFlag, 30*time.Second, "요청 타임아웃 (env: CLI_MAKER_TIMEOUT)")
	rootCmd.PersistentFlags().StringP(clirun.OutputFlag, "o", "auto", "출력 형식: auto|raw|pretty|compact")

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

	// generate 는 런타임 해석의 반대편이다 — 같은 매니페스트를 읽어, 실행하는 대신
	// 그 API 전용 CLI 의 Go 소스를 낸다. 소스는 stdout 으로만 나가므로 파이프·리다이렉트가
	// 자연스럽고, 진단은 stderr 로 간다.
	var genDir, genModule string
	generateCmd := &cobra.Command{
		Use:   "generate <매니페스트>",
		Short: "매니페스트로 독립 CLI 의 Go 소스를 만든다",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			m, err := manifest.Load(args[0])
			if err != nil {
				return err
			}
			// 런타임 등록과 같은 검증을 통과해야 한다. 전역 검증(checkGlobal)은
			// 이 CLI 의 명령 표면에 관한 것이라 생성물과는 무관해 여기선 보지 않는다.
			if err := manifest.Validate(m); err != nil {
				return err
			}

			// --dir 없이 부르면 소스만 stdout 으로 낸다 — 리뷰하거나 파이프로 넘길 때.
			if genDir == "" {
				return generate.Main(cmd.OutOrStdout(), m, args[0])
			}

			module := genModule
			if module == "" {
				module = m.Name
			}

			kept, err := generate.Module(genDir, m, args[0], module)
			if err != nil {
				return err
			}

			// 여기부터는 진단이므로 stderr 로 간다.
			errOut := cmd.ErrOrStderr()
			fmt.Fprintln(errOut, "생성:", filepath.Join(genDir, "main.go"))
			for _, p := range kept {
				fmt.Fprintln(errOut, "유지:", p, "(이미 있어 건드리지 않았다)")
			}
			if len(kept) == 0 {
				fmt.Fprintln(errOut, "생성:", filepath.Join(genDir, "go.mod"))
			}
			// cli-maker 자신의 소스 위치는 실행 중에 알 수 없다 — 자리표시자로 둔다.
			fmt.Fprintf(errOut, `
다음 단계 — cli-maker 가 아직 배포되지 않아 로컬 소스를 가리켜야 한다:
  cd %s
  go mod edit -replace github.com/Kwangseok-Seo/cli-maker=<cli-maker 저장소 경로>
  go mod tidy
  go build -o %s .
`, genDir, module)
			return nil
		},
	}
	generateCmd.Flags().StringVar(&genDir, "dir", "", "main.go 와 go.mod 를 쓸 디렉토리 (없으면 소스를 stdout 으로)")
	generateCmd.Flags().StringVar(&genModule, "module", "", "생성될 모듈 경로 (기본: 매니페스트의 name)")

	rootCmd.AddCommand(generateCmd)

	for _, err := range cli.LoadDir(rootCmd, "apis") {
		fmt.Fprintln(os.Stderr, "cli-maker:", err)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
