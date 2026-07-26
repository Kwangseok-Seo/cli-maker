package cli

import (
	"fmt"
	"io"
	"os"

	"github.com/Kwangseok-Seo/cli-maker/internal/format"
	"github.com/spf13/cobra"
)

const (
	// OutputFlag 는 루트에 persistent flag 로 등록되는 이름이자, 여기서 읽는 이름이다.
	OutputFlag = "output"

	// outputAuto 는 기본값. "유저가 고르지 않았다" 를 뜻하고, 이때만 stdout 이
	// 터미널인지 보고 대신 정한다.
	outputAuto = "auto"
)

// resolveFormatter 는 --output 값으로 Formatter 를 고른다.
//
//	auto     stdout 이 터미널이면 Pretty, 파이프·리다이렉트면 Raw
//	raw      받은 바이트 그대로
//	pretty   두 칸 들여쓰기
//	compact  공백 제거
//
// auto 로 고른 Pretty 는 Warn 을 비워 둔다 — 유저가 요청한 적 없는 포맷이
// 실패한 것이라 경고할 이유가 없다. 명시한 pretty/compact 에만 Warn 을 채운다.
func resolveFormatter(cmd *cobra.Command) (format.Formatter, error) {
	oFlag, err := cmd.Flags().GetString(OutputFlag)
	if err != nil {
		return nil, err
	}

	switch oFlag {
	case outputAuto:
		if isTerminal(cmd.OutOrStdout()) {
			return format.Pretty{}, nil
		}
		return format.Raw{}, nil
	case "raw":
		return format.Raw{}, nil
	case "pretty":
		return format.Pretty{Warn: cmd.ErrOrStderr()}, nil
	case "compact":
		return format.Compact{Warn: cmd.ErrOrStderr()}, nil
	default:
		return nil, fmt.Errorf("--%s %q 는 지원하지 않는다 (auto|raw|pretty|compact)", OutputFlag, oFlag)
	}
}

// isTerminal 은 w 가 콘솔인지 본다.
//
// w 는 io.Writer 로만 알려져 있다 — Write 메서드가 있다는 것 외엔 아무것도 모른다.
// 파일 모드를 물어보려면 실제 타입이 *os.File 인지 확인해 껍질을 벗겨야 한다.
// 그게 타입 단언이고, 두 값을 받는 형태를 쓰면 아니어도 패닉하지 않는다:
//
//	f, ok := w.(*os.File)
//
// 실측(Windows): 콘솔 Dcrw-rw-rw- / 파이프 prw-rw-rw- / 리다이렉트 -rw-rw-rw-
// → os.ModeCharDevice 비트가 셋을 가른다.
func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return ok
	}

	fInfo, err := f.Stat()
	if err != nil {
		return false
	}

	return fInfo.Mode()&os.ModeCharDevice != 0
}
