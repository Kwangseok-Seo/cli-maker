# go-ast

Go 소스를 **문자열이 아니라 구문 트리로** 다루는 표준 라이브러리. `go/format`(판정·정렬) · `go/parser`(읽기) · `go/ast`(걷기).

## 왜 필요한가

[[text-template]] 은 텍스트만 만들 뿐 Go 문법을 모른다. 그래서 코드 생성기는 늘 **2단**이다 — 만들고, 파서에게 물어본다.

## `format.Source` — 검산이자 포매터

```go
var buf bytes.Buffer
if err := mainTmpl.Execute(&buf, data); err != nil { return err }

src, err := format.Source(buf.Bytes())
if err != nil {
	return fmt.Errorf("생성된 소스가 올바른 Go 가 아니다 (템플릿 결함): %w", err)
}
_, err = w.Write(src)
```

**버퍼를 거치는 것이 핵심이다.** `w` 로 바로 흘리면 검산 시점엔 이미 절반이 나가 있다. 템플릿을 일부러 깨뜨려(`root.AddCommand(sub)` → `root.AddCommand(sub`) 확인했다:

```
Error: 생성된 소스가 올바른 Go 가 아니다 (템플릿 결함): 49:22: missing ',' before newline in argument list (and 9 more errors)
exit=1
stdout 으로 나간 바이트 수: 0        ← 리다이렉트된 파일이 반쯤 깨진 채 남지 않는다
```

에러 문구가 *"템플릿 결함"* 이라고 말하는 것도 의도적이다. 유저 매니페스트는 이미 `Validate` 를 통과했으니, 여기서 깨졌다면 잘못은 생성기 쪽이다([[error-handling]]).

## `parser.ParseFile` + `ast.Inspect` — 생성물에서 사실을 꺼내기

"런타임 CLI 와 생성된 CLI 의 명령 표면이 같은가"를 테스트하려면 생성된 **텍스트**에서 표면을 꺼내야 한다. 정규식으로 훑지 않고 파서에게 묻는다.

```go
f, err := parser.ParseFile(token.NewFileSet(), "main.go", src, 0)

for _, decl := range f.Decls {
	fn, ok := decl.(*ast.FuncDecl)          // 타입 단언으로 노드 종류를 가른다
	if !ok || fn.Name.Name != "main" { continue }

	for _, stmt := range fn.Body.List {
		if block, ok := stmt.(*ast.BlockStmt); ok {   // 명령 하나 = 익명 블록 하나
			out = append(out, parseBlock(t, block))
		}
	}
}
```

`ast.Inspect(node, f)` 는 트리를 깊이 우선으로 걸으며 노드마다 `f` 를 부른다. `f` 가 `false` 를 반환하면 그 가지로 내려가지 않는다.

```go
ast.Inspect(block, func(n ast.Node) bool {
	switch v := n.(type) {
	case *ast.CompositeLit:                  // &cobra.Command{Use: ..., Short: ...}
		if !isCobraCommand(v.Type) { return true }   // &clirun.Command{} 도 걸리므로 가른다
		...
	case *ast.CallExpr:                      // sub.Flags().String(name, "", usage)
		sel, ok := v.Fun.(*ast.SelectorExpr)
		...
	}
	return true
})
```

노드 종류를 가르는 것은 전부 **타입 스위치**다([[interfaces]]). `ast.Node` 는 인터페이스이고 구체 타입이 문법 요소 하나하나에 대응한다.

## `strconv.Unquote` — 리터럴을 값으로 되돌린다

```go
lit, ok := e.(*ast.BasicLit)          // "say\"hi"  ← 소스에 적힌 그대로
s, err := strconv.Unquote(lit.Value)  // say"hi     ← 원래 값
```

`BasicLit.Value` 는 **인용부호를 포함한 소스 텍스트**다. 이걸 되돌리면 `%q` 로 넣은 값이 그대로 돌아오는지가 증명된다 — 별도의 인용 테스트를 두지 않고 이 왕복 하나로 확인했다([[text-template]] 의 함정 2).

## 이 테스트가 실제로 잡는가

통과하는 테스트는 실패할 수 있어야 한다. 네 곳을 일부러 어긋냈다:

| 어긋낸 곳 | 결과 |
|---|---|
| 템플릿의 Short 형식 `GET /x` → `GET: /x` | FAIL — `생성물 "GET: /plain", 런타임 "GET /plain"` |
| 템플릿에서 `MarkFlagRequired` 삭제 | FAIL — flag 의 Required 불일치 |
| **런타임**(`build.go`)의 Short 형식 변경 | FAIL — `생성물 "GET /plain", 런타임 "GET @ /plain"` |
| 템플릿 첫 줄 표식 변경 | FAIL — 표식 상수와 갈림 |

**양방향으로 잡힌다.** 생성기가 어긋나도, 런타임이 어긋나도 같은 테스트가 깨진다.

## 겪은 함정

- **AST 로 꺼낸 순서와 cobra 가 주는 순서가 다르다.** 생성물은 매니페스트 순서, `Flags().VisitAll` 은 이름순이다. 비교 전에 한쪽 기준으로 맞춰야 한다. 명령 순서는 `cobra.EnableCommandSorting = false` 로 맞췄고, 덕분에 **ADR-0002 의 순서 보존까지 같은 테스트가 지킨다**.
- **템플릿 설계가 테스트 난이도를 정했다.** 명령마다 익명 블록 `{ ... }` 을 쓴 것은 원래 식별자 생성을 피하려는 선택이었는데, 덕분에 명령의 경계가 구문적으로 뚜렷해져 블록 하나를 통째로 `Inspect` 하면 그 명령의 것만 모인다.
- **가드 자체는 테스트하지 않았다.** 깨진 템플릿을 주입하려면 프로덕션 코드에 테스트용 구멍이 필요해서, 대신 *"유저 입력이 생성을 깨뜨릴 수 없다"* 는 불변식을 본다. 가드는 손으로 한 번 밟아 확인했다(위 0 바이트).

## 관련

[[text-template]] · [[go-embed]] · [[interfaces]] · [[table-driven-tests]] · [[error-handling]] · [[go-toolchain]]
