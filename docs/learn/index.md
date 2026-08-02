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
- [[go-modules]] — `go.mod`/`go.sum`, direct/indirect, `go get`/`tidy`, semver, `go` 줄은 언어 의미론 스위치
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
- [[error-wrapping]] — `%w` 는 감싸고 `%v` 는 납작하게, `errors.Join` 은 여럿을 나란히
- [[maps]] — 없는 키는 zero value, 집합 관용구, nil map 쓰기 패닉
- [[methods]] — 리시버, 값 vs 포인터, 메서드 집합이 인터페이스 만족을 정한다
- [[interfaces]] — 암묵 구현, (타입, 값) 쌍, 타입 단언, 설정은 구현이 필드로
- [[encoding-json]] — 세 층위, `Indent`/`Compact` 는 바이트만 만진다, map 경유하면 키 순서 소실
- [[bitmasks]] — `&` 로 한 비트 꺼내기, `os.FileMode` 는 uint32, C 와 다른 우선순위
- [[identifier-shadowing]] — `byte`·`len` 은 예약어가 아니다, 가려도 vet 이 안 잡는다
- [[go-test]] — 소스 옆 `_test.go`, 단언은 손으로, `Errorf` vs `Fatalf`, 규약은 다섯 줄뿐
- [[table-driven-tests]] — 익명 struct 테이블 + `t.Run`, 이름은 전부 우리가 짓는다, 필드에 함수 담기
- [[httptest]] — 목 대신 진짜 서버, `http.HandlerFunc` 는 함수를 인터페이스로, `Close` 는 기다린다
- [[text-template]] — 텍스트만 만든다(Go 를 모른다), `.` 커서와 `$` 루트, 공백 제어, `%q` 인용
- [[go-embed]] — 빌드 시점에 파일을 바이너리 안으로, 지시문은 변수 선언 바로 위
- [[go-ast]] — `format.Source` 로 검산, `parser`+`ast.Inspect` 로 생성물에서 사실 꺼내기
- [[type-aliases]] — `type X = Y` 는 이름 하나 더, `type X Y` 는 새 타입(메서드 안 따라옴)
- [[internal-packages]] — 컴파일러가 막는 경계, 모듈 vs 패키지, 좁은 façade 로 통과시키기
- [[testing-the-filesystem]] — `t.TempDir()` vs `testdata/`, OS 문구 대신 센티널, mutation testing 으로 pinning 확인
- [[go-doc]] — 선언 위 주석이 문서, 별칭은 필드를 감춘다, 문서를 reflect+파서로 붙들기
- [[absent-vs-empty]] — "안 줬다" vs "빈 것을 줬다", zero value 는 둘을 못 가른다, 표식은 값 밖에
- [[partial-decoding]] — 남의 형식에서 쓰는 자리만, 오타도 같이 조용히 버려진다, 형식 게이트는 손으로
- [[user-config-dir]] — OS 마다 다른 설정 자리, 만들어 주지도 이름 붙여 주지도 않는다, 환경을 조종하는 테스트는 OS 를 옮기면 깨진다

## 마일스톤 로그 (`milestones/`)

- [[M0]] — 프로젝트 골격 + Go 기초 warm-up
- [[M1]] — cobra 루트 명령 + 첫 외부 의존성
- [[M2]] — 매니페스트 파싱 (YAML → struct)
- [[M3]] — 매니페스트 → 명령 동적 생성
- [[M4]] — 제네릭 HTTP 실행기 (첫 실제 API 호출)
- [[M5]] — 인증과 설정 (env 토큰, 우선순위 사슬)
- [[M6]] — 매니페스트 검증 (등록 전에 막는다)
- [[M7]] — 출력 포맷 (터미널이면 pretty, 파이프면 raw)
- [[M8]] — 테스트 (ADR 세 건을 실행 가능한 판정으로)
- [[M9]] — `generate` (매니페스트가 독립 바이너리가 된다)
- [[M10]] — 요청 본문 (`--data`, 보낸다고 광고만 하던 POST 를 닫는다)
- [[M11]] — OpenAPI 임포트 (남의 명세를 읽어 매니페스트를 만든다)

## 최종 산출물 (예정)

프로젝트 완성 후, 이 위키를 **HTML 로 변환해 공개** — "AI 협업 + 학습 동시 성립"의 실증물(goal 2). *완성 시점에 착수.*
