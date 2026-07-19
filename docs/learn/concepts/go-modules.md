# go-modules

프로젝트의 의존성 관리 체계. `go.mod`(무엇이 필요한가) + `go.sum`(받은 게 진짜인가)로 구성.

## 핵심

- **패키지 주소 = import 경로.** cobra 는 `github.com/spf13/cobra`. 우리 모듈은 `github.com/Kwangseok-Seo/cli-maker`.
- `go.mod` 의 `require`: **direct**(내 코드가 import) vs `// indirect`(딸려온 의존성의 의존성).
- `go.sum`: 다운로드 코드의 체크섬(지문) → 공급망 보안. 모듈당 두 줄(`h1:` 코드, `/go.mod`).
- 버전 `v1.10.2` = 주.부.수(semver). major=호환 깨짐.

## 명령

- `go get <경로>` — 등록(go.mod) + 봉인(go.sum).
- `go mod tidy` — go.mod 를 실제 import 와 일치 (import 후 `// indirect` 제거).

## 겪은 함정

- `gihub.com`(오타) → 호스트에 실제 요청 → HTTPS→HTTP 강등 거부. 경로는 네트워크로 해석되고 HTTPS 강제.
- import 전 `go mod tidy` → 안 쓰는 cobra 제거됨. tidy 는 import 뒤에.

## 관련

[[packages-and-main]] · [[cobra]] · [[go-toolchain]]
