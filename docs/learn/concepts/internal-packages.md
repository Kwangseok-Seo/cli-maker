# internal-packages

`internal/` 아래 패키지는 **그 `internal` 의 부모 디렉토리를 뿌리로 하는 트리 안에서만** import 할 수 있다. 관행이 아니라 **컴파일러가 강제**한다.

## 먼저 — 패키지와 모듈은 다르다

| | 무엇 | 우리 repo |
|---|---|---|
| **패키지** | 디렉토리 하나. `import` 의 단위 | `internal/manifest`, `clirun`, … |
| **모듈** | `go.mod` 하나가 지배하는 트리. 버전·배포의 단위 | repo 전체 = 모듈 **하나** |

우리 `internal/` 의 부모는 모듈 루트다. 그래서 **이 모듈 안에서는 누구나** 쓸 수 있고, **모듈 밖에서는 아무도** 못 쓴다.

## 실측 — 두 번 같은 메시지

외부 모듈에서 `replace` 로 우리 repo 를 가리키고 빌드:

```
main.go:6:2: use of internal package github.com/Kwangseok-Seo/cli-maker/internal/executor not allowed
exit=1
```

축소판으로 다시:

```
main.go:7:2: use of internal package example.com/climaker/internal/hidden not allowed
exit=1
```

`replace` 로 소스를 손에 쥐어 줘도 막는다. 파일 접근 권한이 아니라 **import 경로 규칙**이기 때문이다.

## 그래서 façade 를 낸다

생성된 CLI 는 자기 `go.mod` 를 가진 **남의 모듈**이라 우리 `internal/` 을 못 쓴다. 전부 공개하는 대신 **좁은 창구 하나**만 낸다:

```go
// clirun (모듈 루트의 공개 패키지)
package clirun

import "github.com/Kwangseok-Seo/cli-maker/internal/executor"   // ← 같은 모듈이라 통과

func Run(cmd *cobra.Command, m *Manifest, c *Command) error { ... }
```

```
façade 빌드: OK (공개 패키지가 internal 을 import 함)
외부 모듈이 façade 경유로 internal 로직에 도달: OK
```

**모듈 안의 어떤 공개 패키지든 `internal/` 을 쓸 수 있다.** 그래서 "감춘다"와 "못 쓴다"가 같은 말이 아니다 — 무엇을 통과시킬지 우리가 고른다.

우리가 공개한 것은 타입 별칭 4개([[type-aliases]]) + 상수 2개 + 함수 1개다. `executor.Execute` 의 7인자 시그니처도 `format.Formatter` 인터페이스도 밖에서 안 보이므로, 그것들은 계속 자유롭게 고칠 수 있다.

## 다른 언어에선

| | 강제 여부 |
|---|---|
| Go `internal/` | 컴파일러가 막는다 |
| Java `module-info.java` 의 `exports` | 컴파일러/런타임이 막는다 |
| Python `_private` | **관행일 뿐** — 아무도 안 막는다 |

`exports com.example.clirun;` 만 적고 나머지 패키지를 빼면 우리 구조와 같은 모양이 된다.

## 겪은 함정

- **`replace` 를 쓰면 될 줄 알았다.** 로컬 소스를 직접 가리키는데도 막힌다. 규칙은 "받을 수 있느냐"가 아니라 "부를 수 있느냐"다.
- **공개는 되돌리기 비싸다.** `internal/` 에서 꺼내는 순간 그 시그니처가 외부 계약이 된다. 그래서 "필요하니까 executor 를 통째로 공개"가 아니라 **생성물이 실제로 부르는 것만** 통과시켰다.
- **`%T` 는 여전히 internal 경로를 보여 준다** — 접근을 막는 것과 이름을 감추는 것은 다르다([[type-aliases]]).

## 관련

[[type-aliases]] · [[go-modules]] · [[packages-and-main]] · [[text-template]]
