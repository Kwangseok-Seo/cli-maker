# structs

이름 붙은 칸(필드)들의 묶음. Go 에서 데이터를 구조화하는 핵심 수단이자 이 프로젝트(Manifest, Command)의 토대.

## 핵심

```go
type Point struct { X int; Y int }  // 새 타입 정의
p := Point{X: 3, Y: 5}              // 값 생성
p.X                                 // 3 — 점으로 칸 접근
```

- 칸에는 함수도 담을 수 있다 → [[functions-as-values]].
- 라이브러리가 정의한 struct 를 채워 쓰기도 한다 → [[cobra]] 의 `cobra.Command`.

## 관련

[[pointers]] · [[functions-as-values]] · [[cobra]]
