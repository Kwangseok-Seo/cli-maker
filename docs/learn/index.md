# cli-maker 학습 위키 — 색인

cli-maker 를 만들며 쌓은 **Go·도구 지식**의 위키. 개념을 찾을 땐 **이 색인에서 먼저 고르고 해당 페이지만 읽는다** (Karpathy LLM-wiki 패턴 — 토큰 효율). 각 개념 페이지는 `[[wikilink]]` 로 서로 연결돼 그래프를 이룬다.

> 도메인 용어 → [`/CONTEXT.md`](../../CONTEXT.md) · 아키텍처 결정 → [`/docs/adr/`](../adr/) · **여기는 언어·도구 학습 전용.**

## 개념 (`concepts/`)

- [[packages-and-main]] — 패키지 · `func main` · import · exported(대문자) 규칙
- [[variables]] — `:=` vs `=` vs `var`, zero value, 미사용 변수 금지
- [[slices-and-args]] — `[]string`, `os.Args`, 0-based 인덱싱
- [[errors-compile-vs-runtime]] — 컴파일 에러(undefined) vs 런타임 패닉
- [[structs]] — 이름 붙은 칸(필드)의 묶음
- [[pointers]] — `&` 참조
- [[functions-as-values]] — 함수를 값으로, struct 칸에 담기
- [[go-modules]] — `go.mod`/`go.sum`, direct/indirect, `go get`/`tidy`, semver
- [[cobra]] — `cobra.Command` struct, 자동 `--help`
- [[error-handling]] — 에러는 값, `if err != nil`, `os.Exit`
- [[go-toolchain]] — `go run`/`build`/`fmt`/`vet`/`get` 명령

## 마일스톤 로그 (`milestones/`)

- [[M0]] — 프로젝트 골격 + Go 기초 warm-up
- [[M1]] — cobra 루트 명령 + 첫 외부 의존성

## 최종 산출물 (예정)

프로젝트 완성 후, 이 위키를 **HTML 로 변환해 공개** — "AI 협업 + 학습 동시 성립"의 실증물(goal 2). *완성 시점에 착수.*
