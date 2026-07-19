# go-toolchain

`go` 명령 하나로 빌드·실행·포맷·검사·의존성을 다룬다. (`go` 를 인자 없이 치면 전체 목록.)

## 핵심

- `go run .` — 현재 패키지 컴파일 후 즉시 실행 (반복 루프용).
- `go build ./...` — `./...` = 여기부터 모든 하위 패키지. 빌드 검증.
- `gofmt -w .` — 표준 포맷 자동 정렬 (Go 는 포맷 강제).
- `go vet ./...` — 흔한 실수 정적 검사.
- `go get <경로>` / `go mod tidy` — 의존성 → [[go-modules]].
- `go list -m` — 현재 모듈 경로 출력.

## 관련

[[packages-and-main]] · [[go-modules]] · [[errors-compile-vs-runtime]]
