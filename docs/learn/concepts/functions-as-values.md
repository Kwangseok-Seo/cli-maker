# functions-as-values

Go 에서 함수는 **값**이다 — 변수·struct 칸에 담고, 인자로 넘기고, 반환할 수 있다.

## 핵심

```go
Run: func(cmd *cobra.Command, args []string) { ... }  // struct 칸에 함수를 담음
```

- cobra 는 이 `Run` 함수를 명령 실행 시 호출한다 → [[cobra]].
- 함수가 바깥 변수를 붙잡는 것 = **클로저**(예: 반복문에서 `spec := spec` 로 값 캡처).

## 관련

[[structs]] · [[cobra]]
