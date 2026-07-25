# go-toolchain

`go` 명령 하나로 빌드·실행·포맷·검사·의존성을 다룬다. (`go` 를 인자 없이 치면 전체 목록.)

## 핵심

- `go run .` — 현재 패키지 컴파일 후 즉시 실행 (반복 루프용).
- `go build ./...` — `./...` = 여기부터 모든 하위 패키지. 빌드 검증.
- `gofmt -w .` — 표준 포맷 자동 정렬 (Go 는 포맷 강제).
- `go vet ./...` — 흔한 실수 정적 검사. **`go build` 가 통과해도 따로 돌린다** — 버려진 `fmt.Errorf` 결과처럼 문법은 멀쩡하고 의미만 틀린 것을 잡는다 → [[unused-results]].
- `go test ./...` — 테스트 실행. `-run <이름>` 으로 하나만, `-v` 로 `t.Logf` 까지 본다.
- `go get <경로>` / `go mod tidy` — 의존성 → [[go-modules]].
- `go list -m` — 현재 모듈 경로 출력.

## 겪은 함정

- **CRLF vs LF**: Windows 편집기로 새로 만든 `.go` 는 CRLF(`\r\n`)로 저장돼 `gofmt -l .` 에 걸린다 — Go 소스는 OS 불문 **LF 가 정본**. `gofmt -d` 가 "내용은 같은데 전 줄이 바뀜"으로 보이면 줄바꿈 차이 신호. 즉시 고치기는 `gofmt -w .`. **재발 방지는 층이 둘**: `.gitattributes`(`*.go text eol=lf`)는 git 커밋/체크아웃 경계에서 정규화하고, 편집기가 애초에 LF 로 쓰게 하려면 `.editorconfig`(`[*.go] end_of_line = lf`)가 필요하다 — `.gitattributes` 만으론 편집기가 새 파일에 쓰는 CRLF 를 못 막아 gofmt(작업 트리를 읽음)가 여전히 걸린다.
- `gofmt -w` 는 **경로 인자 필수**. 인자 없이 치면 표준입력을 읽어 `-w` 와 충돌(`cannot use -w with standard input`) → `gofmt -w .`.

## 관련

[[packages-and-main]] · [[go-modules]] · [[errors-compile-vs-runtime]] · [[stdout-stderr]] · [[unused-results]]
