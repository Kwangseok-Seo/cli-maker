# net-http

표준 라이브러리로 HTTP 요청을 보내고 응답을 받는다. cli-maker 의 Executor 가 서 있는 자리.

## 핵심 — 네 개의 객체

```go
req, err := http.NewRequestWithContext(ctx, "GET", url, nil) // 만든다 (아직 안 보냄)
resp, err := http.DefaultClient.Do(req)                       // 보낸다
resp.StatusCode  // 200, 404 …
resp.Status      // "404 Not Found" (코드 + 문구)
resp.Body        // io.ReadCloser — 아직 다 도착하지 않은 스트림
```

- **요청 생성과 전송이 분리**돼 있다. 그 사이에 헤더를 붙이거나 [[context]] 를 실을 수 있다.
- 네 번째 인자 `body` 는 보낼 본문. 본문 없는 GET 이면 `nil`.
- `http.DefaultClient` 는 패키지가 미리 만들어 둔 클라이언트. 타임아웃·프록시를 바꾸려면 `&http.Client{...}` 를 직접 만든다.

## 왜 http.Get 을 안 쓰나

`http.Get(url)` / `http.Post(...)` 편의 함수가 있지만 **메서드가 이름에 박혀 있다**. 우리 매니페스트는 `method: GET|POST|…` 를 데이터로 들고 있어 실행 시점에야 정해지고, 헤더(M5 인증)·context 도 실을 수 없다. 그래서 처음부터 `NewRequestWithContext` + `Do` 라는 제네릭 경로를 쓴다 — Command 하나마다 코드가 생기지 않는다는 [[cobra]] 동적 생성의 짝.

## 상태 코드는 에러가 아니다

```go
resp, err := client.Do(req)   // err == nil 인데 resp.StatusCode == 404 일 수 있다
```

- `Do` 의 `err` 은 **전송 실패**(DNS 실패, 연결 거부, 타임아웃)에만 채워진다. 서버가 "404" 라고 *성공적으로 답한* 것은 통신으로는 성공이다.
- 그래서 4xx/5xx 를 실패로 볼지는 **우리가 정하는 정책**이다. cli-maker 는 본문을 stdout 에 내보낸 뒤 `fmt.Errorf("HTTP %s", resp.Status)` 를 반환한다 → [[error-handling]] · [[M4]].

## 겪은 함정

- **`req.Body` 와 `resp.Body` 혼동.** 읽을 것은 서버가 돌려준 `resp.Body`. `req.Body` 는 우리가 보낸 본문이고 GET 이면 `nil` 이라 읽어도 아무것도 없다.
- **`resp` 는 에러일 때 `nil`.** `Do` 가 실패하면 `resp` 를 만지는 순간 패닉 → [[defer]] 의 순서 규칙.

## 관련

[[io-reader-writer]] · [[defer]] · [[context]] · [[error-handling]] · [[url-encoding]]
