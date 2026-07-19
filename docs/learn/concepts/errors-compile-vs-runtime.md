# errors-compile-vs-runtime

Go 에러는 두 시점으로 갈린다: **컴파일 시점**(실행 전)과 **런타임**(실행 중). 어느 쪽인지가 디버깅의 출발점.

## 핵심

- **컴파일 에러** — 실행 전에 걸림. 예: `undefined: greeting`. 형식 `파일:줄:열`, 여러 개를 한 번에 보고.
- **런타임 패닉** — 실행 중 터짐. 예: `panic: runtime error: index out of range [1] with length 1`.
- 정적 타입·컴파일 언어라 많은 실수가 런타임 전에 걸린다 (파이썬 등 인터프리터와 대비).

## 관련

[[variables]] · [[slices-and-args]] · [[go-toolchain]]
