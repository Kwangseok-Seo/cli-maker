# error-handling

Go 는 예외(exception)를 던지지 않는다. 함수가 **에러를 값으로 반환**하고, 호출한 쪽이 `if err != nil` 로 확인한다. 이 관용구가 Go 코드 전반을 지배한다.

## 핵심

```go
if err := rootCmd.Execute(); err != nil {
    os.Exit(1)
}
```

- `if err := f(); err != nil { ... }` — f 실행 → 에러를 `err` 에 받고 즉시 검사. `err` 은 이 `if` 블록 범위에서만 산다.
- 에러는 보통 마지막 반환값: `func f() (T, error)`.
- `os.Exit(1)` — 0 아닌 코드로 종료 = 실패 신호. 종료 코드 0 = 성공.

## 왜 중요한가

- 종료 코드로 스크립트·CI·파이프라인이 성공/실패를 판단한다 (`cmd && next`, `set -e`, `$LASTEXITCODE`).
- 에러를 무시하면 실패해도 0으로 끝나 "성공"으로 오해된다. (M1 의 `greet` 인자 누락이 그 예 — [[cobra]] 가 에러는 출력하지만, 종료 코드는 우리가 챙겨야 했다.)

## 관련

[[cobra]] · [[packages-and-main]] · [[go-toolchain]]
