# cli-maker

유저가 준 API 명세를 런타임에 읽어, 여러 web API 를 하나의 CLI 로 노출하는 도구의 도메인 언어.

## Language

**Manifest**:
우리 커스텀 형식(YAML)으로 기술한 하나의 API 정의 — base URL, 인증, 명령 목록을 담는다.
_Avoid_: Spec (= OpenAPI 를 가리킬 때만), config

**Command**:
매니페스트 안의 한 항목 — HTTP 메서드·경로·파라미터로 기술된, 유저가 CLI 에서 호출할 하나의 동작.
_Avoid_: endpoint (= 서버 측 URL), action

**Param**:
한 Command 이 받는 **이름 붙은** 입력 하나 — path/query/header 중 어디에 실리는지와 타입을 가진다.
_Avoid_: argument, flag (= CLI 표면의 표현일 뿐), body param (= 이름이 없으므로 Param 이 아니다)

**Body**:
한 Command 이 요청에 싣는 본문 — 이름이 없고, 명령당 최대 하나이며, 아예 없을 수도 있다.
_Avoid_: payload, data, body param

**Executor**:
모든 매니페스트·모든 Command 이 공유하는 단 하나의 제네릭 실행기 — Command 데이터로 HTTP 요청을 만들어 보내고 응답을 낸다.
_Avoid_: client, handler, runner

**Runtime interpreter**:
매니페스트를 데이터로 두고 실행 중에 명령을 만들어 내는 우리 아키텍처. 반대 개념은 Code generation (매니페스트를 .go 소스로 물질화).
_Avoid_: dynamic mode

## Relationships

- 하나의 **Manifest** 는 하나 이상의 **Command** 를 가진다
- 하나의 **Command** 는 0개 이상의 **Param** 을 가진다
- 하나의 **Command** 는 0개 또는 1개의 **Body** 를 가진다
- 하나의 **Executor** 가 모든 **Command** 를 실행한다 (Command 마다 코드가 따로 생기지 않음)

## Example dialogue

> **Dev:** "OpenAPI 파일을 주면 그게 곧 **Manifest** 인가요?"
> **Domain expert:** "아니요 — OpenAPI 는 **Spec** 이라 부르고, 우리 **Manifest** 는 별도의 커스텀 형식입니다. 나중에 Spec → Manifest 임포터를 얹을 수 있어요."
>
> **Dev:** "`POST /pet` 이 보내는 JSON 은 **Param** 하나로 적으면 되나요?"
> **Domain expert:** "아니요 — 그건 **Body** 입니다. **Param** 은 이름이 있어서 여러 개를 나열할 수 있지만, **Body** 는 이름이 없고 한 **Command** 에 하나뿐입니다."

## Flagged ambiguities

- "spec" 이 OpenAPI 와 우리 매니페스트 둘 다를 가리켜 혼동됨 — 해소: **Spec** = OpenAPI 등 외부 표준 명세, **Manifest** = 우리 커스텀 형식. 구분해서 쓴다.
- "flag" 와 **Param** 혼동 — 해소: **Param** 은 도메인 개념(요청 입력), flag 는 그 Param 이 CLI 표면에 드러난 표현. 같지 않다.
- "body" 가 **Param** 의 `in` 값인지 별개 개념인지 — 해소: 별개다. **Param** 은 이름 붙은 입력, **Body** 는 이름 없는 단일 페이로드 ([ADR-0010](docs/adr/0010-request-body-as-command-field.md)).
