package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Kwangseok-Seo/cli-maker/clirun"
	"github.com/Kwangseok-Seo/cli-maker/internal/manifest"
	"github.com/spf13/cobra"
)

// LoadDir 는 dir 의 *.yaml 매니페스트를 읽어 root 에 명령 그룹을 붙인다.
// 깨진 매니페스트는 붙이지 않고 그 이유를 반환한다 — 하나가 깨져도 나머지는 산다 (M3 격벽).
//
// 반환이 곧 보고다. 여기서 출력하지 않고, 어디에 찍을지는 호출자가 정한다.
func LoadDir(root *cobra.Command, dir string) []error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil // apis/ 가 없는 건 잘못이 아니다
		}
		return []error{err}
	}

	var errs []error
	for _, e := range entries {
		// .yaml 과 .yml 을 모두 받는다. 한쪽만 보면 다른 쪽으로 저장한 매니페스트가
		// 에러도 명령도 남기지 않고 사라진다 — exit 0 에 아무 말이 없어 유저가
		// 왜 명령이 없는지 알아낼 방법이 없다.
		if ext := filepath.Ext(e.Name()); ext != ".yaml" && ext != ".yml" {
			continue
		}
		path := filepath.Join(dir, e.Name())

		m, err := manifest.Load(path)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s 생략 (%w)", path, err))
			continue
		}

		// 국소 검증: 매니페스트 하나만 보면 판정 가능한 것.
		if err := manifest.Validate(m); err != nil {
			errs = append(errs, fmt.Errorf("%s 생략\n%w", path, err))
			continue
		}

		// 전역 검증: CLI 전역 표면을 알아야 판정 가능한 것.
		// 그룹을 먼저 지어서 함께 넘긴다 — 예약된 flag 이름의 출처가 그룹이기 때문이다.
		group := Build(m)
		if err := checkGlobal(root, group, m); err != nil {
			errs = append(errs, fmt.Errorf("%s 생략\n%w", path, err))
			continue
		}

		root.AddCommand(group)
	}

	return errs
}

// LoadDirs 는 dirs 를 앞에서부터 순회하며 각 디렉토리를 LoadDir 한다.
//
// 앞선 디렉토리가 이긴다. 뒤 디렉토리에 같은 이름의 매니페스트가 또 있으면 checkGlobal 의
// reservedCommand 가 "이미 쓰이고 있는 명령 이름이다" 로 걸러낸다 — 예약어를 상수로 적지
// 않고 root 에게 물어보기 때문에(M6) 디렉토리가 늘어도 여기서 중복을 따로 볼 필요가 없다.
//
// 없는 디렉토리는 잘못이 아니다 — LoadDir 이 이미 os.ErrNotExist 를 nil 로 흘린다.
func LoadDirs(root *cobra.Command, dirs []string) []error {
	var errs []error
	for _, dir := range dirs {
		errs = append(errs, LoadDir(root, dir)...)
	}
	return errs
}

// checkGlobal 은 이미 만들어진 CLI 표면과 충돌하는지 본다.
// 예약어 목록을 상수로 두지 않고 root 에 직접 물어본다 — 명령이나 flag 가 늘어도 저절로 따라온다.
func checkGlobal(root, group *cobra.Command, m *manifest.Manifest) error {
	var errs []error

	if reservedCommand(root, m.Name) {
		errs = append(errs, fmt.Errorf("name %q 는 이미 쓰이고 있는 명령 이름이다", m.Name))
	}

	for i, c := range m.Commands {
		for _, p := range c.Params {
			// help 는 cobra 가 명령마다 자동으로 붙이는 flag 라 Lookup 으로는 안 잡힌다.
			if p.Name == "help" || group.PersistentFlags().Lookup(p.Name) != nil {
				errs = append(errs, fmt.Errorf("commands[%d] %q: param %q 는 예약된 flag 이름이다", i, c.Name, p.Name))
			}
		}
	}

	// 이 하나만 생성 경로와 공유한다 — 아래 주석 참조. nil 은 Join 이 무시한다.
	errs = append(errs, CheckBodyFlags(m))

	return errors.Join(errs...)
}

// CheckBodyFlags 는 param 이름이 cli-maker 가 그 명령에 다는 본문 flag 와 겹치는지 본다.
//
// checkGlobal 안에 두지 않고 따로 뗀 이유는 **이 검사만 생성 경로에도 필요**하기
// 때문이다. 매니페스트 이름 충돌과 그룹 persistent flag 는 이 CLI 의 전역 표면에 관한
// 것이라 생성물과 무관하지만(ADR-0007), 본문 flag 는 명령 자신에게 붙고 생성된 CLI 도
// 같은 clirun.AddBodyFlag 로 그것을 단다. 겹친 채 통과시키면 실행 시점에 pflag 가
// 패닉하는 소스가 나간다 — 생성기는 exit 0 으로 끝나므로 조용하다.
//
// 예약된 이름을 상수로 적지 않고 등록 함수에게 물어본다 — 빈 명령에 같은 Body 로
// 달아 보고 무엇이 붙었는지 본다. 이름이 바뀌거나 늘어도 저절로 따라오고, Body 가
// nil 인 명령은 아무것도 안 붙으므로 같은 param 이름을 써도 걸리지 않는다.
func CheckBodyFlags(m *manifest.Manifest) error {
	var errs []error

	for i, c := range m.Commands {
		if c.Body == nil {
			continue
		}
		probe := &cobra.Command{Use: "probe"}
		clirun.AddBodyFlag(probe, c.Body)

		for _, p := range c.Params {
			if probe.Flags().Lookup(p.Name) != nil {
				errs = append(errs, fmt.Errorf("commands[%d] %q: param %q 는 본문 flag 이름과 겹친다", i, c.Name, p.Name))
			}
		}
	}

	return errors.Join(errs...)
}

// reservedCommand 은 root 에 이미 붙어 있거나 cobra 가 나중에 붙일 명령 이름인지 본다.
// 앞서 등록된 매니페스트도 root 의 자식이므로, 매니페스트끼리의 이름 충돌이 여기서 함께 걸린다.
func reservedCommand(root *cobra.Command, name string) bool {
	// 이 둘은 cobra 가 Execute 시점에 붙여서 지금은 root.Commands() 에 없다.
	if name == "help" || name == "completion" {
		return true
	}
	for _, sub := range root.Commands() {
		if sub.Name() == name {
			return true
		}
	}
	return false
}
