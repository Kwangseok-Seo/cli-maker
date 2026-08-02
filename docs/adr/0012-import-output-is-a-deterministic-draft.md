# 임포트 산출물은 저자 순서를 버리고, 못 옮기는 것은 required 여부로 가른다

Spec 의 `paths` 는 map 이라 순서가 없으므로 경로를 **사전순**으로 정렬해 명령을 낸다 — 같은 Spec 을 두 번 import 해도 같은 파일이 나와야 diff 가 매번 통째로 뜨지 않기 때문이며, 대신 spec 저자가 `/pet`→`/store`→`/user` 로 묶어 둔 의도를 잃는 것을 감수한다.

ADR-0002 는 컬렉션을 map 이 아니라 순서 있는 리스트로 둔 근거 하나로 *"`--help` 순서를 유저가 적은 대로 보존"* 을 들었는데, **그 근거는 임포터까지 덮지 않는다.** 거기서 말한 "유저가 적은 순서"는 유저가 **손으로 적은** 매니페스트의 것이고, 외부 Spec 의 저자 순서까지 지키려면 struct 언마샬 대신 `yaml.Node` 로 문서를 직접 걸어야 해서 비용이 이익을 넘는다. ADR-0002 자체는 그대로 유효하다 — supersede 가 아니라 **적용 범위를 확정한 것**이다. 무순서가 들어오는 자리는 셋(`paths`, `requestBody.content`, `securitySchemes`)이고 전부 같은 방식(사전순)으로 없앤다.

못 옮기는 param 은 **required 여부로 갈린다** — optional 이면 그 param 만 빼고 명령은 살리며, required 면 그 operation 을 통째로 뺀다. 필수 입력을 잃은 명령은 반드시 잘못된 요청을 보내지만, 선택 입력을 잃은 명령은 나머지가 정확하기 때문이다. 뺀 것은 param 이든 operation 이든 전부 stderr 로 한 줄씩 알린다. 산출물은 stdout 으로만 나가므로 `import … > apis/x.yaml` 이 성립한다. 인증은 통째로 못 옮기므로(`securitySchemes` 의 oauth2·apiKey 어느 쪽도 `{type: bearer, env}` 에 대응되지 않는다) `auth` 를 빈 채로 두고 경고한다 — 반쪽으로 채우면 조용히 틀린다.

`requestBody.content` 후보가 여럿이면 `application/json` 을 먼저 고르고, 없으면 사전순 첫 번째를 고르되 그때만 경고한다. 명령 이름은 `operationId` 를 그대로 쓴다 — ADR-0002 가 이미 우리 `name` 을 operationId 의 계보로 적어 뒀고, kebab 변환은 변환 규칙·역변환·충돌 dedupe 를 새로 만들어야 해서 거부했다.

산출물이 **초안**이라는 것이 이 결정들의 전제다. `import` 는 기존 파일을 덮어쓰지 않고 거절한다 — 유저가 손으로 이어 쓴 편집분을 두 번째 import 가 조용히 지우면 되돌릴 방법이 없다.
