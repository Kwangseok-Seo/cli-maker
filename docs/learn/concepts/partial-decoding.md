# partial-decoding (부분 디코드)

**남이 정한 형식에서 내가 쓰는 자리만 struct 로 파는 것.** M2 의 언마샬과 기계는 같은데 처지가 반대다 — 그땐 우리가 형식의 주인이었고, 여기서는 17KB 짜리 OpenAPI Spec 에서 다섯 자리만 꺼내고 나머지 대부분을 버린다([ADR-0011](../../adr/0011-importer-reads-five-places-directly.md)).

## 핵심 — 칸이 없는 키는 조용히 버려진다

```go
type Spec struct {
	OpenAPI string              `yaml:"openapi"`
	Servers []Server            `yaml:"servers"`
	Paths   map[string]PathItem `yaml:"paths"`
}
```

`info`·`tags`·`components`·`externalDocs`·`responses` … 전부 사라진다. **에러가 아니다.** 이게 기능이다 — 남의 형식이 자라도 우리 코드는 안 깨진다.

`.json` 을 `yaml.Unmarshal` 로 읽어도 된다. YAML 1.2 가 JSON 의 상위집합이라, `encoding/json` 을 따로 들이지 않아도 `.json` / `.yaml` 스펙을 한 함수로 받는다([[serialization]]).

## 양날 — 내 오타도 똑같이 조용히 버려진다

```go
type Operation struct {
	Params []Parameter `yaml:"parameter"`   // 스펙은 "parameters" 다
}
```

```
[2] getPetById: operationId="getPetById"  params=[] (len=0)
```

에러도 경고도 없다. 언마샬은 **"버려야 할 키"와 "내가 오타 낸 칸"을 구별하지 못한다.** 필드가 zero value 로 남는 것 말고는 알 방법이 없고, 그래서 "언마샬이 통과했다"는 아무것도 보장하지 않는다. [[struct-tags]] 의 camelCase 함정과 같은 뿌리인데, 형식의 주인이 남이라 훨씬 늦게 드러난다.

## "안 읽는다"를 타입으로 적기

값은 안 보고 **키만** 볼 때가 있다. 그러면 값 타입을 빈 struct 로 둔다.

```go
Content map[string]struct{} `yaml:"content"`
```

`application/json` 같은 미디어 타입 이름만 필요하고 그 밑의 스키마는 안 쓴다는 뜻이 타입에 적혀 있다. 실제로 통과한다 — 값 쪽 키가 전부 "칸 없음"이 되어 버려질 뿐이다.

## 겪은 함정

- **다른 형식이 절반만 통과했다.** Swagger 2.0 문서를 넣었더니 **에러 없이** `paths`·`operationId`·`in: path` 가 그대로 넘어왔다 — 2.0 이 그 자리를 3.0 과 같은 이름으로 쓰기 때문이다. 대신 `servers` 가 없어 baseUrl 이 비고, non-body param 의 타입이 `schema.type` 이 아니라 `type` 에 직접 있어 `type: ""` 로 유실됐다. **경고 0건.** 부분 디코드는 "이 문서가 내가 아는 형식인가"를 절대 물어봐 주지 않으므로, **버전 게이트를 손으로 세워야** 한다.
- **줄인 fixture 는 이 실패를 못 잡는다.** 손으로 만든 최소 spec 은 우리가 이미 아는 키만 담기 때문에, 정작 "우리가 안 판 칸"이 거기 없다. 그래서 테스트 표적을 실물 petstore(17KB)로 잡았고, 판정도 "에러가 안 났다"가 아니라 **무엇이 담겼는지**로 했다 → [[testing-the-filesystem]].

## 관련

[[serialization]] · [[struct-tags]] · [[maps]] · [[absent-vs-empty]] · [[testing-the-filesystem]]
