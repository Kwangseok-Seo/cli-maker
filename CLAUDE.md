# cli-maker — 프로젝트 지침

여러 web API 를 하나의 CLI 로 다루게 해 주는 오픈소스 도구. 유저가 API 명세(매니페스트)를 주면, 런타임에 그걸 읽어 서브커맨드를 동적으로 노출한다. 지향점은 mvanhorn/cli-printing-press (단, 그쪽은 코드 생성 방식 — 우리는 런타임 인터프리터로 시작).

## 이 프로젝트의 두 목표 (둘 다 1급)

1. **도구**: web API 용 CLI 를 만들어 유저가 여러 API 를 CLI 로 쓰게 한다.
2. **학습 실증**: 이 도구를 Go 로 만들어 가는 과정에서 Go·프레임워크를 배우고, 그 과정 자체를 오픈소스로 보여 "AI 에이전트와 협업하면서 학습이 동시에 성립함"을 실증한다.

목표 2 는 코드에 드러나지 않는 핵심 제약이다. 아래 워크플로가 그것을 강제한다.

## 학습-우선 워크플로 (반드시 준수)

- **매 단계마다 이론 학습 대화를 먼저.** 새 기능/개념을 구현하기 전에 관련 Go·프레임워크 이론(왜 이렇게 하는지, idiomatic 한 이유)을 대화로 설명한다.
- **덤프 금지, 공동 작성 지향.** 완성된 코드를 한 번에 쏟지 않는다. 가능하면 사용자가 이해하고 직접(또는 함께) 작성하도록 유도한다 — Socratic 하게.
- **작은 단계로.** README 의 "학습 로드맵" 마일스톤 단위로 진행. 한 마일스톤 = 한 Go 개념 + 한 기능 조각.
- **학습 위키 갱신.** 각 마일스톤마다 배운 개념을 `docs/learn/concepts/<개념>.md` (원자 페이지, `[[wikilink]]` ≥2, "겪은 함정" 포함)에 쓰고, `docs/learn/index.md` 색인과 `docs/learn/milestones/M*.md` 여정 로그를 갱신한다 (Karpathy LLM-wiki 하이브리드). 이 관행 자체가 goal 2 의 산출물이며 skill 승격 후보. 최종엔 이 위키를 HTML 로 변환(프로젝트 완성 후).
- **패턴을 승격.** 반복되는 학습 패턴이 보이면 rule/지침/skill 후보로 제안한다 (부수 목표).
- **정직하게.** "동작합니다" 대신 실제 input→output 을 보여준다 (전역 report-guideline).

## 아키텍처

- **런타임 인터프리터**: 단일 바이너리가 매니페스트(데이터)를 런타임에 읽어 cobra 명령을 동적 생성. 새 API = 매니페스트 추가(재컴파일 없음). 근거·거부한 대안(코드 생성)은 `docs/adr/0001-runtime-interpreter-architecture.md`.
- **CLI 프레임워크**: spf13/cobra (M1 에서 도입).
- **입력 형식**: 커스텀 매니페스트(YAML). 스키마는 M1~M2 에서 확정. OpenAPI 임포트는 후속.
- **코드 생성(`generate`)**: M9 에서 얹었다. 매니페스트 → 독립 모듈(`main.go` + `go.mod`), 생성 코드는 공개 façade `clirun` 만 부른다. 근거·거부한 대안은 `docs/adr/0009-generated-cli-shape.md`.

## 레이아웃 (자라나는 대로 추가 — 빈 디렉토리 선점 금지)

- `main.go` — 얇은 진입점.
- `cmd/` — cobra 명령 정의 (M1~).
- `internal/` — 내부 구현(외부 임포트 불가). `internal/manifest/`(스키마·파싱), `internal/executor/`(HTTP 실행기), `internal/generate/`(템플릿·코드 생성) 등 (M2~).
- `clirun/` — **공개 façade**. 생성된 CLI 가 부르는 유일한 표면(타입 별칭 + `Run`). 여기 있는 것만 외부 계약이므로 넓히기 전에 한 번 더 생각한다 (M9~).
- `apis/` — 유저 매니페스트. `example.yaml` = 초안 스키마 스케치.
- `docs/adr/` — 아키텍처 결정 기록.
- `CONTEXT.md` — 도메인 용어집(glossary 전용).

## 명령

- 빌드: `go build ./...`
- 실행: `go run .`
- 테스트: `go test ./...`
- 포맷/정적검사: `gofmt -l .`, `go vet ./...`

## 문서 권위

- **결정** → `docs/adr/NNNN-*.md` (Y-statement 1문단; 포맷은 전역 rule `adr-context-format`).
- **용어** → `CONTEXT.md` (glossary 전용 — 규칙/결과/파라미터 넣지 않음).
- **진행 상태·로드맵** → `README.md` 학습 로드맵.
- **언어·도구 학습 지식** → `docs/learn/` (`index.md` 색인 + `concepts/` 개념 위키 + `milestones/` 여정 로그. 도메인 용어(CONTEXT)·결정(ADR)과 구분되는 세 번째 축).

## Git

- 기본 브랜치 `main`. 커밋/푸시는 사용자가 요청할 때만.
