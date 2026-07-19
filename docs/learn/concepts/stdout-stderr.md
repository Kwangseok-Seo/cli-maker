# stdout-stderr

프로그램의 두 출력 통로 — 하나는 **데이터**, 하나는 **진단**. 섞지 않는다.

## 핵심

```go
fmt.Println("결과 데이터")                 // stdout(표준 출력)
fmt.Fprintln(os.Stderr, "parse:", err)     // stderr(표준 에러)
```

- **stdout**: 프로그램의 진짜 결과물. 파이프·리다이렉트(`> out.txt`)의 대상.
- **stderr**: 에러·진단 메시지. 리다이렉트해도 화면에 남는다.
- 분리 이유: `parse x.yaml > result.txt` 해도 에러는 눈에 보이고, result.txt 에 에러가 안 섞인다.
- `fmt.Fprintln(w, …)` = `Println` 의 "목적지(`w`) 지정" 버전. 콤마로 넘긴 인자엔 공백 자동, `err` 은 그대로 넘겨도 문자열이 된다(`.Error()` 불필요).

## 겪은 함정

- `fmt.Println("parse err" + err.Error())` → stdout 으로 나가고 공백 없이 `parse erropen…` 으로 뭉개짐. 고치면 `parse: open…`(stderr + 공백 정상). 에러 출력의 표준 짝은 **stderr + `os.Exit(1)`** → [[error-handling]].

## 관련

[[error-handling]] · [[go-toolchain]]
