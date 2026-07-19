# 매니페스트 컬렉션은 name 필드를 가진 순서 있는 리스트

매니페스트의 컬렉션(**commands**, **params**)을 이름을 키로 삼는 map 이 아니라 각 항목이 `name` 필드를 갖는 **순서 있는 YAML 리스트**로 표현한다 — `--help` 의 명령 나열 순서를 유저가 적은 대로 보존하고 각 **Command**/**Param** 이 자기 이름을 품는 self-contained 도메인 객체가 되게 하기 위해, YAML 이 다소 장황해지고(`- name:` 접두) 이름 중복을 파싱 단계가 자동으로 막아주지 못하는 비용을 감수한다.

OpenAPI 등 흔한 스키마는 `paths` 를 URL 키로 삼는 map 으로 두므로 이 선택은 맥락 없이는 의외로 보인다 — 그래서 기록한다. 다만 OpenAPI 조차 명령의 핸들에 해당하는 `operationId` 는 키가 아닌 **필드**로 두며, 우리 `name` 은 그 계보다. Go 표현은 `Manifest.Commands []Command` / `Command.Params []Param` 이고, 이름으로 조회가 필요해지는 시점(path 치환·flag 조회)엔 리스트에서 `map[string]T` 를 파생해 쓴다 — 파싱 단계엔 조회가 없다.
