package generate

import (
	"bytes"
	"go/ast"
	"go/parser"
	"go/token"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/Kwangseok-Seo/cli-maker/internal/cli"
	"github.com/Kwangseok-Seo/cli-maker/internal/manifest"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// probeManifest 는 생성기가 실제로 받을 수 있는 것들을 한 곳에 모은 표본이다.
//
// 험한 이름들(따옴표·역슬래시·개행)은 Validate 가 막지 않는다 — 즉 여기까지 온다.
// 날것으로 템플릿에 흘리면 각각 "리터럴 미종료"·"값이 조용히 바뀜"·"미종료"를 일으킨다.
func probeManifest() *manifest.Manifest {
	return &manifest.Manifest{
		Name:    "probe",
		BaseURL: "https://example.test",
		Auth:    manifest.Auth{Type: "bearer", Env: "PROBE_TOKEN"},
		Commands: []manifest.Command{
			{Name: "plain", Method: "GET", Path: "/plain"},
			{
				Name: "mixed", Method: "GET", Path: "/r/{owner}",
				Params: []manifest.Param{
					{Name: "owner", In: "path", Type: "string", Required: true},
					{Name: "repo-name", In: "query", Type: "string"},
				},
			},
			{
				Name: "nasty", Method: "GET", Path: "/q",
				Params: []manifest.Param{
					{Name: `say"hi`, In: "query", Type: `C:\temp`},
					{Name: "multi", In: "query", Type: "two\nlines"},
				},
			},
		},
	}
}

// TestGeneratedHeader 는 템플릿 첫 줄과 GeneratedHeader 상수가 갈리지 않게 붙든다.
//
// 그 둘은 같은 문자열의 복제이고, 갈리면 --dir 의 덮어쓰기 가드가 자기가 만든
// 파일을 못 알아본다 — 조용히 거부하기만 하므로 이 테스트가 없으면 늦게 발견된다.
func TestGeneratedHeader(t *testing.T) {
	var buf bytes.Buffer
	if err := Main(&buf, probeManifest(), "apis/probe.yaml"); err != nil {
		t.Fatalf("Main() 에러: %v", err)
	}

	first, _, _ := strings.Cut(buf.String(), "\n")
	if first != GeneratedHeader {
		t.Errorf("생성물 첫 줄 = %q\nGeneratedHeader = %q\n두 값이 갈렸다 — main.go.tmpl 첫 줄을 맞춰야 한다", first, GeneratedHeader)
	}
}

// TestGoMod 는 생성될 go.mod 를 못 박는다.
//
// require 를 안 쓰는 것이 결정이다 (go mod tidy 가 실제 import 를 보고 채운다).
// go 줄이 1.22 아래로 내려가면 for 루프 변수의 의미가 바뀌어 생성물이 조용히
// 달라지므로, 그 하한도 여기서 지킨다.
func TestGoMod(t *testing.T) {
	var buf bytes.Buffer
	if err := GoMod(&buf, "example.com/gh"); err != nil {
		t.Fatalf("GoMod() 에러: %v", err)
	}

	want := "module example.com/gh\n\ngo " + goVersion + "\n"
	if got := buf.String(); got != want {
		t.Errorf("GoMod() =\n%q\nwant\n%q", got, want)
	}

	major, minor, ok := strings.Cut(goVersion, ".")
	if !ok {
		t.Fatalf("goVersion = %q, want <major>.<minor>", goVersion)
	}
	if major == "1" {
		n, err := strconv.Atoi(minor)
		if err != nil {
			t.Fatalf("goVersion = %q 의 minor 를 읽을 수 없다: %v", goVersion, err)
		}
		if n < 22 {
			t.Errorf("goVersion = %q — 1.22 미만은 for 루프 변수 의미를 바꾼다", goVersion)
		}
	}
}

// TestGeneratedDelegatesSharedFlags 는 생성물이 공유 flag 를 직접 등록하지 않고
// clirun 에 맡기는지 본다.
//
// 직접 등록하면 기본값이 clirun 의 것과 갈리는데, 갈린 것이 --help 에만 보여서
// 조용하다 — 실측: clirun 의 defaultTimeout 만 1ns 로 바꿔도 --help 는 계속
// "default 30s" 를 광고하고 요청은 즉시 데드라인을 넘겼다. 그래서 등록을 옮겼고,
// 다시 돌아오지 않게 여기서 붙든다.
//
// 문자열이 아니라 파서에게 묻는다 — 주석에 적힌 이름을 호출로 세지 않기 위해서다.
func TestGeneratedDelegatesSharedFlags(t *testing.T) {
	var buf bytes.Buffer
	if err := Main(&buf, probeManifest(), "apis/probe.yaml"); err != nil {
		t.Fatalf("Main() 에러: %v", err)
	}

	f, err := parser.ParseFile(token.NewFileSet(), "main.go", buf.Bytes(), 0)
	if err != nil {
		t.Fatalf("생성된 소스를 파싱할 수 없다: %v", err)
	}

	called := map[string]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
				called[sel.Sel.Name] = true
			}
		}
		return true
	})

	if !called["AddSharedFlags"] {
		t.Error("생성물이 clirun.AddSharedFlags 를 부르지 않는다 — 공유 flag 가 아예 없거나 직접 등록했다")
	}
	if called["PersistentFlags"] {
		t.Error("생성물이 공유 flag 를 직접 등록한다 — 기본값이 clirun 과 갈린다")
	}
}

