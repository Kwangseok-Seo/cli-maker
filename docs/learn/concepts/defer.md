# defer

"이 함수가 끝날 때 이걸 해라"를 미리 등록한다. 정리(cleanup)를 자원을 얻은 자리 바로 옆에 적기 위한 문법.

## 핵심

```go
resp, err := client.Do(req)
if err != nil {
    return err
}
defer resp.Body.Close()   // 이 함수가 어떤 경로로 끝나든 실행된다
```

- 함수에 `return` 이 여럿이어도 **모든 경로**에서 실행된다. [[error-handling]] 의 조기 반환과 짝을 이룬다.
- 여러 개면 **역순**(LIFO)으로 실행된다.
- 응답 본문을 닫지 않으면 연결이 커넥션 풀로 돌아가지 못하고 샌다.

## 두 번째 용도 — cancel

```go
ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
defer cancel()
```

`cancel` 을 안 부르면 타이머가 만료될 때까지 자원이 남는다 → [[context]].

## 순서 규칙 — 검사가 먼저, defer 는 그다음

`Do` 가 실패하면 `resp` 는 `nil` 이다([[net-http]]). 그래서 이 순서는 터진다:

```go
resp, err := client.Do(req)
defer resp.Body.Close()   // ✗ err 검사보다 위
if err != nil { return err }
```

실측 결과 — 연결이 거부되는 포트로 요청했을 때:

```
Do 직후: resp == nil? true
         err = ... connectex: 연결이 거부되었습니다

panic: runtime error: invalid memory address or nil pointer dereference
main.main()
	deferorder.go:17            ← defer 를 쓴 그 줄
```

**패닉 지점이 함수 끝이 아니라 `defer` 를 쓴 줄이다.** `defer f.M()` 은 `M` 의 호출을 미루지만 **누구의 `M` 인지는 `defer` 를 만난 순간 평가**하기 때문. `resp.Body` 를 꺼내려다 거기서 즉시 죽는다.

## 겪은 함정

- 위 패닉의 진짜 손해는 크래시가 아니라 **원인 은폐**였다. `connectex: 연결이 거부되었습니다` 라는 정확한 진단이 이미 손에 있었는데, 패닉이 그걸 덮고 유저에겐 `invalid memory address` 만 남겼다. 잘못된 `defer` 순서는 버그를 만드는 게 아니라 **버그를 못 읽게 만든다** → [[errors-compile-vs-runtime]].

## 관련

[[net-http]] · [[context]] · [[error-handling]] · [[errors-compile-vs-runtime]] · [[io-reader-writer]]
