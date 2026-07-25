# unused-results

**반환값을 버리면 조용히 아무 일도 안 일어난다.** M4 에서 두 번 밟은 함정이라 따로 세운다.

## 왜 컴파일러가 안 막나

Go 는 함수 *호출* 을 문장으로 허용한다 — `fmt.Println(...)` 처럼 부수효과가 목적일 수 있기 때문. 그래서 반환값을 안 받아도 문법상 정상이다.

```go
strings.ReplaceAll(path, "{petId}", "10")   // 컴파일 통과. 아무 일도 안 일어남
fmt.Errorf("HTTP %s", resp.Status)          // 컴파일 통과. 에러를 만들고 즉시 버림
```

값이 아니라 *식* 이면 얘기가 다르다 — `x + y` 만 쓰면 "evaluated but not used" 로 막힌다. 호출만 예외다.

## 실측 — 문자열은 불변

```go
s := "/pet/{petId}"
strings.ReplaceAll(s, "{petId}", "10")
fmt.Println(s)   // /pet/{petId}   ← 그대로

s = strings.ReplaceAll(s, "{petId}", "10")
fmt.Println(s)   // /pet/10
```

`ReplaceAll` 은 `s` 를 고치지 않고 **새 문자열을 만들어 반환**한다. 받지 않으면 버려진다.

## 세 겹의 그물

| 층 | 잡는 것 | 놓치는 것 |
|---|---|---|
| 컴파일러 | 타입 · 미사용 **변수** · `missing return` | 반환값 버리기 |
| `go vet` | `fmt.Errorf`·`errors.New`·`fmt.Sprintf` 등 **목록에 든** 함수의 버려진 결과 | 목록 밖의 같은 실수 |
| 테스트 (M7) | 의미가 틀린 것 | — |

실제로 M4 에서 `strings.ReplaceAll` 은 `go vet` 도 조용히 통과했고, 같은 실수를 `fmt.Errorf` 로 했을 때는 잡혔다:

```
internal\executor\execute.go:43:3: result of fmt.Errorf call not used
```

`vet` 의 목록은 오탐을 피하려고 보수적으로 골라져 있다 — "이 함수의 결과를 버리는 건 **항상** 무의미하다"가 확실한 것만 들어간다. `ReplaceAll` 은 그 확신이 없어 빠져 있다.

**그래서 `go build` 만이 아니라 `go vet` 도 습관적으로 돌린다** → [[go-toolchain]].

## 대비 — 에러는 오히려 컴파일러가 강제한다

```go
req, err := http.NewRequestWithContext(...)
resp, err := client.Do(req)   // ✗ declared and not used: err
```

받아 놓고 안 *읽은* 에러는 미사용 변수 규칙에 걸려 **컴파일이 안 된다**. 예외가 없는 언어에서 에러 무시를 막는 안전망 → [[error-handling]].

## 관련

[[error-handling]] · [[go-toolchain]] · [[url-encoding]] · [[variables]] · [[errors-compile-vs-runtime]]