// TestSurfaceMatchesRuntime 은 이 마일스톤의 자물쇠다.
//
// 같은 매니페스트로 만든 두 CLI — 런타임 인터프리터가 세운 cobra 트리와 생성된
// 소스 — 의 명령 표면이 같아야 한다. 실행 본체(clirun.Run)는 공유하지만 표면을
// 만드는 코드는 internal/cli.Build 와 main.go.tmpl 로 둘이라, 한쪽만 고치면
// 갈린다. 이 테스트가 그 순간 깨진다.
//
// 표면을 텍스트로 훑지 않고 파서에게 물어본다 — 생성기가 문법 판정을
// go/format 에 맡긴 것과 같은 이유다. 덤으로, 험한 이름이 %q 를 거쳐 값 그대로
// 살아 돌아오는지도 여기서 함께 확인된다.
func TestSurfaceMatchesRuntime(t *testing.T) {
	// Build 가 붙인 순서 그대로 읽어야 매니페스트 순서까지 대조된다 (ADR-0002).
	cobra.EnableCommandSorting = false

	m := probeManifest()

	var buf bytes.Buffer
	if err := Main(&buf, m, "apis/probe.yaml"); err != nil {
		t.Fatalf("Main() 에러: %v", err)
	}

	got := parseSurface(t, buf.Bytes())
	want := runtimeSurface(m)

	if len(got) != len(want) {
		t.Fatalf("명령 개수: 생성물 %d, 런타임 %d", len(got), len(want))
	}
	for i := range want {
		if got[i].Use != want[i].Use {
			t.Errorf("commands[%d].Use: 생성물 %q, 런타임 %q", i, got[i].Use, want[i].Use)
		}
		if got[i].Short != want[i].Short {
			t.Errorf("commands[%d] %q .Short: 생성물 %q, 런타임 %q", i, want[i].Use, got[i].Short, want[i].Short)
		}
		if len(got[i].Flags) != len(want[i].Flags) {
			t.Errorf("commands[%d] %q flag 개수: 생성물 %d, 런타임 %d", i, want[i].Use, len(got[i].Flags), len(want[i].Flags))
			continue
		}
		for j := range want[i].Flags {
			if got[i].Flags[j] != want[i].Flags[j] {
				t.Errorf("commands[%d] %q flags[%d]: 생성물 %+v, 런타임 %+v", i, want[i].Use, j, got[i].Flags[j], want[i].Flags[j])
			}
		}
	}
}

type flagSpec struct {
	Name     string
	Usage    string
	Required bool
}

