package cli

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/Kwangseok-Seo/cli-maker/internal/manifest"
	"github.com/spf13/cobra"
)

// manifestYAML 은 검증을 통과하는 최소 매니페스트를 낸다.
//
// 하위 명령 이름까지 받는 이유는 이름 충돌 케이스 때문이다 — 두 파일이 같은
// api 이름을 주장할 때, 이긴 쪽이 어느 파일인지는 하위 명령으로만 구별된다.
func manifestYAML(api, command string) string {
	return "name: " + api + "\n" +
		"baseUrl: https://example.com\n" +
		"commands:\n" +
		"  - name: " + command + "\n" +
		"    method: GET\n" +
		"    path: /x\n"
}

// newRoot 는 판정에 필요한 만큼의 루트를 조립한다.
//
// main.go 의 루트를 빌려올 수 없다 — package main 은 임포트할 수 없다. 그런데 그게
// 이 테스트의 요점이다. checkGlobal 은 예약어를 상수로 두지 않고 root 에 물어보므로
// (ADR-0007), "무엇이 예약어인가"는 여기 조립한 표면이 그대로 정한다. 상수 목록이었다면
// 테스트는 그 목록을 다시 적었을 것이고, 그건 검증이 아니라 복제다.
func newRoot() *cobra.Command {
	root := &cobra.Command{Use: "cli-maker"}
	root.PersistentFlags().String("output", "auto", "")
	root.AddCommand(&cobra.Command{Use: "greet"})
	return root
}

