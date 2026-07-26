# go-modules

프로젝트의 의존성 관리 체계. `go.mod`(무엇이 필요한가) + `go.sum`(받은 게 진짜인가)로 구성.

## 핵심

- **패키지 주소 = import 경로.** cobra 는 `github.com/spf13/cobra`. 우리 모듈은 `github.com/Kwangseok-Seo/cli-maker`.
- `go.mod` 의 `require`: **direct**(내 코드가 import) vs `// indirect`(딸려온 의존성의 의존성).
- `go.sum`: 다운로드 코드의 체크섬(지문) → 공급망 보안. 모듈당 두 줄(`h1:` 코드, `/go.mod`).
- 버전 `v1.10.2` = 주.부.수(semver). major=호환 깨짐.

## 네 줄이 각각 다른 일을 한다

```
module github.com/Kwangseok-Seo/cli-maker   ← ① import 경로의 뿌리
go 1.26                                     ← ② 언어 버전
require github.com/spf13/cobra v1.10.2      ← ③ 의존성 + 버전
require ( ... // indirect )                 ← ④ 딸려온 것
```

repo 사본에서 하나씩 망가뜨려 확인했다.

### ① `module` — 자기 자신도 이 이름으로 부른다

```
$ sed -i 's|^module .*|module example.com/mytool|' go.mod && go build ./...
main.go:8:2: no required module provides package github.com/Kwangseok-Seo/cli-maker/internal/cli
main.go:9:2: no required module provides package .../internal/manifest
```

이름을 바꾼 순간 **자기 repo 안의 코드가 서로를 "그런 모듈 모른다"고 한다.** 디렉토리 이름이 아니라 이 줄이 import 경로를 정하기 때문이다.

### ② `go 1.26` — 최소 버전이 아니라 **언어 의미론 스위치**

이 한 줄만 `go 1.21` 로 바꾸고 코드는 한 글자도 안 건드렸다:

| 명령 | `go 1.26` | `go 1.21` |
|---|---|---|
| `gh whoami` | `{"message":"Requires authentication",…}` exit **1** | `Encourage flow.` exit **0** |
| `gh zen` | `Approachable is better than simple.` exit 0 | `Encourage flow.` exit 0 |

`gh whoami` 가 **젠 명언을 반환한다.** 모든 서브커맨드가 매니페스트의 *마지막* 명령을 실행하고 있다.

원인은 [[closures]] 다. `for _, c := range m.Commands` 에서 클로저가 붙잡는 건 값이 아니라 칸인데, Go 1.22 가 그 칸을 반복마다 새로 만들도록 바꿨고 **그 변경은 이 줄로 모듈마다 켜고 꺼진다.** 우리 코드가 `&c` 를 넘겨도 안전한 건 이 줄 덕분이었다.

무서운 건 **아무도 에러를 안 본다**는 점이다 — exit 0 에 그럴듯한 응답.

### ③ `require` — 없으면 import 자체가 불가

```
$ grep -v 'spf13/cobra' go.mod > ... && go build ./...
internal\cli\build.go:9:2: no required module provides package github.com/spf13/cobra
```

### ④ `// indirect` — `go mod tidy` 가 제자리를 찾아 준다

`internal/manifest/load.go` 가 직접 import 하는데도 yaml 이 indirect 로 적혀 있었다:

```diff
-require github.com/spf13/cobra v1.10.2
+require (
+	github.com/spf13/cobra v1.10.2
+	gopkg.in/yaml.v3 v3.0.1        ← direct 로 승격
+)
 require (
-	gopkg.in/yaml.v3 v3.0.1 // indirect
 )
```

나중에 테스트가 `pflag` 를 직접 import 하자 같은 일이 한 번 더 일어났다 — **direct/indirect 는 실제 import 를 따라 움직인다.**

## `go.sum` — "무엇을"이 아니라 "그게 진짜인지"

