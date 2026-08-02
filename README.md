# cli-maker

여러 web API 를 **하나의 CLI** 로 다루게 해 주는 오픈소스 도구. API 명세(매니페스트)를 주면, cli-maker 가 런타임에 그걸 읽어 서브커맨드를 만들어 냅니다.

> 지향점: [mvanhorn/cli-printing-press](https://github.com/mvanhorn/cli-printing-press). 단, printing press 가 API마다 독립 바이너리를 *코드 생성* 하는 것과 달리, cli-maker 는 단일 바이너리가 매니페스트를 *런타임 해석* 합니다. 이유: [ADR-0001](docs/adr/0001-runtime-interpreter-architecture.md).

## 두 가지 목표

1. **도구** — web API 용 CLI 생성기.
2. **학습 실증** — 이 도구를 Go 로 만들어 가며 Go·프레임워크를 배우고, 그 과정을 오픈소스로 공개해 *AI 에이전트와 협업하면서 학습이 동시에 성립함* 을 보인다.

그래서 이 저장소는 "완성된 코드"가 아니라 **배우며 자라는 과정**을 커밋 히스토리로 남깁니다.

## 상태

M11 완료 — 매니페스트로 만들어진 명령이 **실제 API 를 호출**하고, **환경변수의 토큰으로 인증**하며, **보는 사람에 맞춰 출력을 냅니다**. **요청 본문도 보낼 수 있고**, 그 약속들을 **테스트가 지키며**, 매니페스트를 **그 API 전용 독립 바이너리로 뽑아낼 수도** 있습니다. 이제 매니페스트를 손으로 쓰지 않고 **OpenAPI 명세에서 만들어 낼 수도** 있습니다.

```
$ cli-maker gh --help
Available Commands:
  whoami      GET /user
  repo        GET /repos/{owner}/{repo}     ← apis/github.yaml 에서 생성된 명령

$ cli-maker gh whoami                       # 토큰 없이
warning: GITHUB_TOKEN is not set — sending unauthenticated request    ← stderr
{"message":"Requires authentication",...}
Error: HTTP 401 Unauthorized                                          ← 종료 코드 1

$ GITHUB_TOKEN=<진짜 토큰> cli-maker gh whoami
{"login":"...","id":...}                                              ← 인증됨
```

매니페스트에는 **토큰이 아니라 어느 환경변수를 볼지**만 적습니다(`auth: {type: bearer, env: GITHUB_TOKEN}`). 토큰이 없으면 거부하지 않고 경고만 남기고 인증 없이 보냅니다 — 익명으로도 되는 엔드포인트를 막지 않기 위해서입니다([ADR-0006](docs/adr/0006-missing-token-policy.md)).

Param 은 모두 flag 로 드러나고(`--이름 값`), 본문은 stdout 으로만 나가므로 `| jq` 가 안전하며, 4xx/5xx 는 종료 코드 1 로 알립니다.

출력은 **누가 보고 있는지**에 따라 달라집니다 — 터미널이면 읽기 좋게 들여쓰고, 파이프나 리다이렉트면 받은 바이트를 그대로 냅니다([ADR-0008](docs/adr/0008-output-format-defaults-to-tty.md)).

```
$ cli-maker gh repo --owner spf13 --repo cobra        # 터미널
{
  "id": 12574344,
  "name": "cobra",                                    ← 서버가 보낸 필드 순서 그대로
  …

$ cli-maker gh repo --owner spf13 --repo cobra | jq -r .stargazers_count
44314                                                 ← 파이프엔 원본 바이트가 간다
```

`-o raw|pretty|compact` 로 직접 고를 수 있습니다. 응답이 JSON 이 아니면(`cli-maker gh zen` 은 평문 한 줄) 원본을 그대로 내보내고 종료 코드도 건드리지 않습니다 — 명시한 경우에만 stderr 로 한 줄 알립니다.

요청 타임아웃은 세 층에서 정해집니다 — `--timeout 5s` > `CLI_MAKER_TIMEOUT=60s` > 기본 30초.

## 요청 본문

매니페스트에 `body:` 를 적은 명령에만 `--data` 가 붙습니다 ([ADR-0010](docs/adr/0010-request-body-as-command-field.md)). 본문은 이름 붙은 입력이 아니므로 Param 이 아니라 Command 의 필드입니다 — OpenAPI 도 2.0 의 `in: body` 파라미터에서 3.0 의 `requestBody` 별도 필드로 옮겨갔습니다.

```yaml
commands:
  - name: addPet
    method: POST
    path: /pet
    body:
      required: true                 # --data 를 안 주면 요청 전에 막힌다
      contentType: application/json  # 생략하면 application/json
```

curl 계보의 세 입력원을 받습니다.

```
$ cli-maker pstore addPet --data '{"name":"rex"}'          # 리터럴
$ cli-maker pstore addPet --data @pet.json                 # 파일
$ echo '{"name":"rex"}' | cli-maker pstore addPet --data - # stdin

$ cli-maker pstore addPet
Error: required flag(s) "data" not set        ← 요청은 나가지 않습니다

$ cli-maker pstore getPetById --petId 1 --data x
Error: unknown flag: --data                   ← body: 가 없는 명령엔 flag 자체가 없습니다
```

`--data ''` 는 **빈 본문을 보냅니다** — `--data` 를 아예 안 준 것(본문 없음)과 다릅니다. stdin 의 개행은 지우지 않습니다(curl 의 `--data-binary` 쪽 동작) — 지우는 것은 payload 를 조용히 바꾸는 일이라서입니다.

어떤 입력원이든 본문을 다 읽어 길이를 알 수 있는 형태로 보냅니다. 스트리밍으로 흘리면 요청이 `chunked` 로 나가고 리다이렉트 재전송이 조용히 실패합니다.

`apis/` 에 매니페스트 YAML(`.yaml` 또는 `.yml`)을 하나 더 떨어뜨리면 **재컴파일 없이** 새 API 명령이 생깁니다 ([ADR-0003](docs/adr/0003-dynamic-command-surface.md)).

매니페스트는 **두 곳**에서 찾습니다 — 현재 디렉토리의 `./apis` 와 유저 설정 디렉토리(`os.UserConfigDir()` 아래의 `cli-maker/apis`. Windows `%AppData%`, macOS `~/Library/Application Support`, Linux `$XDG_CONFIG_HOME` 또는 `~/.config`). 한쪽을 찾았다고 끝내지 않고 **둘 다 읽어 합칩니다** ([ADR-0013](docs/adr/0013-manifest-discovery-paths.md)).

```
$ cd ~/anywhere && cli-maker --help    # apis/ 가 없는 곳에서도
  gh          gh API                   ← 설정 디렉토리에서

$ cd ~/myproject && cli-maker --help
  internal    internal API             ← ./apis/ 에서
  gh          gh API                   ← 설정 디렉토리에서
```

이름이 겹치면 가까운 쪽(`./apis`)이 이기고, **가려진 쪽은 조용히 사라지지 않습니다.**

```
$ cli-maker gh --help                  # 양쪽에 gh 가 있을 때
cli-maker: C:\Users\...\cli-maker\apis\gh.yaml 생략
name "gh" 는 이미 쓰이고 있는 명령 이름이다    ← stderr
```

> 실행 파일 옆 경로·환경변수·`--manifest` override 는 아직 없습니다.

매니페스트는 명령으로 등록되기 전에 검증을 통과해야 합니다. 문제가 있으면 그 파일만 빠지고 나머지는 그대로 동작합니다 ([ADR-0007](docs/adr/0007-load-time-validation.md)).

```
$ cli-maker greet 철수
cli-maker: apis\broken.yaml 생략
baseURL "GET TT": scheme 이 http/https 가 아니다
commands[1] "ping": method "get" 는 지원하지 않는다
commands[2] "ping": 이름이 중복이다
안녕, 철수!                       ← 다른 명령은 살아 있습니다
```

위 약속들은 문서로만 있지 않습니다. `go test ./...` 가 4xx 의 **출력 순서**(ADR-0005), 토큰이 없을 때의 **익명 전송**(ADR-0006), 터미널 여부에 따른 **기본 포맷**(ADR-0008)을 확인합니다. HTTP 는 목(mock) 없이 `httptest` 로 진짜 서버를 띄워, 서버가 실제로 받은 경로·쿼리·헤더로 판정합니다.

## OpenAPI 명세에서 만들기

`import` 는 OpenAPI 3.x 명세(**Spec**)를 읽어 매니페스트를 냅니다 ([ADR-0011](docs/adr/0011-importer-reads-five-places-directly.md), [ADR-0012](docs/adr/0012-import-output-is-a-deterministic-draft.md)). `.json` 도 `.yaml` 도 됩니다.

```
$ curl -sO https://petstore3.swagger.io/api/v3/openapi.json

$ cli-maker import openapi.json --out apis/pstore.yaml \
      --base-url https://petstore3.swagger.io/api/v3
cli-maker: securityScheme [api_key petstore_auth] 를 옮기지 않았다 — auth 는 손으로 적어야 한다
cli-maker: deletePet: param "api_key" 를 뺐다 — in "header" 는 지원하지 않는다   ← stderr
생성: apis/pstore.yaml (명령 19개)

$ cli-maker pstore getPetById --petId 1 -o compact
{"id":1,"name":"Pet1","photoUrls":[…],"tags":[],"status":"available"}
```

**19 operation 이 전부 넘어오고 재컴파일은 없습니다** — 산출물이 매니페스트라, 런타임이 읽던 그대로 읽습니다.

산출물은 **초안**입니다. 옮기지 못한 것은 빼고 stderr 에 한 줄씩 남깁니다 — 반쪽을 조용히 내보내면 유저는 요청이 나간 뒤에야 알게 되니까요. 명세 자체는 stdout 으로만 나가므로 `import spec.json --name pstore > apis/pstore.yaml` 도 안전합니다.

| 자리 | 어떻게 되나 |
|---|---|
| `servers[0].url` 이 상대 URL (petstore 가 그렇다) | `--base-url` 로 받습니다. 없으면 에러 |
| `paths` 가 map 이라 순서가 없음 | 경로를 **사전순**으로 냅니다 — 두 번 import 해도 같은 파일 |
| `securitySchemes` (oauth2·apiKey) | 대응이 없어 `auth` 를 비워 두고 알립니다 |
| `in: header` param | optional 이면 그것만 빼고, **required 면 그 명령을 통째로** 뺍니다 |
| `requestBody.content` 후보가 여럿 | `application/json` 을 먼저, 없으면 사전순 첫 번째 |
| Swagger **2.0** | 거부합니다 — 이름을 대고, 왜인지 함께 |

이름은 `--name` > `--out` 파일명 > 에러 순으로 정합니다. `info.title` 은 쓰지 않습니다 — petstore 의 경우 `swagger-petstore-openapi-3-0` 이 되어 명령 이름이 못 됩니다.

기존 파일은 **덮어쓰지 않고 거절합니다.** 손으로 이어 쓴 편집분을 두 번째 import 가 조용히 지우면 되돌릴 방법이 없기 때문입니다.

## 독립 바이너리로 뽑기

`generate` 는 같은 매니페스트를 읽어, 실행하는 대신 **그 API 전용 CLI 의 Go 소스**를 냅니다 ([ADR-0009](docs/adr/0009-generated-cli-shape.md)).

```
$ cli-maker generate apis/github.yaml --dir out/gh
생성: out\gh\main.go
생성: out\gh\go.mod

$ cd out/gh && go mod tidy && go build -o gh .

$ ./gh repo --owner spf13 --repo cobra -o compact
{"id":12574344,"node_id":"MDEwOlJlcG9zaXRvcnkxMjU3NDM0NA==",…      ← yaml 없이 돕니다
```

> 생성된 `go.mod` 에는 `require` 를 적지 않습니다 — `go mod tidy` 가 실제 import 를 보고 cli-maker 버전을 스스로 채웁니다.

**생성된 코드는 당신 것입니다.** 원하는 라이선스로 두면 되고 cli-maker 의 MIT 고지를 이어붙일 필요가 없습니다. 다만 그 CLI 를 빌드하면 [spf13/cobra](https://github.com/spf13/cobra)(Apache-2.0)·[spf13/pflag](https://github.com/spf13/pflag)(BSD-3-Clause) 가 정적 링크되므로, **바이너리를 재배포할 때는** 그쪽 고지가 필요합니다.

생성된 코드는 명령마다 `cobra.Command` 가 펼쳐져 있어 읽을 수 있고, 실행은 런타임 경로와 **같은 함수**(`clirun.Run`)를 부릅니다. 두 CLI 의 명령 표면이 갈리지 않는지는 테스트가 생성 소스를 파싱해 런타임 트리와 대조하며 지킵니다.

매니페스트에 어떤 이름을 적어도 생성물이 깨지지 않습니다 — 유저 문자열은 전부 `%q` 로 인용되고, 생성된 소스는 `go/format` 검산을 통과한 뒤에야 파일로 나갑니다.

## 설치

**바이너리** — [Releases](https://github.com/Kwangseok-Seo/cli-maker/releases) 에서 받아 풀고 `cli-maker` 를 PATH 에 둡니다. Go 가 필요 없습니다. linux·macOS·Windows 의 amd64/arm64 를 냅니다.

```
$ sha256sum -c checksums.txt        # 받은 것 검증
cli-maker_v0.2.0_linux_amd64.tar.gz: OK
```

**Go 로** —

```
go install github.com/Kwangseok-Seo/cli-maker@latest
```

Go 1.26+ 필요. 설치한 뒤 무엇을 받았는지는 `--version` 으로 확인합니다.

```
$ cli-maker --version
cli-maker v0.2.0 (go1.26.3 windows/amd64)
```

버전 문자열은 소스 어디에도 적혀 있지 않습니다 — go 도구가 빌드할 때 박아 둔 것을 읽습니다. 그래서 저장소를 클론해 직접 빌드하면 태그 대신 커밋을 가리키는 pseudo-version 이 나오고, 커밋 안 된 변경이 섞여 있으면 뒤에 `+dirty` 가 붙습니다.

## 개발

- 빌드: `go build ./...`
- 실행: `go run .`
- 테스트: `go test ./...`

Go 1.26+ 필요.

## 학습 로드맵

각 마일스톤 = 배우는 Go 개념 하나 + 기능 조각 하나.

| 마일스톤 | 기능 | 배우는 개념 |
|---|---|---|
| **M0** ✅ | 프로젝트 골격, `go run` | 모듈 시스템, 패키지, 진입점 |
| **M1** ✅ | Cobra 루트 + 첫 서브커맨드 | 의존성 관리(go get/go.sum), 명령 트리, 인터페이스 |
| **M2** ✅ | 매니페스트 파싱 (YAML→struct) | struct 태그, 언마샬, 에러 처리 |
| **M3** ✅ | 매니페스트→명령 동적 생성 | 클로저 캡처, 컴포지트 리터럴, 디렉토리 스캔 |
| **M4** ✅ | 제네릭 HTTP 실행기 | net/http, io.Reader/Writer, defer, context |
| **M5** ✅ | 인증·설정 (env 토큰) | os 패키지, 설정 우선순위 |
| **M6** ✅ | 매니페스트 검증 (등록 전에 막는다) | map(집합), `errors.Join` |
| **M7** ✅ | 출력 포맷 (`-o raw\|pretty\|compact`) | encoding/json, 인터페이스, 타입 단언 |
| **M8** ✅ | 테스트 | table-driven test, httptest |
| **M9** ✅ | `generate` (코드 생성) | text/template, `//go:embed`, go/ast, 타입 별칭 |
| **M10** ✅ | 요청 본문 (`--data`) | 값 vs 포인터, `io.Reader` 를 요청 쪽으로, chunked |
| **M11** ✅ | OpenAPI 임포트 (`import`) | map 의 무순서, 부분 디코드, `yaml.Marshal` |
| **M12** 🚧 | 배포 — 발견 경로 ✅ · `--version` ✅ · v0.2.0 발행 ✅ · 바이너리 릴리스 ✅ · CI ✅ | `os.UserConfigDir`, `debug.ReadBuildInfo`, semver·proxy, `GOOS`/`GOARCH`, GitHub Actions |

> 각 마일스톤에서 쌓은 이론·문법·겪은 함정은 [`docs/learn/`](docs/learn/) 에 개념별 지식베이스로 정리합니다 — "무엇을 배웠나"의 실증.

## 라이선스

MIT — [LICENSE](LICENSE). `generate` 가 낸 소스는 당신 것입니다(위 참조).

의존성은 전부 permissive 이지만 MIT 는 아닙니다 — cobra·mousetrap Apache-2.0, pflag BSD-3-Clause, yaml.v3 MIT+Apache-2.0. 소스로 쓸 때는 추가 의무가 없고, **바이너리를 배포할 때** 그쪽 고지가 필요합니다.

테스트 fixture `internal/openapi/testdata/petstore.json` 은 [Swagger Petstore](https://petstore3.swagger.io/api/v3/openapi.json) 명세 원본(v1.0.27)이며, 그 문서가 `info.license` 에 스스로 밝힌 대로 Apache-2.0 입니다 — 테스트에서만 읽고 바이너리에는 들어가지 않습니다.