// TestLoadDir 는 디렉토리 스캔부터 명령 등록까지의 계약을 못 박는다.
//
// 케이스마다 둘을 함께 본다 — 무엇이 보고됐는가(errs)와 무엇이 실제로 붙었는가
// (root 의 자식). 하나만 보면 격벽이 무너진 것을 놓친다: 에러를 냈는데 등록도
// 해 버린 경우와, 조용히 아무것도 안 한 경우가 한쪽 눈으로는 구별되지 않는다.
func TestLoadDir(t *testing.T) {
	// main.go 가 끄는 것과 같은 스위치를 테스트도 끈다. 켜 두면 cobra 가 자식을
	// 알파벳순으로 재배열해, LoadDir 이 파일명 순으로 붙였다는 사실을 볼 수 없다 (ADR-0002).
	cobra.EnableCommandSorting = false

	tests := []struct {
		name string
		// files 는 디렉토리에 써 넣을 파일. os.ReadDir 이 이름순으로 주므로
		// 순서가 판정에 걸리는 케이스는 파일 이름으로 통제한다.
		files map[string]string
		// dirs 는 .yaml 이름을 가진 디렉토리 — 항목이 파일이 아닐 때.
		dirs []string
		// noDir 이면 스캔할 디렉토리 자체를 만들지 않는다.
		noDir bool
		// wantErrs[i] 는 i번째 에러에 들어 있어야 할 조각들. 개수도 곧 기대치다.
		wantErrs [][]string
		// wantNotErrs 는 어느 에러에도 나오면 안 되는 조각. 과잉 보고를 잡는다 —
		// "무엇을 잡았나" 만 보면 "무엇을 안 잡았어야 하나" 가 검증되지 않는다.
		wantNotErrs []string
		// wantCmds 는 LoadDir 이 새로 붙인 그룹 이름 — 붙은 순서대로.
		wantCmds []string
		// wantSubs 는 그룹별 하위 명령. 충돌에서 누가 이겼는지가 여기서 드러난다.
		wantSubs map[string][]string
	}{
		{
			// apis/ 가 없는 것은 잘못이 아니다 — 매니페스트 없이도 CLI 는 돈다.
			name:  "디렉토리가 없으면 에러가 아니다",
			noDir: true,
		},
		{
			name: "빈 디렉토리도 에러가 아니다",
		},
		{
			// api 이름을 일부러 역순(zebra, alpha)으로 둬서 알파벳 정렬과 구별되게 했다.
			// 붙는 순서를 정하는 것은 api 이름이 아니라 파일 이름이다.
			name: "매니페스트는 파일명 순으로 붙는다",
			files: map[string]string{
				"aaa.yaml": manifestYAML("zebra", "ping"),
				"bbb.yaml": manifestYAML("alpha", "ping"),
			},
			wantCmds: []string{"zebra", "alpha"},
		},
		{
			// M3 이 세운 격벽 — 하나가 깨져도 나머지는 산다.
			name: "깨진 YAML 은 그 파일만 빠진다",
			files: map[string]string{
				"aaa.yaml": manifestYAML("good", "ping"),
				"bbb.yaml": "name: [unclosed",
			},
			wantErrs: [][]string{{"bbb.yaml 생략 (", "yaml:"}},
			wantCmds: []string{"good"},
		},
		{
			// ADR-0007: 먼저 등록된 쪽이 이긴다 (os.ReadDir 이 이름순이라 결정론적).
			// 하위 명령 이름을 다르게 둬서 이긴 쪽을 눈으로 확인한다.
			name: "이름이 충돌하면 먼저 읽힌 파일이 이긴다",
			files: map[string]string{
				"aaa.yaml": manifestYAML("dup", "first"),
				"bbb.yaml": manifestYAML("dup", "second"),
			},
			wantErrs: [][]string{{"bbb.yaml 생략", `name "dup" 는 이미 쓰이고 있는 명령 이름이다`}},
			wantCmds: []string{"dup"},
			wantSubs: map[string][]string{"dup": {"first"}},
		},
		{
			// greet 는 newRoot 이 미리 붙여 둔 명령이다. 예약어를 상수로 적지 않고
			// root 에 물어보기 때문에, 여기 greet 를 붙인 것만으로 예약어가 된다.
			name: "이미 있는 명령 이름은 거부한다",
			files: map[string]string{
				"aaa.yaml": manifestYAML("greet", "ping"),
			},
			wantErrs: [][]string{{"aaa.yaml 생략", `name "greet" 는 이미 쓰이고 있는 명령 이름이다`}},
		},
		{
			// newRoot 이 --output 을 persistent flag 로 달았으므로 param output 은
			// 그 flag 를 가려 버린다. M7 에서 --output 이 생긴 순간 이 검사가 코드 수정
			// 없이 넓어졌다 — 목록을 상수로 적지 않은 대가로 받은 것이다.
			name: "param 이 예약된 flag 이름이면 거부한다",
			files: map[string]string{
				"aaa.yaml": `
name: api
baseUrl: https://example.com
commands:
  - name: ping
    method: GET
    path: /x
    params:
      - name: output
        in: query
        type: string
`,
			},
			wantErrs: [][]string{{"aaa.yaml 생략", `param "output" 는 예약된 flag 이름이다`}},
		},
		{
			// --output 과 달리 본문 flag 는 그룹이 아니라 명령 자신에게 붙는다. 그래서
			// 같은 이름의 param 은 가려지는 게 아니라 pflag 를 패닉시킨다 — 등록 전에
			// 막아야 하는 이유다.
			//
			// 두 번째 명령이 이 케이스의 요점이다. 예약어를 상수 "data" 로 적지 않고
			// clirun.AddBodyFlag 에게 물어보기 때문에, body: 가 없어 flag 가 안 붙는
			// noBody 는 같은 param 이름을 써도 걸리지 않는다. 상수 목록이었다면 둘 다
			// 걸렸을 것이고, 그건 없는 문제를 보고하는 것이다.
			name: "본문 flag 이름과 겹치는 param 은 거부한다",
			files: map[string]string{
				"aaa.yaml": `
name: api
baseUrl: https://example.com
commands:
  - name: withBody
    method: POST
    path: /x
    body:
      required: true
    params:
      - name: data
        in: query
        type: string
  - name: noBody
    method: GET
    path: /y
    params:
      - name: data
        in: query
        type: string
`,
			},
			wantErrs:    [][]string{{"aaa.yaml 생략", `commands[0] "withBody": param "data" 는 본문 flag 이름과 겹친다`}},
			wantNotErrs: []string{"noBody"},
		},
		{
			// apis/ 에 매니페스트 아닌 파일이 있어도 된다 — README, 메모 같은 것들.
			// 이 케이스가 없으면 확장자 검사를 지우는 mutant 가 아무 테스트도 깨뜨리지
			// 않는다. pinning 은 우리가 걸어 둔 자리에만 있다.
			name: "매니페스트가 아닌 파일은 무시한다",
			files: map[string]string{
				"aaa.yaml": manifestYAML("good", "ping"),
				"zzz.txt":  "이건 매니페스트가 아니다",
			},
			wantCmds: []string{"good"},
		},
		{
			// .yaml 만 보던 시절 .yml 은 에러도 명령도 남기지 않고 사라졌다 —
			// exit 0 에 아무 말이 없어 유저가 알아낼 수 없는 조용한 유실이었다.
			name: "확장자가 .yml 이어도 받는다",
			files: map[string]string{
				"aaa.yml": manifestYAML("api", "ping"),
			},
			wantCmds: []string{"api"},
		},
		{
			// 확장자가 둘이 되면서 생긴 자리. 같은 api 이름을 주장하므로 이름 충돌
			// 검사가 그대로 잡는다 — 확장자별 특례를 따로 둘 필요가 없다.
			// aaa.yaml 이 aaa.yml 보다 먼저다 ("a.ya" < "a.ym").
			name: "같은 이름의 .yaml 과 .yml 이 있으면 충돌로 잡힌다",
			files: map[string]string{
				"aaa.yaml": manifestYAML("dup", "first"),
				"aaa.yml":  manifestYAML("dup", "second"),
			},
			wantErrs: [][]string{{"aaa.yml 생략", `name "dup" 는 이미 쓰이고 있는 명령 이름이다`}},
			wantCmds: []string{"dup"},
			wantSubs: map[string][]string{"dup": {"first"}},
		},
		{
			// .yaml 이름을 가진 디렉토리. os.ReadFile 이 내는 문구는 OS 가 쓰므로
			// (Windows "Incorrect function." / 리눅스 "is a directory") 단언하지 않고,
			// 우리가 붙인 "생략 (" 만 본다.
			name: "이름이 .yaml 인 디렉토리는 그 항목만 건너뛴다",
			files: map[string]string{
				"aaa.yaml": manifestYAML("good", "ping"),
			},
			dirs:     []string{"zzz.yaml"},
			wantErrs: [][]string{{"zzz.yaml 생략 ("}},
			wantCmds: []string{"good"},
		},
		{
			// Load 는 통과하고 Validate 가 셋을 한꺼번에 잡는다 (M6 의 "모아서 낸다").
			name: "빈 파일은 Validate 가 잡는다",
			files: map[string]string{
				"aaa.yaml": "",
			},
			wantErrs: [][]string{{
				"aaa.yaml 생략",
				"name 이 비어 있다",
				"baseURL 이 비어 있다",
				"Command 가 비어 있다",
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// t.TempDir() 안에 apis 를 따로 만든다 — noDir 케이스에서 "없는 디렉토리"를
			// 가리키려면 부모는 있고 자신만 없어야 한다.
			dir := filepath.Join(t.TempDir(), "apis")
			if !tt.noDir {
				if err := os.Mkdir(dir, 0o755); err != nil {
					t.Fatal(err)
				}
				for name, body := range tt.files {
					if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
						t.Fatal(err)
					}
				}
				for _, name := range tt.dirs {
					if err := os.Mkdir(filepath.Join(dir, name), 0o755); err != nil {
						t.Fatal(err)
					}
				}
			}

			root := newRoot()

			// LoadDir 이 붙인 것만 세려면 붙기 전 이름을 알아 둬야 한다.
			// 값은 필요 없으니 map 을 집합으로 쓴다 (M6 의 관용구).
			before := map[string]bool{}
			for _, c := range root.Commands() {
				before[c.Name()] = true
			}

			errs := LoadDir(root, dir)

			if len(errs) != len(tt.wantErrs) {
				t.Fatalf("에러 %d개, want %d개\n실제:\n%s", len(errs), len(tt.wantErrs), formatErrs(errs))
			}
			for i, frags := range tt.wantErrs {
				for _, f := range frags {
					if !strings.Contains(errs[i].Error(), f) {
						t.Errorf("errs[%d] 에 %q 가 없다\n실제: %v", i, f, errs[i])
					}
				}
			}
			for _, f := range tt.wantNotErrs {
				for i, e := range errs {
					if strings.Contains(e.Error(), f) {
						t.Errorf("errs[%d] 에 %q 가 있다 (없어야 한다)\n실제: %v", i, f, e)
					}
				}
			}

			var got []string
			for _, c := range root.Commands() {
				if !before[c.Name()] {
					got = append(got, c.Name())
				}
			}
			if !slices.Equal(got, tt.wantCmds) {
				t.Errorf("붙은 그룹 = %v, want %v", got, tt.wantCmds)
			}

			for group, want := range tt.wantSubs {
				var subs []string
				for _, c := range root.Commands() {
					if c.Name() != group {
						continue
					}
					for _, sub := range c.Commands() {
						subs = append(subs, sub.Name())
					}
				}
				if !slices.Equal(subs, want) {
					t.Errorf("%s 의 하위 명령 = %v, want %v", group, subs, want)
				}
			}
		})
	}
}

