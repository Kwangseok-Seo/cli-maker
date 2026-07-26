# type-aliases

`type X = Y` 는 **같은 타입에 이름을 하나 더** 붙인다. `=` 가 없는 `type X Y` 는 **새 타입**을 만든다. 글자 하나 차이인데 성질이 다르다.

## 실측

```go
type Base struct{ Name string }
func (b Base) Hello() string { return "hello " + b.Name }

type Alias = Base    // 별칭
type Defined Base    // 정의

func take(b Base) string { return b.Hello() }
```

```
별칭  take(a)   → hello alias        통과
별칭  a.Hello() → hello alias        메서드가 따라온다
별칭  %T        → main.Base          이름이 하나 더 붙었을 뿐, 타입은 Base

정의  %T        → main.Defined       다른 타입이다
정의  take(d)   → cannot use d (variable of struct type Defined) as Base value
정의  d.Hello() → d.Hello undefined (type Defined has no field or method Hello)
```

**정의는 메서드를 물려받지 않는다.** 바탕(underlying type)만 같을 뿐이다.

## 어디에 쓰나 — internal 을 감싼 공개 표면

우리 도메인 타입은 `internal/manifest` 에 있어 [[internal-packages]] 규칙 때문에 남의 모듈에서 못 쓴다. 그런데 생성된 CLI 는 매니페스트를 Go 값으로 적어야 한다.

```go
// clirun (공개 패키지)
type (
	Manifest = manifest.Manifest
	Command  = manifest.Command
	Param    = manifest.Param
	Auth     = manifest.Auth
)
```

생성된 코드는 `internal/manifest` 라는 **이름을 부를 수 없지만**, `clirun.Manifest` 라는 다른 이름으로 같은 타입에 닿는다:

```go
m := &clirun.Manifest{Name: "gh", Commands: []clirun.Command{...}}
```

그리고 이 값이 `executor.Execute(ctx, m, ...)` 에 **변환 없이 그대로** 들어간다. 정의(`type Manifest manifest.Manifest`)로 뒀다면 `cannot use ... as ... value` 로 막힌다.

## 남는 흔적

별칭은 타입의 정체를 바꾸지 않으므로, 리플렉션이나 에러 메시지엔 원래 이름이 드러난다:

```
정적 타입: mf.Command        ← 별칭으로 만들었어도 internal 패키지 이름이 보인다
```

접근을 막는 것과 이름을 감추는 것은 다르다. 규칙이 막는 건 **import 경로**이지 타입의 정체가 아니다.

## 겪은 함정

- **`type X Y` 를 무심코 쓰면 컴파일이 한참 뒤에 깨진다.** 선언 자체는 통과하고, 그 값을 다른 함수에 넘기는 자리에서야 터진다. 별칭이 의도라면 `=` 를 빼먹지 않는지가 유일한 관문이다.
- **모든 것을 별칭으로 내보내고 싶어진다.** 공개 표면은 곧 계약이므로([[internal-packages]]), **생성물이 실제로 이름을 불러야 하는 것만** 내보냈다 — 타입 4개 + 상수 2개 + 함수 1개.

## 관련

[[internal-packages]] · [[structs]] · [[methods]] · [[interfaces]] · [[text-template]]
