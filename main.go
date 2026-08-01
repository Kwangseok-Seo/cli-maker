package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Kwangseok-Seo/cli-maker/internal/cli"
	"github.com/Kwangseok-Seo/cli-maker/internal/generate"
	"github.com/Kwangseok-Seo/cli-maker/internal/manifest"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
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
			// %+v 가 아니라 YAML 로 낸다. fmt 는 struct 안에 중첩된 포인터를 펼치지 않아
			// body 가 Body:0xc000... 주소로 찍혔다 (M10). 다시 YAML 로 내면 포인터가
			// 펼쳐지고, 유저가 적은 것과 같은 언어라 입력과 나란히 놓고 볼 수 있다.
			// omitempty 를 쓰지 않으므로 빈 필드도 그대로 보인다 — 디버그 도구에서는
			// "이 자리가 비었다"가 곧 답일 때가 많다.
			out, err := yaml.Marshal(m)
			if err != nil {
				fmt.Fprintln(os.Stderr, "parse:", err)
				os.Exit(1)
			}
			cmd.OutOrStdout().Write(out)
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
			// 런타임 등록과 같은 검증을 통과해야 한다.
			if err := manifest.Validate(m); err != nil {
				return err
			}
			// 전역 검증 중 매니페스트 이름 충돌·그룹 persistent flag 는 이 CLI 의 명령
			// 표면에 관한 것이라 생성물과 무관하다. 본문 flag 충돌만 다르다 — 생성된
			// CLI 도 같은 flag 를 달기 때문에, 통과시키면 실행 시점에 패닉하는 소스가 나간다.
			if err := cli.CheckBodyFlags(m); err != nil {
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
			// go.mod 에 require 를 적지 않으므로 tidy 가 cli-maker 버전을 스스로 채운다.
			fmt.Fprintf(errOut, `
다음 단계:
  cd %s
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
