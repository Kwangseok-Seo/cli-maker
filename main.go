package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/Kwangseok-Seo/cli-maker/internal/cli"
	"github.com/Kwangseok-Seo/cli-maker/internal/generate"
	"github.com/Kwangseok-Seo/cli-maker/internal/manifest"
	"github.com/Kwangseok-Seo/cli-maker/internal/openapi"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// cli-maker: 여러 web API 를 하나의 CLI 로 다루기 위한 런타임 인터프리터.
// apis/ 의 매니페스트를 런타임에 읽어 <api> <command> 형태의 명령 트리를 만든다 (ADR-0003).
func main() {
	// 매니페스트에 적힌 순서를 --help 에 그대로 보이기 위해 cobra 의 자동 알파벳 정렬을 끈다 (ADR-0002).
	cobra.EnableCommandSorting = false

	rootCmd := &cobra.Command{
		Use:     "cli-maker",
		Short:   "여러 web API 를 하나의 CLI 로 다루는 도구",
		Version: buildVersion(),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("안녕, cli-maker! (cobra 루트 명령)")
		},
	}

	// cobra 는 Version 이 비어 있지 않을 때만 --version 을 단다. 기본 템플릿이
	// "cli-maker version v0.1.0" 이라 v 가 두 번 나오므로 형식만 바꾼다.
	rootCmd.SetVersionTemplate("{{.Name}} {{.Version}}\n")

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
			// auth 와 param 의 빈 필드는 그대로 보인다 — 디버그 도구에서는 "이 자리가
			// 비었다"가 곧 답일 때가 많다(오타 난 키는 조용히 버려지므로 빈 값이 유일한
			// 단서다). Params·Body 만 omitempty 인데, 그 둘은 import 산출물에서 노이즈가
			// 되면서 오타 단서 역할은 안 하기 때문이다 — 근거는 manifest.Command 주석.
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

	rootCmd.AddCommand(newImportCmd())

	for _, err := range cli.LoadDirs(rootCmd, apiDirs()) {
		fmt.Fprintln(os.Stderr, "cli-maker:", err)
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// buildVersion 은 이 바이너리에 박힌 버전을 읽는다.
//
// 버전 문자열을 소스에 적거나 -ldflags -X 로 주입하지 않는다 — go 도구가 이미 박아 두기
// 때문이다. 무엇이 박히는지는 어떻게 빌드했는지로 갈린다 (실측):
//
//	go install …@v0.1.0     v0.1.0                                태그 그대로
//	go build (repo, 깨끗)   v0.1.1-0.20260802101701-26ab395c90b0  직전 태그 다음의 pseudo-version
//	go build (repo, 더러움) 위와 같고 뒤에 +dirty
//	go build (repo 밖)      (devel)                               vcs 정보가 없다
//
// 커밋 해시와 시각이 pseudo-version 안에 이미 들어 있으므로 vcs.revision 을 따로 꺼내지
// 않는다. 커밋 안 된 변경으로 만든 바이너리인지도 go 가 +dirty 로 붙여 준다 — 버그 리포트에
// 붙은 버전이 어느 소스인지 되짚을 수 있어야 하고, 그 판정을 우리가 다시 구현할 이유가 없다.
func buildVersion() string {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		// 모듈 모드로 빌드되지 않은 경우. 지금의 빌드 방식에서는 오지 않는다.
		return "unknown"
	}
	return fmt.Sprintf("%s (%s %s/%s)", bi.Main.Version, bi.GoVersion, runtime.GOOS, runtime.GOARCH)
}

// apiDirs 는 매니페스트를 찾을 디렉토리를 가까운 것부터 나열한다.
//
// 첫째는 프로세스 cwd 의 apis/ — 그 프로젝트 전용 API 가 놓이는 자리다.
// 둘째는 유저 설정 디렉토리 — 설치본이 어디서 실행되든 따라오는 자리다.
//
// os.UserConfigDir 은 OS 마다 다른 규약을 하나로 감싼다 (Windows %AppData%, macOS
// ~/Library/Application Support, Linux $XDG_CONFIG_HOME 또는 ~/.config). 다만 디렉토리를
// 만들어 주지도, 앱 이름을 붙여 주지도 않는다 — cli-maker/apis 는 우리가 붙인다.
//
// 설정 디렉토리를 못 알아내는 환경(%AppData% 가 없는 서비스 계정 등)에서는 그 자리를
// 건너뛴다. 발견 경로 하나가 없는 것은 실패가 아니다 — cwd 의 apis/ 만으로도 도구는 돈다.
func apiDirs() []string {
	dirs := []string{"apis"}
	if cfg, err := os.UserConfigDir(); err == nil {
		dirs = append(dirs, filepath.Join(cfg, "cli-maker", "apis"))
	}
	return dirs
}

