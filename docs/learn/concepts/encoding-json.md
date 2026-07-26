# encoding-json

한 패키지 안에 추상화 수준이 다른 **세 벌**의 API 가 있다. 어느 층을 쓰느냐가 결과를 바꾼다.

## 세 층

| 층 | API | 하는 일 | 스키마를 알아야 하나 |
|---|---|---|---|
| 바이트 → 바이트 | `json.Indent`, `json.Compact`, `json.Valid` | 토큰만 훑고 **공백만** 조절 | ❌ |
| 바이트 ↔ Go 값 | `json.Marshal`, `json.Unmarshal` | 구조체·map 으로 변환 | ⭕ |
| 스트림 ↔ Go 값 | `json.NewDecoder(r)`, `json.NewEncoder(w)` | `io.Reader/Writer` 위에서 값 단위 처리 | ⭕ |

[[serialization]] 의 `yaml.Unmarshal` 은 가운데 층이다 — 매니페스트는 **우리가 스키마를 정했으니** 그게 맞았다. API 응답은 스키마를 서버가 정하므로 다르다.

## cli-maker 는 첫 번째 층만 쓴다

```go
func Indent(dst *bytes.Buffer, src []byte, prefix, indent string) error
func Compact(dst *bytes.Buffer, src []byte) error
```

- `src []byte` — **`io.Reader` 가 아니다.** 들여쓰려면 끝까지 봐야 하므로 `io.ReadAll` 로 전부 읽어야 한다. [[io-reader-writer]] 의 32KB 스트리밍이 여기서 끝난다.
- `dst *bytes.Buffer` — `io.Writer` 가 아니라 **구체 타입**이다. 그래서 stdout 에 바로 못 흘리고 버퍼를 거친다. 대신 아래의 "실패해도 안 더럽혀짐"이 공짜로 따라온다.
- 시그니처가 같아서 `json.Compact` 는 `func(*bytes.Buffer, []byte) error` 자리에 **함수값 그대로** 넘어간다([[functions-as-values]]). `Indent` 는 인자가 넷이라 클로저로 감싼다([[closures]]).

## 측정 ① — 가운데 층을 쓰면 키 순서가 사라진다

같은 GitHub 응답을 두 층으로 pretty 하게 만든 결과:

```
층 1: json.Indent                   층 2: Unmarshal(map) → MarshalIndent
{                                   {
  "id": 12574344,                     "allow_forking": true,
  "node_id": "MDEwOlJlcG9z…",         "archive_url": "https://api.github…",
  "name": "cobra",                    "archived": false,
```

두 단계로 소실된다: ① `map[string]any` 로 받는 순간 **Go map 에는 순서가 없고** ② `json.Marshal` 은 재현성을 위해 **키를 정렬한다**([[maps]]).

## 측정 ② — 실패하면 dst 가 안 더럽혀진다

```
입력: {"ok":1} <html>사실은 프록시 에러 페이지</html>
→ invalid character '<' after top-level value
→ dst 에 쓰인 바이트: 0     (앞의 유효한 {"ok":1} 조차 남지 않음)
```

앞부분이 유효해도 되돌린다. 그래서 "JSON 이 아니면 원본을 그대로" 폴백이 **이중 출력 없이** 안전하다.

## 겪은 함정

- 응답이 JSON 이 아닐 수 있다. `GET https://api.github.com/zen` 은 평문 한 줄(`Encourage flow.`)을 주는 **정상** 엔드포인트다. 포맷 실패를 에러로 취급하면 이런 API 가 막힌다.
- GitHub 은 이미 compact 로 응답한다(5125 B → `json.Compact` 후에도 5125 B). `--compact` 가 일하는 건 pretty 로 주는 API 쪽이다(jsonplaceholder 84 B → 66 B). **"이 옵션이 실제로 뭘 바꾸나"는 표적을 두 개 이상 재 봐야 보인다.**

## 관련

[[serialization]] · [[io-reader-writer]] · [[maps]] · [[functions-as-values]] · [[closures]] · [[interfaces]]
