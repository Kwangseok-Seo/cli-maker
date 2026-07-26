package clirun

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
)

const (
	// TimeoutFlag 는 API 그룹에 persistent flag 로 등록되는 이름이자, 여기서 읽는 이름이다
	// (생성된 CLI 에선 그 루트가 곧 그룹이다). 같은 이름의 param 은 M6 부터 등록 시점에
	// 거부되므로 로컬 flag 가 이걸 가릴 일은 없다 — 다만 shorthand(-o)는 여전히 안 본다.
	TimeoutFlag = "timeout"

	defaultTimeout = 30 * time.Second
	timeoutEnv     = "CLI_MAKER_TIMEOUT"
)

// resolveTimeout 은 세 층에서 타임아웃을 정한다 — 좁은 수명이 넓은 수명을 이긴다.
//
//  1. --timeout flag   (이번 실행 한 번)
//  2. CLI_MAKER_TIMEOUT 환경변수 (이 셸 세션)
//  3. defaultTimeout   (프로그램에 박힌 값)
//
// 각 층은 "값"만으로는 부족하고 "유저가 실제로 줬는가"를 함께 물어야 한다.
// flag 는 cmd.Flags().Changed(이름), 환경변수는 os.LookupEnv 의 두 번째 반환값이 그 답을 준다.
func resolveTimeout(cmd *cobra.Command) (time.Duration, error) {
	if cmd.Flags().Changed(TimeoutFlag) {
		return cmd.Flags().GetDuration(TimeoutFlag)
	}
	t, ok := os.LookupEnv(timeoutEnv)
	if ok {
		d, err := time.ParseDuration(t)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", timeoutEnv, err)
		}
		return d, nil
	}

	return defaultTimeout, nil
}