// newImportCmd 는 import 명령을 만든다.
//
// import 는 generate 의 반대편 입구다 — generate 가 매니페스트를 소스로 내보낸다면,
// import 는 남의 명세(OpenAPI Spec)를 매니페스트로 들여온다. 산출물은 어디까지나
// 매니페스트이고, 런타임이 Spec 을 직접 읽는 일은 없다 (CONTEXT.md 의 Spec ≠ Manifest).
//
// 다른 명령들과 달리 생성자로 뺀 이유는 테스트 때문이다. flag 를 담는 변수가 이
// 함수 안에서 태어나므로 부를 때마다 새 통이 생기고, 테스트가 한 프로세스에서
// 여러 번 실행해도 앞 실행의 flag 값이 남지 않는다.
func newImportCmd() *cobra.Command {
	var impName, impBaseURL, impOut string

	importCmd := &cobra.Command{
		Use:   "import <spec>",
		Short: "OpenAPI Spec 을 읽어 매니페스트 YAML 을 만든다",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true

			spec, err := openapi.Load(args[0])
			if err != nil {
				return err
			}

			// 이름은 Spec 에서 못 얻는다 — info.title 은 정규화해도 명령 이름이
			// 되지 않는다("swagger-petstore-openapi-3-0"). --name 이 없으면
			// --out 파일명에서 가져오고, 둘 다 없으면 물어보는 수밖에 없다.
			name := impName
			if name == "" && impOut != "" {
				base := filepath.Base(impOut)
				name = strings.TrimSuffix(base, filepath.Ext(base))
			}
			if name == "" {
				return errors.New("--name 이 필요하다 (--out 을 주면 그 파일명에서 가져온다)")
			}

			// 덮어쓰기는 거절한다. 이 파일은 유저가 손으로 이어 쓰는 초안이라,
			// 두 번째 import 가 편집분을 조용히 지우면 되돌릴 방법이 없다.
			if impOut != "" {
				if _, err := os.Stat(impOut); err == nil {
					return fmt.Errorf("%s 가 이미 있다 — 지우거나 다른 --out 을 주어야 한다", impOut)
				}
			}

			m, warnings, err := openapi.Convert(spec, name, impBaseURL)
			// 경고는 실패했을 때도 낸다 — 무엇을 잃었는지가 원인의 단서일 때가 있다.
			for _, w := range warnings {
				fmt.Fprintln(cmd.ErrOrStderr(), "cli-maker:", w)
			}
			if err != nil {
				return err
			}
			// 생성된 매니페스트가 등록될 수 있는지까지 본다. Validate 가 못 보는 것은
			// 본문 flag 충돌뿐이라 generate 와 같은 검사를 건다.
			if err := cli.CheckBodyFlags(m); err != nil {
				return err
			}

			// M2 에서 배운 언마샬의 역방향. 같은 struct 태그가 양쪽을 다 정한다.
			out, err := yaml.Marshal(m)
			if err != nil {
				return err
			}

			if impOut == "" {
				_, err := cmd.OutOrStdout().Write(out)
				return err
			}
			if err := os.WriteFile(impOut, out, 0o644); err != nil {
				return err
			}
			fmt.Fprintln(cmd.ErrOrStderr(), "생성:", impOut, fmt.Sprintf("(명령 %d개)", len(m.Commands)))
			return nil
		},
	}
	importCmd.Flags().StringVar(&impName, "name", "", "매니페스트 이름 = CLI 의 그룹 명령 (기본: --out 파일명)")
	importCmd.Flags().StringVar(&impBaseURL, "base-url", "", "baseUrl 로 쓸 절대 URL (spec 의 servers 가 상대 URL 이면 필수)")
	importCmd.Flags().StringVar(&impOut, "out", "", "쓸 파일 (없으면 stdout)")

	return importCmd
}
