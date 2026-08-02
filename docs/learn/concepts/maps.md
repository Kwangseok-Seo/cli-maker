# maps

키로 값을 찾는 표. `map[키타입]값타입`. cli-maker 에서는 flag 값 모음(`map[string]string`)과 **집합(set)** 두 용도로 쓴다.

## 없는 키를 조회해도 패닉하지 않는다

```go
seen := map[string]bool{}
if seen["ping"] { }   // false — 에러도 패닉도 아니다
```

**없는 키는 값 타입의 zero value 를 준다.** `bool` 이면 `false`, `string` 이면 `""`, `int` 면 `0`.

이 성질이 집합 관용구를 만든다 — "본 적 없음 = `false`" 가 정확히 원하는 뜻이라, 존재 확인용 분기가 따로 필요 없다.

```go
seen := map[string]bool{}
for _, name := range names {
	if seen[name] { /* 중복! */ }
	seen[name] = true
}
```

**이번엔 zero value 가 우리 편이다.** M4·M5 에서는 같은 성질이 "안 줬다"를 지워 버려서 세 번 발목을 잡았다(`""`, `0`, pflag 기본값 — [[config-precedence]]). 성질은 하나고, 그 자리에서 zero value 가 옳은 뜻인지가 매번 다르다.

## 존재 자체를 알아야 하면 comma-ok

zero value 와 "없음"을 구별해야 하면 두 번째 반환값을 받는다.

```go
v, ok := m["key"]   // ok = 키가 있었는가
```

`os.LookupEnv` 와 같은 모양이다([[environment-variables]]).

## nil map — 읽기는 되고 쓰기는 패닉

```go
var seen map[string]bool   // nil
_ = seen["a"]              // OK — false
seen["a"] = true           // panic: assignment to entry in nil map
```

그래서 `map[string]bool{}` 또는 `make(map[string]bool)` 로 **초기화까지** 해야 한다. `http.Header` 를 다룰 때 이 패닉을 안 밟은 건 `NewRequestWithContext` 가 대신 초기화해 주기 때문이다([[http-headers]]).

## 두 가지 집합 — 쌓아 가는 것과 고정된 것

```go
seen := map[string]bool{}                       // 돌면서 채운다 (중복 탐지)
var allowedMethods = map[string]bool{"GET": true, ...}   // 미리 고정 (화이트리스트)
```

화이트리스트는 `!allowedMethods[c.Method]` 로 쓴다 — 없는 키는 `false` 니 `!` 를 붙이면 "허용 목록에 없다"가 된다. 고정 집합은 **함수 밖**에 둔다. 안에 두면 함수를 부를 때마다 map 을 새로 만든다.

## 선언 위치가 규칙의 범위를 표현한다

```go
seen := map[string]bool{}          // 함수당 하나 → command 이름은 매니페스트 전체에서 유일
for i, c := range m.Commands {
	seenParam := map[string]bool{}  // command 마다 새로 → param 이름은 command 안에서만 유일
	for j, p := range c.Params { ... }
}
```

`gh repo --owner` 와 `gh issues --owner` 가 공존해야 하므로 `seenParam` 은 안쪽에서 태어나야 한다. **변수의 수명이 곧 규칙의 적용 범위**다.

## 순회 순서는 의도적으로 무작위다

`range` 는 순서를 보장하지 않을 뿐 아니라 **매번 섞는다.** 같은 프로세스에서 같은 map 을 세 번 돈 결과다.

```
range #1 앞 4개: [/user/logout /user/{username} /pet /pet/findByTags]
range #2 앞 4개: [/pet /pet/findByTags /pet/{petId} /pet/{petId}/uploadImage]
range #3 앞 4개: [/pet/{petId}/uploadImage /store/order /store/order/{orderId} /user/createWithList]
```

우연히 정렬돼서 그걸 믿는 코드가 생기는 것을 막으려는 설계다. 순서가 필요하면 **키를 모아 정렬**한다.

```go
keys := make([]string, 0, len(m))   // 옛 관용구
for k := range m {
	keys = append(keys, k)
}
sort.Strings(keys)

keys := slices.Sorted(maps.Keys(m))  // Go 1.23+ — 위 4줄이 한 줄
```

`maps.Keys` 가 돌려주는 것은 슬라이스가 아니라 **iterator**(`iter.Seq[string]`)다 — 값을 다 만들어 쌓아 두는 대신 "다음 것 줘"를 반복하는 함수다. `slices.Sorted` 가 그걸 받아 모으고 정렬한다.

**더 나은 수는 애초에 map 을 안 받는 것이다.** OpenAPI 의 `paths./pet` 아래도 map(키가 `get`/`post`/…)이지만, 키의 가짓수가 HTTP 메서드로 정해져 있어 struct 로 받을 수 있다. 그러면 순회 순서가 **필드 선언 순서**로 고정돼 결정론이 공짜로 따라오고, 정렬해야 할 자리가 하나 줄어든다.

```go
type PathItem struct {
	Get  *Operation `yaml:"get"`
	Post *Operation `yaml:"post"`
	…                              // 없는 메서드는 nil
}
```

## 겪은 함정

- **누적자를 루프 안에서 만들었다.** M4 의 `BuildURL` 에서 `q := url.Values{}` 를 루프 안에 둬서 쿼리 누적자가 매 반복마다 새로 태어났다. 같은 실수를 `seen` 에서 반복하지 않으려면 "누적하는 것은 누적 구간 밖에서 태어난다"로 기억한다 — 위의 이중 루프는 그 규칙을 어긴 게 아니라, 누적 구간이 command 하나인 것이다.
- map 은 **순회 순서가 무작위**다. 순서가 필요하면 리스트를 따로 든다 — 매니페스트 컬렉션을 map 이 아니라 리스트로 둔 이유가 그것이다([ADR-0002](../../adr/0002-manifest-collections-as-ordered-lists.md)).
- **그 근거가 어디까지 덮는지는 따로 정해야 했다.** M11 의 임포터는 남의 Spec(`paths` 가 map)에서 명령을 만드는데, 여기서 "유저가 적은 순서"를 지키려면 struct 언마샬 대신 `yaml.Node` 로 문서를 걸어야 한다. 비용이 이익을 넘어 **사전순 정렬로 갔다** — ADR-0002 를 뒤집은 게 아니라 적용 범위를 확정한 것이다([ADR-0012](../../adr/0012-import-output-is-a-deterministic-draft.md)).

## 관련

[[config-precedence]] · [[environment-variables]] · [[http-headers]] · [[variables]] · [[slices-and-args]] · [[partial-decoding]]
