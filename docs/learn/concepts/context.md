# context

"언제까지 / 취소됐는지"를 호출 사슬 **아래로** 전파하는 값. 관례상 함수의 첫 인자이고 이름은 `ctx`.

## 핵심

```go
ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
defer cancel()

req, err := http.NewRequestWithContext(ctx, c.Method, url, nil)
```

- `WithTimeout(부모, 기간)` → `(자식 ctx, cancel 함수)`. 기간이 지나면 자식 ctx 가 "끝났다"고 신호한다.
- `cancel` 은 **반드시** 부른다. 안 부르면 만료까지 타이머가 남는다 → `defer cancel()` ([[defer]]).
- `cmd.Context()` 는 [[cobra]] 가 주는 부모 ctx (기본은 `context.Background()`). 나중에 Ctrl-C 신호를 여기 연결하면 진행 중인 요청이 즉시 끊긴다.

## 왜 필요한가

서버가 죽어 있으면 `Do` 는 즉시 실패한다. 문제는 **서버가 살아 있으면서 답을 안 줄 때** — 기본 설정의 Go 는 영원히 기다리고 CLI 가 그대로 멈춘다. ctx 는 그 상한을 건다.

## 넣었다고 도는 것은 아니다 — 실측

코드에 `ctx` 가 있는 것과 그게 실제로 전송을 끊는 것은 다른 얘기라 확인했다. 3초 뒤 응답하는 `httptest` 서버에 500ms 짜리 ctx 로 요청:

```
경과 = 500ms                          ← 3초를 다 기다리지 않았다
err  = ... context deadline exceeded
out  = ""
```

`NewRequestWithContext` 로 실은 ctx 가 전송 계층까지 닿아 있음이 확인됐다. `NewRequest`(ctx 없는 버전)를 쓰면 같은 코드가 3초를 다 기다린다.

## 겪은 함정

- 타임아웃 값을 코드에 박았다(30초). 되돌리기 쉬운 국소 결정이라 그대로 뒀지만, 느린 API 를 쓰는 유저는 고칠 방법이 없다 — 설정으로 빼는 것이 M5 의 몫.

## 관련

[[net-http]] · [[defer]] · [[cobra]] · [[error-handling]]
