# identifier-shadowing

`byte`·`len`·`copy` 는 **예약어가 아니라 미리 선언된 이름**일 뿐이다. 변수로 쓰면 조용히 가려지고, 컴파일러도 `go vet` 도 아무 말 하지 않는다.

## 핵심

Go 의 예약어는 25개뿐(`func`, `if`, `range`, …). 그 밖의 익숙한 이름들은 전부 **universe 스코프에 미리 선언된 식별자**다.

```
타입   bool byte rune int int8..64 uint... float32/64 complex64/128 string error any
함수   append cap clear close complex copy delete imag len make max min new panic
       print println real recover
값     true false iota nil
```

이것들은 가려도 **에러가 아니다** — 안쪽 스코프의 선언이 이긴다.

## 겪은 함정 ① — `byte` 를 변수 이름으로

```go
byte, err := io.ReadAll(src)   // 빌드 OK, go vet OK
```

빌드도 vet 도 통과한다. 대가는 나중에 온다 — 그 스코프에서 `byte` 는 더 이상 타입이 아니다.

```
.\main.go:10:16: byte (local variable) is not a type
```

`[]byte` 나 `make([]byte, n)` 을 쓰려는 순간 터진다. `body` 로 바꾸면 끝나는 일이지만, **에러가 선언 지점이 아니라 사용 지점에서** 나므로 원인이 멀어 보인다.

## 겪은 함정 ② — 함수 이름을 변수로

```go
isTerminal := isTerminal(cmd.OutOrStdout())
//  ^^^^^^ 이 줄 이후로 함수 isTerminal 을 못 부른다
```

오른쪽은 아직 함수, 왼쪽부터는 변수다(`:=` 는 선언을 **끝낸 뒤** 스코프에 넣는다). 한 번만 부르면 돌아가므로 발견이 늦다. 애초에 `if isTerminal(cmd.OutOrStdout())` 로 바로 쓰면 변수가 필요 없다.

## 겪은 함정 ③ — 루프 안에서 명명된 반환값을 `:=` 로

```go
func f() (warnings []string) {
	for _, p := range items {
		warnings := append(warnings, "…")   // 컴파일 OK, go vet OK
		_ = warnings
	}
	return                                  // 언제나 nil
}
```

```
=  쓴 것: [뺐다: api_key 뺐다: x-trace] (len=2)
:= 쓴 것: []                              (len=0)
```

함수 몸통 맨 위였다면 `:=` 가 *"no new variables on left side"* 로 막힌다. 그런데 **루프 몸통은 안쪽 블록**이라 "같은 이름의 새 지역변수"가 합법적으로 만들어지고, 루프가 끝나면 사라진다. 바깥 `warnings` 는 그대로 `nil`.

증상이 "에러"가 아니라 **"아무 일도 안 일어남"** 이라 앞의 둘보다 찾기 어렵다. 누적할 때는 `=`.

## 왜 Go 는 이걸 막지 않나

`nil`·`error` 조차 가릴 수 있는 건 언어 사양을 단순하게 유지하려는 선택이다. 예약어 목록이 짧을수록 컴파일러가 단순하고, universe 스코프는 "가장 바깥 블록"일 뿐이라 스코프 규칙이 하나로 통일된다.

대신 **도구가 안 잡아 준다.** `gofmt`·`go vet` 어느 쪽도 경고하지 않는다([[go-toolchain]]). 사람이 이름을 고르는 수밖에 없다.

## 실용 규칙

- `[]byte` 를 담는 변수는 `body`·`data`·`buf`.
- 개수는 `n`·`count` (`len` 금지).
- 함수와 같은 이름의 변수를 만들지 않는다. 조건문에 바로 넣으면 대개 변수 자체가 불필요하다.

## 관련

[[variables]] · [[errors-compile-vs-runtime]] · [[go-toolchain]] · [[slices-and-args]]
