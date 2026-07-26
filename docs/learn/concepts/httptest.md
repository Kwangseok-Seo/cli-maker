# httptest

목(mock)을 만들지 않는다. **진짜 HTTP 서버**를 `127.0.0.1` 의 빈 포트에 띄우고, 우리 코드에겐 목적지만 바꿔 준다.

## 핵심

```go
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	io.WriteString(w, `{"message":"Not Found"}`)
}))
t.Cleanup(srv.Close)

m := &manifest.Manifest{BaseURL: srv.URL}   // ← 바꾸는 건 이것뿐
```

`srv.URL` 은 `http://127.0.0.1:<임의포트>` 다. 실제로 TCP 를 듣고 있고, 우리 코드는 평소 경로(`http.DefaultClient.Do`)를 그대로 탄다([[net-http]]).

## 왜 의존성 주입이 필요 없었나

`Execute` 는 전역 `http.DefaultClient` 를 쓴다. 보통 "테스트하기 어려운 설계"로 배우는 모양인데 여기선 문제가 안 됐다 — **목적지가 이미 데이터(`BaseURL`)로 들어오기** 때문이다. 클라이언트를 갈아 끼울 필요 없이 서버를 우리 것으로 세우면 된다.

바꿔 말하면, 주입이 필요한지는 "전역을 쓰는가"가 아니라 **"바깥으로 나가는 결정이 데이터로 표현돼 있는가"** 로 갈린다.

## `http.HandlerFunc` — 함수가 인터페이스를 만족한다

`httptest.NewServer` 가 받는 것은 인터페이스다.

```go
type Handler interface{ ServeHTTP(ResponseWriter, *Request) }
```

우리가 넘긴 건 함수인데, 표준 라이브러리가 이렇게 이어 준다:

```go
type HandlerFunc func(ResponseWriter, *Request)                            // 함수 타입에 이름을 붙이고
func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) { f(w, r) }   // 거기에 메서드를 단다
```

`http.HandlerFunc(fn)` 은 **타입 변환**이고, 변환된 순간 `fn` 이 `ServeHTTP` 를 갖게 되어 인터페이스를 만족한다([[interfaces]]).

메서드는 struct 전용이 아니다 — **이름 있는 타입이면 무엇에든** 붙는다. 함수 타입에도([[methods]] · [[functions-as-values]]).

## 클로저가 목의 자리를 대신한다

```go
var gotReq *http.Request
srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	gotReq = r                       // ← 서버가 받은 것을 바깥으로 흘려보낸다
	io.WriteString(w, "ok")
}))
// ... Execute 실행 ...
if gotReq.URL.RawQuery != "sort=stars" { ... }
if gotReq.Header.Get("Authorization") != "Bearer sekret" { ... }
```

핸들러가 바깥 변수를 붙잡는다([[closures]]). `verify(client).send(captor.capture())` 를 대신하는 자리다.

판정 방향이 중요하다 — **우리 코드가 무엇을 했는지가 아니라 상대편이 무엇을 받았는지**를 본다. 그래서 순수 함수 테스트(`BuildURL`)가 덮지 못하는 "그게 정말 선을 타고 나갔는가"까지 걸린다.

## `Close` 는 기다린다

`httptest.Server.Close()` 는 **처리 중인 요청이 끝날 때까지 블록한다.** 서버가 300ms 자는 타임아웃 테스트에서 이게 눈에 보였다:

```
execute_test.go:242: Execute 가 돌아오기까지 50.3496ms
--- PASS: TestExecuteHonorsContextTimeout (0.30s)
```

`Execute` 는 제때(50ms) 끊었고, 나머지 250ms 는 정리 단계가 자는 핸들러를 기다린 시간이다. **보고된 테스트 시간을 동작 시간으로 읽으면 오진한다.**

## 겪은 함정

- **0.30s 를 보고 "타임아웃이 안 듣나?" 로 의심할 뻔했다.** 실제 대상(`Execute` 의 반환 시각)을 재고 나서야 `Close` 대기였음이 드러났다. 값이 예상과 다르면 **대상보다 측정 장치를 먼저 의심한다** — M7 의 TTY 측정에서 배운 것과 같은 형태.
- **타임아웃 판정은 문자열이 아니라 센티널로.** `errors.Is(err, context.DeadlineExceeded)` 는 `net/http` 가 `*url.Error` 로 감싼 것을 뚫고 찾는다([[error-wrapping]] · [[context]]). 메시지 문자열을 비교했다면 Go 가 문구를 다듬는 날 깨진다.

## 관련

[[go-test]] · [[net-http]] · [[interfaces]] · [[methods]] · [[closures]] · [[context]] · [[error-wrapping]]
