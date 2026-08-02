# build-info

프로그램이 자기 버전을 아는 방법. 문자열을 소스에 적는 대신, go 도구가 바이너리에 박아 둔 것을 읽는다.

## 버전은 코드에 없다

`debug.ReadBuildInfo()` 는 링커가 넣어 둔 모듈 정보를 돌려준다 — 모듈 경로, 버전, Go 버전, 의존성 목록, 빌드 설정.

```go
bi, ok := debug.ReadBuildInfo()
if !ok {
    return "unknown"        // 모듈 모드로 빌드되지 않은 경우
}
return bi.Main.Version
```

## 무엇이 박히는지는 빌드 방식이 정한다 — 실측

| 빌드 방식 | `Main.Version` | vcs 설정 |
|---|---|---|
| `go install …@v0.1.0` | `v0.1.0` | **없음** (proxy 소스엔 git 이 없다) |
| `go build` (git repo, 깨끗) | `v0.1.1-0.20260802101701-26ab395c90b0` | `vcs.revision`·`vcs.time`·`vcs.modified=false` |
| `go build` (git repo, 더러움) | 위와 같고 뒤에 **`+dirty`** | `vcs.modified=true` |
| `go build` (git repo **밖**) | `(devel)` | 없음 |

두 번째 줄의 긴 문자열이 **pseudo-version** 이다 — `직전 태그 다음 버전` + `커밋 시각(UTC)` + `커밋 해시 12자리`. 태그가 없는 커밋을 semver 자리에 넣기 위해 go 가 합성한다 → [[go-modules]]

**커밋 해시와 시각이 이미 그 안에 들어 있으므로** `vcs.revision` 을 따로 꺼내 붙일 이유가 없다. `+dirty` 도 go 가 붙여 주므로 `vcs.modified` 를 읽어 분기할 필요가 없다.

## `go run` 은 스탬핑하지 않는다

```
go build 한 바이너리   cli-maker v0.1.1-0.20260802101701-26ab395c90b0+dirty (go1.26.3 windows/amd64)
go run . --version     cli-maker (devel) (go1.26.3 windows/amd64)
```

같은 소스, 같은 커밋인데 결과가 다르다. **개발 중 `go run` 으로 확인하면 영원히 `(devel)` 이라, 코드가 잘못됐다고 의심하게 된다.** [[user-config-dir]] 의 `os.Executable()` 이 `go run` 에서 임시 경로를 주는 것과 같은 계열 — `go run` 은 임시 빌드라 프로덕션 빌드의 성질을 갖지 않는다 → [[go-toolchain]]

git repo 밖에서 빌드해도 `(devel)` 이다. CI 가 소스 아카이브(zip/tarball)를 받아 빌드하면 `.git` 이 없어 버전이 사라진다.

## `-ldflags -X` 를 안 쓴 이유

전통적인 방법은 링커로 변수를 덮어쓰는 것이다:

```
go build -ldflags "-X main.version=1.2.3"
```

두 가지가 걸린다. **빌드 명령을 쥔 사람만 값을 넣을 수 있다** — `go install github.com/…@v0.1.0` 로 받는 유저는 그 플래그를 안 준다. 그런데 실측하면 그렇게 설치한 바이너리에도 `v0.1.0` 이 이미 박혀 있다. 그리고 **버전이 두 군데가 된다** — 태그와 빌드 스크립트가 갈릴 수 있고, 갈려도 아무도 모른다.

## `go version -m` 이 같은 것을 보여 준다

코드 없이 아무 Go 바이너리나 열어 볼 수 있다:

```
$ go version -m cli-maker.exe
cli-maker.exe: go1.26.3
        path    github.com/Kwangseok-Seo/cli-maker
        mod     github.com/Kwangseok-Seo/cli-maker      v0.1.0
        dep     github.com/spf13/cobra                  v1.10.2
        build   vcs.revision=26ab395c90b0daa…
```

의존성과 그 버전까지 나오므로 남이 만든 바이너리의 공급망을 확인할 때도 쓴다. `--version` 을 구현하기 **전에** 무엇이 박히는지 재 볼 수 있는 것도 이 명령 덕분이다.

## cobra 쪽

`Version` 필드가 비어 있지 않을 때만 cobra 가 `--version` 을 단다 — 비어 있으면 flag 자체가 없고 아무 말도 안 한다 → [[cobra]]

```go
rootCmd.Version = buildVersion()
rootCmd.SetVersionTemplate("{{.Name}} {{.Version}}\n")
```

기본 템플릿은 `cli-maker version v0.1.0` 이라 `version` 과 `v` 가 겹친다. `-v` 축약은 cobra 가 자동으로 붙이는데, **이미 `v` shorthand 가 있으면 붙이지 않는다** — 나중에 `-v`(verbose)를 등록해도 조용히 뺏기지 않는다.

## 겪은 함정

- **`go run` 으로 검증하면 `(devel)` 이 나온다.** 구현이 틀린 게 아니라 확인 방법이 틀린 것이다.
- **`ok=false` 를 무시하고 싶어진다.** 실제로는 거의 안 오지만, 반환값이 둘인 API 는 [[maps]]·[[environment-variables]] 의 `ok` 와 같은 자리다 — 없는 것을 zero value 로 받아 넘기면 `--version` 이 빈 줄을 낸다.
- **버전이 v0.1.0 인데 도구는 그보다 앞서 있을 수 있다.** `--version` 은 *바이너리가 어느 소스에서 왔는지*를 말할 뿐, 그 소스가 최신인지는 말하지 않는다.

## 관련

[[go-modules]] · [[go-toolchain]] · [[cobra]] · [[user-config-dir]] · [[environment-variables]]
