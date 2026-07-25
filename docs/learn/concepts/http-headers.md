# http-headers

`http.Header` 는 `map[string][]string` 이다 — 키 하나에 값이 **여러 개** 올 수 있기 때문(`Set-Cookie` 를 생각하면 된다).

## Set vs Add

```go
req.Header.Set("Authorization", "Bearer "+token)  // 있으면 교체
req.Header.Add("Cookie", "x=1")                   // 있어도 뒤에 추가
```

실제로 담기는 모습:

```
map[Authorization:[Bearer a] Cookie:[x=1 y=2]]
```

## 키는 정규화된다

소문자로 넣어도 저장은 `Authorization` 이다. `Set`/`Get`/`Add` 가 `Cookie`, `Content-Type` 같은 표준 표기로 바꿔 준다. HTTP 헤더 이름은 대소문자를 구분하지 않으니 조회할 때 표기를 걱정하지 않아도 된다.

## 붙이는 자리 — 생성 뒤, 전송 전

```go
req, err := http.NewRequestWithContext(ctx, method, url, nil)
// ← 여기
resp, err := http.DefaultClient.Do(req)
```

`Do` 가 요청을 직렬화해 소켓에 쓰는 순간이므로, 그 전에 map 에 들어가 있어야 실린다.

## nil map 함정을 왜 안 밟는가

Go 에서 **nil map 에 쓰면 패닉**이다:

```
nil map 에 Set → assignment to entry in nil map
```

그런데 `NewRequestWithContext` 가 돌려준 `req` 는 `Header` 가 이미 초기화돼 있다(`req.Header == nil` → `false`). 그래서 받자마자 `Set` 을 불러도 안전하다. 직접 `&http.Request{}` 를 조립하면 이 보장이 없다.

## 헤더가 실렸는지 확인하는 법 — 실측

우리 코드를 들여다보는 대신 **서버에게 물으면** 된다. GitHub `/user` 는 두 상황에 다른 문장을 준다:

| 우리가 보낸 것 | 응답 |
|---|---|
| 헤더 없음 | 401 `Requires authentication` |
| 엉터리 토큰 헤더 | 401 `Bad credentials` |

`Bad credentials` 가 나왔다는 것 자체가 헤더가 도착했다는 증거다. 익명 허용 엔드포인트라면 더 선명하다 — 헤더 없음은 200, 엉터리 토큰은 401 로 갈린다([[environment-variables]]).

## 겪은 함정

- 경고 메시지에 개행을 안 붙여 응답 본문이 달라붙었다(`...unauthenticated request{"id":10,...}`). `fmt.Fprintf` 는 개행을 자동으로 붙이지 않는다.

## 관련

[[net-http]] · [[environment-variables]] · [[io-reader-writer]] · [[stdout-stderr]]
