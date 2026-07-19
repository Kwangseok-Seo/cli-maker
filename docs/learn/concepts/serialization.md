# serialization (마샬 · 언마샬)

텍스트·바이트 ↔ 메모리 속 값의 상호 변환. **언마샬** = 텍스트를 struct 로 붓는 것(M2 의 심장).

## 핵심

```go
var m Manifest
if err := yaml.Unmarshal(data, &m); err != nil { ... }  // data 를 파싱해 m 에 채운다
```

- **마샬(marshal)**: 메모리 값 → 텍스트/바이트. **언마샬(unmarshal)**: 그 반대. M2 는 YAML → `Manifest` 언마샬.
- **decode-into 패턴**: 결과를 반환받는 대신 "채울 곳(`&m`)"을 넘긴다. `json.Unmarshal` 도 같은 꼴.
- **왜 `&`(주소)인가**: Go 는 인자를 값(복사본)으로 넘긴다. `m` 을 그냥 주면 복사본만 채워지고 원본은 빈 채. `&m`(주소)을 주면 Unmarshal 이 **그 자리에 써넣는다** → [[pointers]].
- YAML 키 ↔ Go 필드 매핑은 [[struct-tags]] 가 담당.

## 겪은 함정

- **에러 검사 위치 뒤엉킴**: ReadFile 에러를 확인하기 전에 Unmarshal 을 먼저 돌렸다. 파일이 없어도 마침 `Unmarshal(nil, …)` 이 무해하고, `if err := …` 의 `:=` 가 **새 `err`(shadowing)** 를 만들어 바깥 err 을 안 건드린 탓에 "운으로" 동작했다. 관용구는 **에러를 만든 호출 바로 뒤에서 즉시 검사(fail fast)**.

## 관련

[[pointers]] · [[struct-tags]] · [[file-io]]
