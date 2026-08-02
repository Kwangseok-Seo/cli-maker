# module-publishing

Go 모듈을 세상에 내놓는 방법. 레지스트리에 업로드하는 절차가 없다 — **git 태그가 곧 발행**이다.

## 태그 이름이 버전이다

버전을 적는 파일이 없다. `git tag -a v0.2.0` 을 푸시하면 그것으로 끝이고, `v` 접두사와 semver 형식이 강제된다. 태그가 없으면 `@latest` 는 커밋을 가리키는 pseudo-version 으로 떨어진다 → [[build-info]]

annotated(`-a`)와 lightweight 둘 다 동작하지만, annotated 는 태거·시각·메시지를 담아 `git tag -n1` 로 읽힌다.

## 발행은 되돌릴 수 없다

푸시된 태그는 `proxy.golang.org` 가 그 시점 소스를 받아 **영구 캐시**한다. `git tag -d` 후 force push 해도 proxy 는 계속 서빙한다.

잘못 냈을 때의 정답은 *지우기*가 아니라 **다음 버전 내기**다. `go.mod` 의 `retract` 지시어로 *"이 버전은 고르지 마라"* 를 선언할 수 있지만, 이후 해석에만 영향을 주고 **이미 받아 간 사람에게는 소용없다.**

그래서 태그 전에 확인할 것이 하나 있다 — **태그된 소스가 자기 자신에 대해 참을 말하는가.** README 의 설치 안내가 "지금 `@latest` 는 v0.1.0" 이라고 적혀 있는데 그 커밋에 v0.2.0 을 달면, 그 거짓말이 영구히 박힌다. 문서를 먼저 고치고 그 커밋에 태그한다.

## "방금 태그했는데 왜 안 와요" — 캐시가 여러 층이다

같은 세션에서 잰 것:

```
git push --tags                                     태그 올라감

curl …/cli-maker/@latest      → v0.2.0             proxy 는 즉시 안다
curl …/cli-maker/@v/list      → v0.1.0             그런데 목록은 캐시가 남아 있다

go install …@latest           → v0.1.0             ← 안 온다
go install …@v0.2.0           → v0.2.0             ← 명시하면 온다

rm $GOMODCACHE/cache/download/…/@v/list
go install …@latest           → v0.2.0             ← 이제 온다
```

**proxy 와 로컬 모듈 캐시가 별개 층이고, `@latest` 해석은 버전 *목록*에 의존한다.** 그래서 태그 자체는 멀쩡한데 `@latest` 만 옛 버전에 묶이는 상태가 생긴다.

진단은 바깥에서 안으로:

1. 태그가 remote 에 있나 — `git ls-remote --tags origin`
2. proxy 가 아나 — `curl https://proxy.golang.org/<모듈>/@v/v0.2.0.info`
3. 로컬 캐시 — `$GOMODCACHE/cache/download/<모듈>/@v/list`

proxy URL 의 모듈 경로에서 **대문자는 `!소문자` 로 인코딩**된다 — `Kwangseok-Seo` → `!kwangseok-!seo`. 대소문자를 구분하지 않는 파일시스템에서 `github.com/Foo` 와 `github.com/foo` 가 같은 디렉토리로 뭉개지는 것을 막기 위한 규칙이다.

## major 2부터는 모듈 경로가 바뀐다

`v2` 이상은 모듈 경로 자체에 접미사가 붙는다 (Semantic Import Versioning):

```
go.mod    module github.com/Kwangseok-Seo/cli-maker/v2
import    "github.com/Kwangseok-Seo/cli-maker/v2/clirun"
```

**한 프로그램이 v1 과 v2 를 동시에 의존할 수 있게** 하려는 설계다(경로가 다르니 다른 패키지다). 대신 major 를 올리는 비용이 크다 — 이 repo 라면 `go.mod`·생성 템플릿·이미 생성돼 나간 CLI 들의 import 까지 전부 따라 바뀐다.

`v0.x` 는 semver 규격상 *"무엇이든 바뀔 수 있다"* 는 뜻이라 하위 호환 약속이 없다. 그 자유는 `v1.0.0` 을 내는 순간 사라진다 → [[go-modules]]

## 겪은 함정

- **태그가 문서 배지인 줄 알기 쉽다.** 이 repo 에서는 `generate` 산출물이 `go mod tidy` 로 proxy 를 보므로, 태그가 뒤처지면 생성된 CLI 가 **컴파일되지 않는다.** 발행 시점이 곧 계약이다 → [ADR-0014](../../adr/0014-publish-when-the-surface-grows.md)
- **repo 안에서는 이 격차가 안 보인다.** 테스트는 작업 트리의 소스를 보고 유저만 proxy 를 본다. 세 마일스톤 동안 아무 신호가 없었던 이유다.
- **`@latest` 가 안 바뀐다고 태그를 다시 달면 안 된다.** 위 진단 순서로 어느 층인지 먼저 가른다 — 대개 캐시고, 태그는 멀쩡하다.

## 관련

[[go-modules]] · [[build-info]] · [[go-toolchain]] · [[internal-packages]]
