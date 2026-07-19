# 01 — Cobra 와 외부 의존성 (M1)

> 로드맵: M1 · 커밋: `feat: add cobra root command (M1)`

`os.Args` 손파싱을 걷어내고 cobra 루트 명령으로 교체. 첫 외부 라이브러리를 프로젝트에 들이는 전 과정.

## 한 줄 요약

struct · 포인터 · 함수값을 배우고, `go get` → import → `go mod tidy` 흐름으로 cobra 를 들여 `--help` 가 자동 생성되는 걸 확인했다.

## 배운 개념 (이론/문법)

**struct (구조체)** — 이름 붙은 칸들의 묶음. 프로젝트 전체(Manifest, Command)의 토대.
```go
type Point struct { X int; Y int }
p := Point{X: 3, Y: 5}
p.X  // 3  (점으로 칸 접근)
```

**포인터 `&`** — 값에 대한 참조(주소). `&cobra.Command{...}` = 복사본이 아니라 원본을 가리킴. cobra 가 명령을 공유·수정해야 해서 포인터를 요구. (심화는 후속 마일스톤.)

**함수를 값으로** — struct 칸에 함수를 담을 수 있다.
```go
Run: func(cmd *cobra.Command, args []string) { ... }  // cobra 가 실행 시 호출
```
`args []string` = 프로그램 이름·플래그를 걸러낸 깔끔한 인자 (os.Args 손파싱 불필요).

**메서드 호출** — `rootCmd.Execute()` : cobra 가 os.Args 를 읽어 맞는 `Run` 을 부른다.

**외부 의존성**
- **패키지 주소 = import 경로.** cobra 는 `github.com/spf13/cobra` 에 살고 그대로 import. 우리 모듈도 `github.com/Kwangseok-Seo/cli-maker`.
- `go.mod` 의 `require` = 필요 모듈 선언. **direct**(내 코드가 import) vs **`// indirect`**(딸려온 의존성의 의존성).
- `go.sum` = 다운로드 코드의 **체크섬(지문)**. 변조 탐지 = 공급망 보안. 모듈당 두 줄: `h1:`(코드 전체) + `/go.mod`(그 모듈의 go.mod).
- 시맨틱 버저닝 `v1.10.2` = 주.부.수 (major=호환깨짐 / minor=기능추가 / patch=버그수정).

**cobra 루트 명령**
```go
rootCmd := &cobra.Command{
    Use:   "cli-maker",
    Short: "...",
    Run:   func(cmd *cobra.Command, args []string) { ... },
}
rootCmd.Execute()
```
`Use`/`Short` 묘사만으로 `--help`·`-h` 가 자동 생성됨 (선언적 API).

## 손에 익힌 명령/도구

- `go get github.com/spf13/cobra` — 외부 모듈 등록(go.mod) + 봉인(go.sum). 딸린 의존성도 자동.
- `go mod tidy` — go.mod/go.sum 을 **실제 import 와 일치**시킴. import 후 돌리면 cobra 의 `// indirect` 가 떨어짐.
- `go run . --help` — cobra 자동 도움말.

## 겪은 실수 · 함정

- `gihub.com`(오타) → Go 가 그 호스트에 실제 요청 → 광고 페이지로 HTTPS→HTTP 강등 → **거부**. 교훈: 모듈 경로는 네트워크로 해석되고, Go 는 HTTPS 를 강제한다.
- import 하기 *전에* `go mod tidy` 를 돌리면 cobra 가 통째로 제거됨(쓰는 코드 없음). **tidy 는 import 뒤에.**

## 관련 코드/결정

- 코드: `main.go` (cobra 루트), `go.mod` / `go.sum`.
- 결정: `docs/adr/0001-runtime-interpreter-architecture.md` (cobra 채택 맥락).
