# struct-tags

struct 필드 옆에 붙는 백틱 꼬리표. 언마샬 라이브러리에게 "이 필드는 저 키에서 채워라"를 알려주는 메타데이터.

## 핵심

```go
type Manifest struct {
	BaseURL string `yaml:"baseUrl"`   // YAML 의 baseUrl 키 → Go 의 BaseURL 필드
}
```

- 태그는 필드에 붙는 **메타데이터** — 라이브러리가 리플렉션으로 읽어 매핑에 쓴다.
- 전제: 라이브러리가 값을 채우려면 필드가 **exported(대문자 시작)** 여야 한다 → [[packages-and-main]] 의 대문자 규칙.
- 기본 매칭: yaml.v3 는 필드명을 **소문자화**해 키와 비교 → `Name` ↔ `name` 은 태그 없이도 맞는다.

## 겪은 함정

- **camelCase 키는 태그가 필수**: `baseUrl` 은 필드 `BaseURL` 을 소문자화한 `baseurl` 과 **안 맞는다**. 태그 `` `yaml:"baseUrl"` `` 이 있어야 이어진다. "태그가 왜 존재하나"의 실전 증거 — 필드명은 Go 관례(`BaseURL`)대로 두고, 태그가 YAML 표면(`baseUrl`)과의 **다리**가 된다.

## 관련

[[structs]] · [[serialization]] · [[packages-and-main]]
