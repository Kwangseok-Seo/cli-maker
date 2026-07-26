# go-doc

선언 바로 위 주석이 그 선언의 문서다. 별도 문법이 없다 — `go doc` 이 소스에서 그대로 읽는다.

## 핵심

```go
// Package clirun 은 cli-maker 가 생성한 CLI 가 쓰는 공개 표면이다.
package clirun          // ← 바로 위 주석이 패키지 문서, go doc 의 첫 화면

// Run 은 API 명령 하나를 실행한다.
func Run(...) error     // ← 관례: 문장을 선언 이름으로 시작한다
```

- Go 1.19+ 는 doc comment 에 **구조**가 있다 — `# 제목`, 목록, 들여쓴 블록(코드로 렌더). `gofmt` 가 doc comment 도 정규화한다.
- `go doc -all ./pkg` 이 **exported 전량**을 낸다. 공개 표면이 무엇인지 기억에 의존하지 않고 도구에게 묻는 자리다.

## 코드로 표현할 수 없는 것의 자리

태그를 붙이면 공개 표면이 계약이 된다. 그런데 계약의 일부는 시그니처로 쓸 수 없다.

| 적어야 할 것 | 왜 코드로 안 되나 |
|---|---|
| "지원 소비자는 `generate` 가 낸 코드다" | 컴파일러가 소비자를 구분할 수 없다 |
| "cobra 가 새는 것은 의도적이다" | 시그니처는 사실만 말하고 **의도**를 말하지 않는다 |
| "v0.x — 좁히는 변경은 릴리스 노트에 적는다" | 코드에 그런 자리가 없다 |

doc 에 없으면 소비자는 알 방법이 없다. 그래서 doc 이 유일한 자리다.

## 별칭은 필드를 감춘다 (실측)

```
$ go doc ./clirun Manifest
type Manifest = manifest.Manifest
    (주석만 나오고 필드가 없다)
```

원본이 `internal/` 에 있어 `go doc` 이 펼치지 않고, 밖에서는 그 패키지를 볼 수도 없다([[type-aliases]] · [[internal-packages]]). **별칭은 타입을 통과시키지만 문서를 통과시키지 않는다.**

그래서 필드 모양을 doc 에 적었다. 그러면 그 목록이 **복제**가 된다 — `manifest` 쪽에 필드가 붙는 날 조용히 뒤처진다.

## 문서를 코드로 붙든다

양쪽을 손으로 적지 않는다. 필드는 `reflect` 로 실물에서, doc 은 파서로 소스에서 끌어온다.

```go
doc := aliasDoc(t)                                  // parser.ParseComments + GenDecl.Doc.Text()
for f := range reflect.TypeOf(Param{}).Fields() {   // Go 1.26 의 iterator
    if !strings.Contains(doc, f.Name) {
        t.Errorf("%s 가 별칭 doc 에 없다", f.Name)
    }
}
```

파일을 문자열로 훑지 않는 이유는 **주석이 어디 붙은 것인지는 위치가 아니라 구조가 정하기** 때문이다 — `GenDecl.Doc` 는 그 선언에 붙은 주석만 준다([[go-ast]]).

실측: `manifest.Param` 에 필드 하나를 더하니 `TestAliasFieldsAreDocumented` 가 즉시 깨졌다.

## 겪은 함정

- **한 `type (...)` 블록의 doc 이 항목마다 반복 출력된다.** 별칭 넷을 한 블록에 묶었더니 `go doc -all` 이 같은 문단을 네 번 냈다. 항목별 설명이 필요하면 블록을 쪼개야 한다.
- **문서에 적은 목록도 복제다.** "코드가 아니니 괜찮다"가 아니라, 컴파일러도 `vet` 도 안 잡으니 **오히려 더 조용히** 썩는다. 실물에서 끌어오는 대신 손으로 적었다면 그 자리에 자물쇠를 채워야 한다.
- **`go doc` 은 exported 만 낸다** — 그래서 표면이 좁은지 확인하는 도구로도 쓴다. cli-maker 의 공개 표면은 이걸로 7개(별칭 4 + `Run` + 상수 2)임을 확인했다.

## 관련

[[go-ast]] · [[type-aliases]] · [[internal-packages]] · [[go-test]] · [[go-toolchain]] · [[testing-the-filesystem]]
