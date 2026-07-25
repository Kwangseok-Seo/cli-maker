package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

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
		if filepath.Ext(e.Name()) != ".yaml" {
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
		if err := checkGlobal(root, m); err != nil {
			errs = append(errs, fmt.Errorf("%s 생략\n%w", path, err))
			continue
		}

		root.AddCommand(Build(m))
	}

	return errs
}

// checkGlobal 은 이미 만들어진 CLI 표면과 충돌하는지 본다.
// 예약어 목록을 상수로 두지 않고 root 에 직접 물어본다 — 명령이나 flag 가 늘어도 저절로 따라온다.
func checkGlobal(root *cobra.Command, m *manifest.Manifest) error {
	var errs []error

	if reservedCommand(root, m.Name) {
		errs = append(errs, fmt.Errorf("name %q 는 이미 쓰이고 있는 명령 이름이다", m.Name))
	}

	for i, c := range m.Commands {
		for _, p := range c.Params {
			// help 는 cobra 가 명령마다 자동으로 붙이는 flag 라 Lookup 으로는 안 잡힌다.
			if p.Name == "help" || root.PersistentFlags().Lookup(p.Name) != nil {
				errs = append(errs, fmt.Errorf("commands[%d] %q: param %q 는 예약된 flag 이름이다", i, c.Name, p.Name))
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
