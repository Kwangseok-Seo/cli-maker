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

## 반대 방향 — `Marshal`

```go
out, err := yaml.Marshal(m)   // Manifest → []byte
```

M11 의 `import` 가 이 방향이다: Spec 을 언마샬해서 struct 로 받고 → `Manifest` 로 옮기고 → 마샬해서 YAML 파일로 낸다. **같은 struct 태그가 양쪽을 다 정한다** — 읽을 때의 키 이름이 곧 쓸 때의 키 이름이다([[struct-tags]]).

### `omitempty` — 한 struct, 두 요구

```go
Params []Param `yaml:"params,omitempty"`   // 비면 키 자체가 안 나온다
```

같은 struct 를 마샬하는 자리가 둘인데 요구가 반대일 수 있다.

| | 원하는 것 | 이유 |
|---|---|---|
| `parse` (디버그 도구) | 빈 필드도 **보이길** | 오타 난 키는 조용히 버려지므로, `required: false` 라는 빈 값이 *"네가 적은 게 안 들어왔다"* 의 유일한 단서다 |
| `import` (초안 생성) | 빈 필드는 **숨기길** | 유저가 손으로 이어 쓸 파일에 `params: []`·`body: null` 이 19개 명령마다 붙는다 |

cli-maker 는 **필드별로 갈랐다.** 노이즈를 내는 것(`Params`·`Body`)과 오타를 드러내는 것(`Auth`·`Param.Required`)이 서로 다른 필드였기 때문이다. 앞의 둘에만 붙여 산출물이 180줄 → 161줄이 되면서, `parse` 는 `autth`/`requiredd` 오타를 그대로 드러낸다.

`omitempty` 는 zero value 를 기준으로 하므로 `false` 와 `""` 도 사라진다는 점을 같이 봐야 한다 — "명시적으로 false 를 적었다"를 남길 수 없다([[absent-vs-empty]]).

## 겪은 함정

- **`fmt` 로 struct 를 찍었더니 포인터가 주소로 나왔다.** `%+v` 는 중첩 포인터를 펼치지 않아 `Body:0xc000...` 이 찍혔다(M10). 다시 YAML 로 마샬하면 펼쳐지고, **유저가 적은 것과 같은 언어**라 입력과 나란히 놓고 볼 수 있다 — `parse` 가 `%+v` 대신 `yaml.Marshal` 을 쓰는 이유다.
- **에러 검사 위치 뒤엉킴**: ReadFile 에러를 확인하기 전에 Unmarshal 을 먼저 돌렸다. 파일이 없어도 마침 `Unmarshal(nil, …)` 이 무해하고, `if err := …` 의 `:=` 가 **새 `err`(shadowing)** 를 만들어 바깥 err 을 안 건드린 탓에 "운으로" 동작했다. 관용구는 **에러를 만든 호출 바로 뒤에서 즉시 검사(fail fast)**.

## 관련

[[pointers]] · [[struct-tags]] · [[file-io]] · [[partial-decoding]] · [[absent-vs-empty]]
