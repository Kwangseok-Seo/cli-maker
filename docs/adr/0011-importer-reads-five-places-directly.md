# 임포터는 Spec 의 다섯 자리만 직접 읽는다

OpenAPI Spec 을 라이브러리 없이 직접 **부분 디코드**한다 — `servers`·`paths`·`operationId`·`parameters`·`securitySchemes` 다섯 자리에만 struct 칸을 파고 나머지는 언마샬이 조용히 버리게 둔다. 매니페스트로 옮길 수 있는 것이 그 다섯뿐이고, petstore 의 `$ref` 43개가 전부 `requestBody`/`responses` 안 — 우리가 보는 자리엔 0개 — 이라 `$ref` 해석기가 할 일이 없기 때문이며, 대신 **우리가 오타 낸 yaml 태그와 원래 안 판 칸이 구별되지 않는** 대가를 감수한다.

그 대가가 실재한다는 것은 실측으로 확인했다. `parameters` 를 `parameter` 로 오타 낸 struct 는 에러 없이 빈 슬라이스를 냈다. 그래서 이 패키지의 테스트는 "언마샬이 성공했다"가 아니라 **실물 petstore 를 넣어 무엇이 담겼는지**를 본다 (`loadPetstore`) — 손으로 줄인 fixture 는 우리가 이미 아는 것만 담아서, 정작 우리가 안 판 칸을 못 잡는다.

Swagger 2.0 은 **거부하고 이름을 대어 알린다**. 2.0 은 `paths`·`operationId`·`in: path` 를 3.0 과 같은 이름으로 쓰므로 버전 게이트가 없으면 절반이 그냥 옮겨지는데, `servers` 자리가 `host`+`basePath`+`schemes` 로 흩어져 baseUrl 이 비고 non-body param 의 타입이 `schema.type` 이 아니라 `type` 에 직접 있어 조용히 유실된다 — 실측상 **경고 한 줄 없이** 반쪽 매니페스트가 나왔다. 판정과 에러 문구는 `supportedMajor` 상수 하나를 함께 쓴다. 둘이 갈리면 지원 범위를 넓힐 때 판정만 바뀌고 문구는 옛 범위를 계속 광고한다.

거부한 대안: **OpenAPI 라이브러리**(kin-openapi 등)는 `$ref` 해석·스키마 검증·3.1 대응을 주지만, 우리가 읽는 다섯 자리엔 `$ref` 가 없고 스키마에서 쓰는 것은 `type` 한 필드뿐이라 얻는 것이 없다. `components.parameters` 재사용을 쓰는 spec 을 만나면 그때 재검토한다 — 그때는 `$ref` 가 우리가 보는 자리에 처음 들어온다.
