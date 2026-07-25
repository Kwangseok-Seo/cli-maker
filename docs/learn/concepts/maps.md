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

## 겪은 함정

- **누적자를 루프 안에서 만들었다.** M4 의 `BuildURL` 에서 `q := url.Values{}` 를 루프 안에 둬서 쿼리 누적자가 매 반복마다 새로 태어났다. 같은 실수를 `seen` 에서 반복하지 않으려면 "누적하는 것은 누적 구간 밖에서 태어난다"로 기억한다 — 위의 이중 루프는 그 규칙을 어긴 게 아니라, 누적 구간이 command 하나인 것이다.
- map 은 **순회 순서가 무작위**다. 순서가 필요하면 리스트를 따로 든다 — 매니페스트 컬렉션을 map 이 아니라 리스트로 둔 이유가 그것이다([ADR-0002](../../adr/0002-manifest-collections-as-ordered-lists.md)).

## 관련

[[config-precedence]] · [[environment-variables]] · [[http-headers]] · [[variables]] · [[slices-and-args]]
