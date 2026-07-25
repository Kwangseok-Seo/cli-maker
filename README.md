# cli-maker

여러 web API 를 **하나의 CLI** 로 다루게 해 주는 오픈소스 도구. API 명세(매니페스트)를 주면, cli-maker 가 런타임에 그걸 읽어 서브커맨드를 만들어 냅니다.

> 지향점: [mvanhorn/cli-printing-press](https://github.com/mvanhorn/cli-printing-press). 단, printing press 가 API마다 독립 바이너리를 *코드 생성* 하는 것과 달리, cli-maker 는 단일 바이너리가 매니페스트를 *런타임 해석* 합니다. 이유: [ADR-0001](docs/adr/0001-runtime-interpreter-architecture.md).

## 두 가지 목표

1. **도구** — web API 용 CLI 생성기.
2. **학습 실증** — 이 도구를 Go 로 만들어 가며 Go·프레임워크를 배우고, 그 과정을 오픈소스로 공개해 *AI 에이전트와 협업하면서 학습이 동시에 성립함* 을 보인다.

그래서 이 저장소는 "완성된 코드"가 아니라 **배우며 자라는 과정**을 커밋 히스토리로 남깁니다.

## 상태

M4 완료 — 매니페스트로 만들어진 명령이 **실제 API 를 호출**합니다.

```
$ cli-maker --help
Available Commands:
  greet       이름을 받아 인사한다
  parse       파일을 받아 parse 한다
  petstore    petstore API        ← apis/example.yaml 에서 생성된 명령

$ cli-maker petstore getPetById --help
GET /pet/{petId}
Flags:
      --petId string   path - int

$ cli-maker petstore getPetById --petId 10
{"id":10,"name":"Rufus","photoUrls":["example.com"],"tags":[]}

$ cli-maker petstore getPetById --petId abc
{"code":400,"message":"Input error: couldn't convert `abc` to type `class java.lang.Long`"}
Error: HTTP 400 Bad Request        ← stderr, 종료 코드 1
```

매니페스트의 Param 은 모두 flag 로 드러납니다(`--이름 값`). 본문은 stdout 으로만 나가므로 `| jq` 가 안전하고, 4xx/5xx 는 종료 코드 1 로 알립니다.

`apis/` 에 매니페스트 YAML 을 하나 더 떨어뜨리면 **재컴파일 없이** 새 API 명령이 생깁니다 ([ADR-0003](docs/adr/0003-dynamic-command-surface.md)).

## 개발

- 빌드: `go build ./...`
- 실행: `go run .`
- 테스트: `go test ./...`

Go 1.26+ 필요.

## 학습 로드맵

각 마일스톤 = 배우는 Go 개념 하나 + 기능 조각 하나.

| 마일스톤 | 기능 | 배우는 개념 |
|---|---|---|
| **M0** ✅ | 프로젝트 골격, `go run` | 모듈 시스템, 패키지, 진입점 |
| **M1** ✅ | Cobra 루트 + 첫 서브커맨드 | 의존성 관리(go get/go.sum), 명령 트리, 인터페이스 |
| **M2** ✅ | 매니페스트 파싱 (YAML→struct) | struct 태그, 언마샬, 에러 처리 |
| **M3** ✅ | 매니페스트→명령 동적 생성 | 클로저 캡처, 컴포지트 리터럴, 디렉토리 스캔 |
| **M4** ✅ | 제네릭 HTTP 실행기 | net/http, io.Reader/Writer, defer, context |
| **M5** | 인증·설정 (env 토큰) | os 패키지, 설정 우선순위 |
| **M6** | 출력 포맷 (`--json`/`--compact`) | encoding/json, 인터페이스 |
| **M7** | 테스트 | table-driven test, httptest |
| 이후 | `generate` (코드 생성) | text/template |

> 각 마일스톤에서 쌓은 이론·문법·겪은 함정은 [`docs/learn/`](docs/learn/) 에 개념별 지식베이스로 정리합니다 — "무엇을 배웠나"의 실증.

## 라이선스

MIT — [LICENSE](LICENSE).
