# functions-as-values

Go 에서 함수는 **값**이다 — 변수·struct 칸에 담고, 인자로 넘기고, 반환할 수 있다.

## 핵심

```go
Run: func(cmd *cobra.Command, args []string) { ... }  // struct 칸에 함수를 담음
```

- cobra 는 이 `Run` 함수를 명령 실행 시 호출한다 → [[cobra]].
- 함수가 바깥 변수를 붙잡는 것 = **클로저** → [[closures]].

## 겪은 함정

- 처음 이 페이지에는 "반복문에서 `spec := spec` 로 **값** 캡처"라고 적었는데 두 군데가 틀렸다: 클로저는 값이 아니라 **변수(칸)** 를 붙잡고, 그 자기대입 관용구는 Go 1.22 부터 불필요하다. M3 에서 실측으로 교정 → [[closures]].

## 관련

[[structs]] · [[cobra]] · [[closures]]
