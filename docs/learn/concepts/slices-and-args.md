# slices-and-args

슬라이스(`[]T`)는 순서 있는 리스트. 명령줄 인자는 `os.Args`(`[]string`)로 들어온다 — CLI 의 입력 통로.

## 핵심

```go
os.Args     // []string — 명령줄 인자들
os.Args[0]  // 프로그램 경로 (go run 은 임시 exe)
os.Args[1]  // 사용자가 친 첫 인자
```

- 인덱스는 **0부터**.
- `go run . 안녕 world` → `[..exe 안녕 world]` (3개).

## 겪은 함정

- 인자 없이 `os.Args[1]` → 런타임 패닉 (index out of range). 인덱스 전 길이 확인이 필요 → [[cobra]] 도입 동기.
- **해소**: [[cobra]] 의 `Args: cobra.ExactArgs(n)` 가 인덱싱 전에 개수를 검증 → panic 대신 친절한 에러. `Run` 의 `args[0]` 은 그 뒤라 안전.

## 관련

[[variables]] · [[errors-compile-vs-runtime]] · [[cobra]]
