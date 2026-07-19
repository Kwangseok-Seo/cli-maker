# 00 — Go 기초

> 로드맵: M0(골격) + 첫 Go 문법 warm-up · 커밋: bootstrap ~ M1

cli-maker 골격을 만들고, cobra 로 가기 전에 익힌 Go 언어의 밑바닥.

## 한 줄 요약

`main.go` 한 파일을 읽고·고치고·실행하며 패키지 / 함수 / 변수 / 슬라이스 / 에러라는 Go 의 뼈대를 익혔다.

## 배운 개념 (이론/문법)

**프로그램의 뼈대**
- `package main` — 모든 Go 파일은 패키지에 속한다. `main` 은 특별해서 실행 파일이 된다.
- `func main()` — 프로그램의 진입점. 인자·반환 없음.
- `import "fmt"` — 표준 라이브러리 패키지 가져오기. 여러 개면 `import ( ... )` 로 묶는다.
- `fmt.Println(...)` — `패키지.함수` 문법으로 호출.
- **대문자로 시작 = 공개(exported), 소문자 = 비공개.** `Println` 이 대문자 P 인 이유.

**변수**
```go
greeting := "hello"        // 선언 + 대입 (타입 자동 추론, 함수 안에서만)
greeting = "안녕"           // 이미 있는 변수에 재대입 (:= 아님)
var name string = "world"  // var 로 명시 선언
var n int                  // 값 없으면 제로 값 (int→0, string→"", bool→false)
```
- `:=` 는 declare+assign, `=` 는 assign only. 없는 변수에 `=` 만 쓰면 `undefined` 에러.
- **선언하고 안 쓰면 컴파일 에러** (`declared and not used`). Go 의 죽은 코드 봉쇄.
- 문자열은 `+` 로 이어붙인다.

**슬라이스와 명령줄 인자**
```go
os.Args        // []string — 명령줄 인자들의 리스트(슬라이스)
os.Args[0]     // 프로그램 경로 (go run 은 임시 exe 경로)
os.Args[1]     // 사용자가 친 첫 인자 (인덱스는 0부터)
```

**두 종류의 에러 (핵심 구분)**
- **컴파일 에러** — 실행 *전에* 걸림. 예: `undefined: greeting`. 형식은 `파일:줄:열`, 여러 개를 한 번에 보고.
- **런타임 패닉** — 실행 *중에* 터짐. 예: 인자 없이 `os.Args[1]` → `panic: runtime error: index out of range [1] with length 1`.

## 손에 익힌 명령/도구

- `go run .` — 현재 패키지 컴파일 후 즉시 실행 (반복 루프용).
- `go build ./...` — `./...` = 여기부터 모든 하위 패키지. 빌드 검증.
- `gofmt -w .` — 표준 포맷으로 자동 정렬 (Go 는 포맷이 강제됨).
- `go vet ./...` — 흔한 실수 정적 검사.
- `go` (인자 없이) — 서브커맨드 목록을 스스로 알려줌.

## 겪은 실수 · 함정

- `greeting = "hello"` 를 첫 줄에 씀 → `undefined: greeting`. **`=` 는 없는 변수를 못 만든다** (`:=` 필요).
- 인자 없이 `os.Args[1]` 접근 → 런타임 패닉. **인덱스 전에 존재 여부 확인 필요** — 이 고통이 cobra 도입의 동기가 됐다.

## 관련 코드/결정

- 코드: `main.go` (초기 버전 → os.Args 실험).
- 결정: `docs/adr/0001-runtime-interpreter-architecture.md`.
