# variables

값을 담는 이름. Go 는 정적 타입이라 한번 정해진 타입은 고정된다.

## 핵심

```go
greeting := "hello"        // 선언+대입, 타입 자동 추론 (함수 안에서만)
greeting = "안녕"           // 재대입 (이미 있는 변수)
var name string = "world"  // var 명시 선언 (함수 밖에선 필수)
var n int                  // 값 없으면 zero value (int→0, string→"", bool→false)
```

- `:=` = declare+assign, `=` = assign only.
- 문자열은 `+` 로 이어붙인다.

## 겪은 함정

- 없는 변수에 `=` 만 쓰면 `undefined: greeting` → [[errors-compile-vs-runtime]].
- 선언하고 안 쓰면 컴파일 거부 (`declared and not used`) — Go 의 죽은 코드 봉쇄.

## 관련

[[packages-and-main]] · [[errors-compile-vs-runtime]] · [[slices-and-args]]