type cmdSurface struct {
	Use   string
	Short string
	Flags []flagSpec
}

// runtimeSurface 는 런타임 인터프리터가 세운 cobra 트리에서 표면을 읽는다.
func runtimeSurface(m *manifest.Manifest) []cmdSurface {
	var out []cmdSurface
	for _, sub := range cli.Build(m).Commands() {
		s := cmdSurface{Use: sub.Use, Short: sub.Short}
		sub.Flags().VisitAll(func(f *pflag.Flag) {
			_, required := f.Annotations[cobra.BashCompOneRequiredFlag]
			s.Flags = append(s.Flags, flagSpec{Name: f.Name, Usage: f.Usage, Required: required})
		})
		sortFlags(s.Flags)
		out = append(out, s)
	}
	return out
}

// parseSurface 는 생성된 Go 소스에서 같은 표면을 읽는다.
//
// 명령 하나가 익명 블록 하나이므로, main 의 몸통에서 블록만 골라 각각을 읽으면 된다.
func parseSurface(t *testing.T, src []byte) []cmdSurface {
	t.Helper()

	f, err := parser.ParseFile(token.NewFileSet(), "main.go", src, 0)
	if err != nil {
		t.Fatalf("생성된 소스를 파싱할 수 없다: %v", err)
	}

	var out []cmdSurface
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "main" {
			continue
		}
		for _, stmt := range fn.Body.List {
			if block, ok := stmt.(*ast.BlockStmt); ok {
				out = append(out, parseBlock(t, block))
			}
		}
	}
	return out
}

func parseBlock(t *testing.T, block *ast.BlockStmt) cmdSurface {
	t.Helper()

	var s cmdSurface
	required := map[string]bool{}

	ast.Inspect(block, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.CompositeLit:
			// &clirun.Command{...} 도 여기 걸리므로 cobra.Command 만 본다.
			if !isCobraCommand(v.Type) {
				return true
			}
			for _, elt := range v.Elts {
				kv, ok := elt.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				key, ok := kv.Key.(*ast.Ident)
				if !ok {
					continue
				}
				switch key.Name {
				case "Use":
					s.Use = mustString(t, kv.Value)
				case "Short":
					s.Short = mustString(t, kv.Value)
				}
			}
		case *ast.CallExpr:
			sel, ok := v.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			switch {
			case sel.Sel.Name == "String" && len(v.Args) == 3: // sub.Flags().String(name, "", usage)
				s.Flags = append(s.Flags, flagSpec{
					Name:  mustString(t, v.Args[0]),
					Usage: mustString(t, v.Args[2]),
				})
			case sel.Sel.Name == "MarkFlagRequired" && len(v.Args) == 1:
				required[mustString(t, v.Args[0])] = true
			}
		}
		return true
	})

	for i := range s.Flags {
		s.Flags[i].Required = required[s.Flags[i].Name]
	}
	sortFlags(s.Flags)
	return s
}

func isCobraCommand(e ast.Expr) bool {
	sel, ok := e.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := sel.X.(*ast.Ident)
	return ok && pkg.Name == "cobra" && sel.Sel.Name == "Command"
}

// mustString 은 리터럴을 값으로 되돌린다. %q 로 인용해 넣은 것이 그대로
// 돌아오는지가 여기서 드러난다 — say"hi 가 살아 돌아오면 인용이 옳았던 것이다.
func mustString(t *testing.T, e ast.Expr) string {
	t.Helper()

	lit, ok := e.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		t.Fatalf("문자열 리터럴이 아니다: %T", e)
	}
	s, err := strconv.Unquote(lit.Value)
	if err != nil {
		t.Fatalf("리터럴 %s 를 해석할 수 없다: %v", lit.Value, err)
	}
	return s
}

// 생성물은 매니페스트 순서, cobra 의 VisitAll 은 이름순으로 준다.
// 비교 전에 한쪽 기준으로 맞춘다.
func sortFlags(fs []flagSpec) {
	sort.Slice(fs, func(i, j int) bool { return fs[i].Name < fs[j].Name })
}
