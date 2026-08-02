# github-actions

이벤트가 오면 깨끗한 VM 을 띄우고, repo 를 체크아웃하고, 적어 둔 명령을 돌리고, VM 을 버린다. **매번 새 머신**이라 상태가 남지 않는다.

## 네 층과 격리 규칙

```
Event                    push · pull_request · schedule · workflow_dispatch …
└── Workflow             .github/workflows/*.yml 파일 하나
    └── Job              ← 하나의 job = 하나의 VM
        └── Step         ← run: 셸 명령, 또는 uses: 남의 action
```

| 경계 | 공유되는 것 | 공유 **안** 되는 것 |
|---|---|---|
| step ↔ step (같은 job) | 파일시스템, 작업 디렉토리 | **셸 프로세스** — `export` 가 안 넘어간다 (`$GITHUB_ENV` 로) |
| job ↔ job | 없음 | **파일시스템 전체** — 다른 머신이다 (artifact 로 넘긴다) |

job 은 기본 병렬이고 `needs:` 로 순서를 건다. cli-maker 의 릴리스가 **한 job** 인 이유는 Go 크로스 컴파일이 한 머신에서 전 플랫폼을 내기 때문이다 — matrix 로 쪼개면 artifact 로 다시 모으는 단계가 붙는다.

## 권한 — 기본값은 read 다

Actions 는 job 마다 `GITHUB_TOKEN` 을 자동 발급한다(secret 을 만들 필요 없고 job 이 끝나면 만료). 그런데 **권한은 repo 설정이 정한다.**

```
$ gh api repos/<owner>/<repo>/actions/permissions/workflow
{"default_workflow_permissions":"read", …}
```

이 상태로 `gh release create` 를 하면 403 이다. workflow 에 명시하는 쪽이 repo 설정을 넓히는 것보다 좁다:

```yaml
permissions:
  contents: write     # 적지 않은 나머지(issues·packages·actions …)는 전부 none 이 된다
```

## `uses:` 는 남의 코드를 내 VM 에서 돌리는 것

`actions/checkout@v7` 은 그 repo 의 코드를 내 러너에서 실행한다 — **내 토큰이 있는 환경에서**. `@v7` 은 태그이고 태그는 옮길 수 있으므로, 저자(또는 저자 계정을 쥔 사람)가 태그를 옮기면 다음 실행부터 다른 코드가 돈다. 좁히려면 40자리 SHA 로 고정한다.

`gh api repos/actions/checkout/releases/latest -q .tag_name` 로 최신 major 를 확인할 수 있다. 뒤처지면 *"Node.js 20 is deprecated"* 같은 annotation 이 붙는다.

## `${{ }}` 는 실행 **전에** 치환된다

셸 변수가 아니다. Actions 가 스크립트를 만들기 전에 문자열을 그 자리에 박아 넣으므로, 커밋 메시지·이슈 제목처럼 **남이 쓰는 값**을 `run:` 에 직접 넣으면 그 안의 `` ` `` 나 `$(...)` 가 내 러너에서 실행된다.

```yaml
- env:
    MSG: ${{ github.event.head_commit.message }}
  run: echo "$MSG"        # 셸 변수로 받으면 확장되지 않는다
```

## workflow 파일은 별도 scope 으로 잠겨 있다

`repo` 권한만 있는 토큰으로는 `.github/workflows/` 를 밀 수 없다.

```
! [remote rejected] main -> main
  (refusing to allow an OAuth App to create or update workflow ... without `workflow` scope)
```

workflow 는 CI 에서 임의 코드를 실행하므로, repo 쓰기만으로 workflow 를 심을 수 있다면 **토큰 탈취가 곧 코드 실행**이 된다. `gh auth refresh -h github.com -s workflow` 로 더한다(브라우저 인증이 필요하다).

## 실측 — 얕은 체크아웃은 태그를 안 가져온다

첫 실행이 **통과하면서** 이 버전을 냈다:

```
cli-maker v0.0.0-20260802123306-0d076ac27677     ← CI
cli-maker v0.2.1-0.20260802103502-e49d710adc62   ← 같은 커밋, 로컬
```

`actions/checkout` 의 기본은 얕은 체크아웃이라 태그가 오지 않는다. Go 는 태그가 하나도 없는 repo 로 보고 **`v0.0.0` 부터 센다** → [[build-info]]

게이트가 `(devel)` 만 보고 있어서 이걸 통과시켰다. 두 실패는 다르다:

| 증상 | 뜻 |
|---|---|
| `(devel)` | vcs 스탬핑 자체가 없다 (`.git` 없음, worktree 등) → [[cross-compilation]] |
| `v0.0.0-…` | 스탬핑은 됐는데 **태그를 못 봤다** (fetch-depth) |

`fetch-depth: 0` 으로 전체 히스토리를 가져오면 해결된다. 버전 계산에는 태그 객체뿐 아니라 **거기까지의 거리**가 필요하다.

## 그래서 workflow 가 자기 산출물을 검사한다

CI 는 성공했다고 말하면서 틀린 것을 낼 수 있다. 릴리스에 붙기 **전에** 확인한다:

```yaml
- run: |
    V=$(./stage/linux_amd64/cli-maker --version)
    case "$V" in
      *'(devel)'*)   echo "::error::스탬핑 안 됨"; exit 1 ;;
      *' v0.0.0-'*)  echo "::error::태그를 못 봄"; exit 1 ;;
    esac
- run: |
    mode=$(tar -tzvf dist/*_linux_amd64.tar.gz | awk '$NF=="cli-maker"{print $1}')
    case "$mode" in -rwx*) ;; *) exit 1 ;; esac
```

`::error::` 는 Actions 가 알아듣는 workflow command 로, 로그가 아니라 실행 요약에 뜬다.

## 디버깅은 느리다

**로컬에서 못 돌린다.** push 해야 돈다. 그래서 `workflow_dispatch` 를 트리거에 같이 달아 태그를 쓰지 않고 시험한다 — cli-maker 는 이때 빌드·검증·artifact 까지만 하고 릴리스는 건너뛴다:

```yaml
- name: 릴리스에 붙인다
  if: github.ref_type == 'tag'
```

건너뛴 step 은 실행 목록에 `-` 로 표시된다.

## 겪은 함정

- **첫 실행이 성공했는데 산출물이 틀렸다.** 초록 체크는 "내가 검사한 것이 통과했다"는 뜻이지 "결과가 옳다"가 아니다.
- **`chmod`·`tar` 가 OS 마다 다르게 군다.** Windows 에서 만든 tar 는 실행 비트가 없고 `--force-local` 이 필요했는데, 리눅스 러너에서는 둘 다 문제가 아니다 → [[cross-compilation]]
- **버전을 두 군데 적기 쉽다.** `go-version-file: go.mod` 로 물어보면 go.mod 를 올릴 때 workflow 가 뒤처지지 않는다.

## 관련

[[cross-compilation]] · [[build-info]] · [[module-publishing]] · [[go-toolchain]]