```
github.com/spf13/cobra v1.10.2 h1:DMTTonx5m65Ic0GOoRY2c16WCbHxOOw6xxezuLaBpcU=
github.com/spf13/cobra v1.10.2/go.mod h1:7C1pvHqHw5A4vrJfjNwvOdzYu0Gml16OCs2GRiTUUS4=
```

해시 앞에 글자 하나(`X`)를 끼워 넣으면:

```
verifying github.com/spf13/cobra@v1.10.2: checksum mismatch
SECURITY ERROR
This download does NOT match an earlier download recorded in go.sum.
```

그래서 **둘 다 커밋한다.** `go.mod` 만 있으면 "cobra v1.10.2"라는 *이름*만 보장되지, 그 이름으로 온 *바이트*는 보장되지 않는다.

## 생성물의 `go.mod` 는 두 줄이면 된다

`generate --dir` 이 내는 것:

```
module gh

go 1.26
```

`require` 를 한 줄도 안 쓴다. 실측해 보니 `go mod tidy` 가 실제 import 를 보고 정확히 채운다(cobra `v1.10.2` 를 스스로 찾아냈다). **우리가 버전을 지어내면 그게 곧 거짓이 된다** — 특히 cli-maker 자신은 아직 배포 전이라 어떤 버전을 적어도 받아올 수 없다:

```
github.com/Kwangseok-Seo/cli-maker: git ls-remote ... exit status 128
```

`replace` 로 로컬 소스를 가리키면 `tidy` 가 그쪽에서 해결하고 `require` 까지 채워 준다.

## 다른 생태계와 대조

| `go.mod` 의 일 | Python | Java / Spring |
|---|---|---|
| `module` = import 경로의 뿌리 | **대응 없음** (경로는 디렉토리가 정한다) | **대응 없음** (경로는 `package` 선언이 정한다) |
| `go 1.26` = 언어 버전 | `requires-python` (설치 게이트일 뿐) | `<maven.compiler.release>` |
| `require` | `[project] dependencies` | `<dependencies>` |
| `// indirect` | 구분 없음 | pom 에 안 적음 |

두 가지가 Go 특유다.

- **모듈 이름 = import 경로.** `artifactId` 를 바꿔도 `import com.example.Foo;` 는 멀쩡하고, `pip install scikit-learn` 해도 `import sklearn` 이다. Go 는 그 둘이 같은 문자열이라 위 ① 실험이 성립한다.
- **lock 파일이 따로 없다.** `go.mod` 엔 범위 문법이 아예 없고 정확한 버전만 적는다. 여러 요구가 충돌하면 **가장 높은 것**을 고른다(Minimal Version Selection).

| | 선언(범위) | 확정(버전) | 검증(해시) |
|---|---|---|---|
| Python | `pyproject.toml` | `poetry.lock` | `poetry.lock` |
| **Go** | *(없음)* | **`go.mod`** | **`go.sum`** |

## 명령

- `go get <경로>` — 등록(go.mod) + 봉인(go.sum).
- `go mod tidy` — go.mod 를 실제 import 와 일치 (import 후 `// indirect` 제거).
- `go mod edit -replace A=B` — A 대신 로컬 경로/다른 모듈을 쓰게 한다. `pip install -e ../pkg` · Gradle `includeBuild` 와 같은 자리.

## 겪은 함정

- `gihub.com`(오타) → 호스트에 실제 요청 → HTTPS→HTTP 강등 거부. 경로는 네트워크로 해석되고 HTTPS 강제.
- import 전 `go mod tidy` → 안 쓰는 cobra 제거됨. tidy 는 import 뒤에.
- **`go` 줄을 "최소 요구 버전"으로만 읽었다.** 실제로는 컴파일러가 어떤 언어 규칙으로 읽을지를 정한다. 그래서 생성기가 뱉는 `go` 줄을 자동 추론이 아니라 **눈에 보이는 상수**로 두고, 1.22 하한을 테스트로 지킨다.

## 관련

[[packages-and-main]] · [[internal-packages]] · [[closures]] · [[cobra]] · [[go-toolchain]] · [[go-embed]]
