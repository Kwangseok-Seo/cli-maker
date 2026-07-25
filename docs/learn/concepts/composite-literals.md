# composite-literals

`&cobra.Command{...}` 처럼 **값을 그 자리에서 조립하는 표현식**. 중괄호가 블록과 생김새는 같고 성질은 다르다.

## 핵심

| 종류 | 예 | 안에 올 수 있는 것 |
|---|---|---|
| 컴포지트 리터럴 | `&cobra.Command{Use: "x"}` | `필드: 값,` 쌍뿐 (표현식) |
| 블록 | `func … { }`, `for … { }`, `if … { }` | 문장(statement) |

- 리터럴은 **표현식**(값을 만드는 식)이라 `for`·`if` 같은 문장이 들어갈 자리가 없다.
- `Run: func(...) { ... }` 은 예외가 아니다 — `Run` 이라는 *필드에 담기는 값*이 함수이고, 함수가 자기 본문 블록을 갖는 것뿐 → [[functions-as-values]].
- `&T{...}` = 값을 조립하고 그 주소를 취한다 → [[pointers]].

## 겪은 함정 (M3)

리터럴 안에 루프를 넣어 컴파일 실패:

```go
group := &cobra.Command {
    for _, c := range m.Commands {   // ← 리터럴 안의 문장
```

```
build.go:12:3: syntax error: unexpected keyword for, expected expression
build.go:24:2: syntax error: non-declaration statement outside function body
```

- 두 번째 에러는 파서가 12번 줄에서 길을 잃은 **여파**일 뿐 별개 문제가 아니다 — 문법 에러는 **첫 줄부터** 고친다 → [[errors-compile-vs-runtime]].
- 교정: 리터럴을 먼저 닫아 값을 완성하고, 루프는 함수 본문의 문장으로 꺼낸다. 순서에도 의미가 있다 — 루프에서 `group.AddCommand` 를 부르려면 그 시점에 `group` 이 이미 존재해야 한다.
- `gofmt` 는 `&cobra.Command {` 의 공백을 붙이고 필드 값들을 세로 정렬해 준다 → [[go-toolchain]].

## 관련

[[structs]] · [[functions-as-values]] · [[pointers]] · [[errors-compile-vs-runtime]]
