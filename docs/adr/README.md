# 아키텍처 결정 기록 (ADR)

이 repo 의 확정된 설계 결정 색인. 코드/설계를 바꾸기 전에 관련 결정을 먼저 확인한다. 번복은 침묵 덮어쓰기 금지 — supersede ADR 로.

| # | 제목 | 상태 |
|---|------|------|
| [0001](0001-runtime-interpreter-architecture.md) | 런타임 인터프리터 아키텍처 | accepted |
| [0002](0002-manifest-collections-as-ordered-lists.md) | 매니페스트 컬렉션은 name 필드를 가진 순서 있는 리스트 | accepted |
| [0003](0003-dynamic-command-surface.md) | 동적 명령 표면: 매니페스트 하나 = 2단 그룹, apis/ 스캔으로 발견 | accepted |
| [0004](0004-params-as-flags.md) | 모든 Param 은 flag 로 노출한다 | accepted |
| [0005](0005-http-error-surface.md) | HTTP 4xx/5xx 는 본문을 내보낸 뒤 종료 코드 1 로 알린다 | accepted |
| [0006](0006-missing-token-policy.md) | 토큰이 없으면 경고만 남기고 인증 없이 보낸다 | accepted |
| [0007](0007-load-time-validation.md) | 매니페스트는 등록 전에 검증하고, 문제가 있으면 그 파일을 통째로 건너뛴다 | accepted |
| [0008](0008-output-format-defaults-to-tty.md) | 출력 포맷의 기본값은 stdout 이 터미널인지로 정한다 | accepted |
| [0009](0009-generated-cli-shape.md) | 생성된 CLI 는 얇은 공개 façade 를 부르고, 명령 표면은 펼쳐서 낸다 | accepted |
| [0010](0010-request-body-as-command-field.md) | 요청 본문은 Param 이 아니라 Command 의 선택적 필드다 | accepted |
| [0011](0011-importer-reads-five-places-directly.md) | 임포터는 Spec 의 다섯 자리만 직접 읽는다 | accepted |
| [0012](0012-import-output-is-a-deterministic-draft.md) | 임포트 산출물은 저자 순서를 버리고, 못 옮기는 것은 required 여부로 가른다 | accepted |
