# packages-and-main

Go 프로그램의 뼈대. 모든 `.go` 파일은 **패키지**에 속하고, 실행 파일이 되는 특별한 패키지가 `main` 이다.

## 핵심

- `package main` — 이 패키지만 실행 파일이 된다. 다른 이름이면 라이브러리.
- `func main()` — 프로그램 진입점 (인자·반환 없음).
- `import "fmt"` — 표준 라이브러리 가져오기. 여러 개면 `import ( ... )` 로 묶고 표준/외부를 빈 줄로 그룹.
- `fmt.Println(...)` — `패키지.함수` 호출.
- **대문자로 시작 = 공개(exported), 소문자 = 비공개.** `Println` 이 대문자인 이유.
- 관례: `main.go` 는 얇게(호출만), 로직은 `cmd/`·`internal/` 로.

## 관련

[[go-toolchain]] · [[variables]] · [[go-modules]]