// TestCheckBodyFlags 는 생성 경로와 공유하는 검사를 직접 부른다.
//
// LoadDir 을 통해서도 밟히지만, 이건 밖으로 내보낸 함수라 계약이 따로 있다 —
// generate 가 부르는 것이 이 함수이기 때문이다. main.go 의 그 호출 자체는 여기서
// 덮이지 않는다 (package main 은 임포트할 수 없다).
func TestCheckBodyFlags(t *testing.T) {
	body := &manifest.Body{Required: true}

	tests := []struct {
		name    string
		cmds    []manifest.Command
		wantErr string
	}{
		{
			name: "본문이 없으면 같은 이름이어도 통과한다",
			cmds: []manifest.Command{{
				Name: "noBody", Method: "GET", Path: "/y",
				Params: []manifest.Param{{Name: "data", In: "query"}},
			}},
		},
		{
			name: "본문이 있으면 겹치는 param 을 잡는다",
			cmds: []manifest.Command{{
				Name: "withBody", Method: "POST", Path: "/x", Body: body,
				Params: []manifest.Param{{Name: "data", In: "query"}},
			}},
			wantErr: `commands[0] "withBody": param "data" 는 본문 flag 이름과 겹친다`,
		},
		{
			name: "겹치지 않는 param 은 통과한다",
			cmds: []manifest.Command{{
				Name: "withBody", Method: "POST", Path: "/x", Body: body,
				Params: []manifest.Param{{Name: "petId", In: "path", Required: true}},
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckBodyFlags(&manifest.Manifest{Name: "api", Commands: tt.cmds})

			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("CheckBodyFlags() = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("CheckBodyFlags() = nil, want %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("에러에 %q 가 없다: %v", tt.wantErr, err)
			}
		})
	}
}

// TestLoadDirs 는 발견 경로가 여럿일 때의 계약을 못 박는다.
//
// 합집합이다 — 앞 디렉토리를 찾았다고 거기서 끝내지 않는다. 그래서 이름이 겹칠 때
// 무슨 일이 일어나는지가 곧 계약이 된다: 앞선 쪽이 이기고, 가려진 쪽은 조용히
// 사라지지 않고 이유가 남는다. 그 판정은 여기서 새로 짠 것이 아니다 — checkGlobal 이
// 예약어를 root 에게 물어보므로(ADR-0007), 디렉토리가 늘어난 것만으로 따라온다.
//
// 환경변수는 건드리지 않는다. os.UserConfigDir 은 GOOS 마다 다른 변수를 보고
// (windows %AppData%, darwin $HOME+접미사, unix $XDG_CONFIG_HOME) macOS 에서는
// $HOME 을 바꿔도 뒤에 /Library/Application Support 가 강제로 붙는다. 그것을 조종하는
// 테스트는 OS 를 옮기는 순간 깨진다 — 그래서 환경을 읽는 층(main.apiDirs)과 목록을
// 쓰는 층(여기)을 갈랐고, 검증 가치는 후자에 있다.
func TestLoadDirs(t *testing.T) {
	cobra.EnableCommandSorting = false

	t.Run("앞선 디렉토리가 이기고 가려진 쪽은 이유가 남는다", func(t *testing.T) {
		near := makeAPIDir(t, map[string]string{"gh.yaml": manifestYAML("gh", "near")})
		far := makeAPIDir(t, map[string]string{
			"gh.yaml":   manifestYAML("gh", "far"),
			"only.yaml": manifestYAML("only", "x"),
		})

		root := newRoot()
		errs := LoadDirs(root, []string{near, far})

		// 가려진 쪽이 조용하지 않다는 것이 이 테스트의 첫 요점이다.
		if len(errs) != 1 {
			t.Fatalf("에러 %d개, want 1개\n실제:\n%s", len(errs), formatErrs(errs))
		}
		for _, frag := range []string{"gh", "이미 쓰이고 있는"} {
			if !strings.Contains(errs[0].Error(), frag) {
				t.Errorf("에러에 %q 가 없다: %v", frag, errs[0])
			}
		}

		// 뒤 디렉토리에만 있는 것도 붙는다 — 앞에서 끝냈다면 없었을 것이다.
		if findCommand(root, "only") == nil {
			t.Error("only 가 안 붙었다 — 뒤 디렉토리를 안 읽었다")
		}

		// 이긴 쪽이 어느 파일이었는지는 하위 명령으로만 구별된다.
		gh := findCommand(root, "gh")
		if gh == nil {
			t.Fatal("gh 가 안 붙었다")
		}
		if findCommand(gh, "near") == nil {
			t.Error("gh 의 하위 명령이 near 가 아니다 — 뒤 디렉토리가 이겼다")
		}
	})

	t.Run("없는 디렉토리는 건너뛴다", func(t *testing.T) {
		// 설정 디렉토리가 없는 것은 흔한 정상 상태다. 그것이 에러가 되면 아무 API 도
		// 안 쓰는 유저가 매 실행마다 경고를 본다.
		missing := filepath.Join(t.TempDir(), "없다")
		dir := makeAPIDir(t, map[string]string{"gh.yaml": manifestYAML("gh", "x")})

		root := newRoot()
		errs := LoadDirs(root, []string{missing, dir})

		if len(errs) != 0 {
			t.Fatalf("에러 %d개, want 0개\n실제:\n%s", len(errs), formatErrs(errs))
		}
		if findCommand(root, "gh") == nil {
			t.Error("없는 디렉토리 뒤의 것이 안 붙었다")
		}
	})
}

// makeAPIDir 는 파일들을 담은 디렉토리를 만들어 그 경로를 돌려준다.
func makeAPIDir(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// findCommand 는 parent 의 자식 중 이름이 name 인 것을 찾는다. 없으면 nil.
func findCommand(parent *cobra.Command, name string) *cobra.Command {
	for _, c := range parent.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}

// formatErrs 는 개수가 어긋났을 때 실제로 무엇이 왔는지 한 번에 보이기 위한 것이다.
func formatErrs(errs []error) string {
	var b strings.Builder
	for _, e := range errs {
		b.WriteString("  ")
		b.WriteString(e.Error())
		b.WriteString("\n")
	}
	return b.String()
}
