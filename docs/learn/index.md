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
- [[closures]] — 함수가 붙잡는 건 값이 아니라 칸, Go 1.22 루프 변수 변경점
- [[composite-literals]] — `&T{...}` 리터럴 vs 블록, 필드:값만
- [[go-modules]] — `go.mod`/`go.sum`, direct/indirect, `go get`/`tidy`, semver
- [[cobra]] — `cobra.Command` struct, 자동 `--help`
- [[error-handling]] — 에러는 값, `if err != nil`, `os.Exit`
- [[go-toolchain]] — `go run`/`build`/`fmt`/`vet`/`get` 명령
- [[file-io]] — `os.ReadFile`, 파일 통째로 `[]byte`
- [[serialization]] — 마샬/언마샬, `yaml.Unmarshal(&m)`, decode-into 패턴
- [[struct-tags]] — `yaml:"키"`, exported 필드, camelCase 함정
- [[stdout-stderr]] — 데이터 vs 진단, `fmt.Fprintln(os.Stderr, …)`
- [[net-http]] — 요청 생성/전송 분리, 상태 코드는 에러가 아니다
- [[io-reader-writer]] — 메서드 하나짜리 인터페이스, `io.Copy` 의 32KB 버퍼
- [[defer]] — 함수 끝에 예약, 수신자는 `defer` 시점에 평가
- [[context]] — 취소·마감의 전파, `WithTimeout` + `defer cancel()`
- [[url-encoding]] — 경로(`%20`)와 쿼리(`+`)는 규칙이 다르다
- [[unused-results]] — 반환값을 버리면 조용히 아무 일도 안 일어난다
- [[environment-variables]] — `os.LookupEnv` 의 `ok`, 비밀은 매니페스트가 아니라 env 에
- [[http-headers]] — `map[string][]string`, Set vs Add, 생성 뒤 전송 전
- [[config-precedence]] — 좁은 수명이 이긴다, 기본값은 "안 줬다"를 지운다
- [[error-wrapping]] — `%w` 는 감싸고 `%v` 는 납작하게, 진단은 출처를 말한다

## 마일스톤 로그 (`milestones/`)

- [[M0]] — 프로젝트 골격 + Go 기초 warm-up
- [[M1]] — cobra 루트 명령 + 첫 외부 의존성
- [[M2]] — 매니페스트 파싱 (YAML → struct)
- [[M3]] — 매니페스트 → 명령 동적 생성
- [[M4]] — 제네릭 HTTP 실행기 (첫 실제 API 호출)
- [[M5]] — 인증과 설정 (env 토큰, 우선순위 사슬)

## 최종 산출물 (예정)

프로젝트 완성 후, 이 위키를 **HTML 로 변환해 공개** — "AI 협업 + 학습 동시 성립"의 실증물(goal 2). *완성 시점에 착수.*
