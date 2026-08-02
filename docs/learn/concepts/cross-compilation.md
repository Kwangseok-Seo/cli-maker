# cross-compilation

한 대의 머신에서 다른 OS·CPU 용 바이너리를 만드는 것. Go 에서는 환경변수 두 개면 된다.

## GOOS 와 GOARCH

```
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o cli-maker .
```

Windows 한 대에서 linux·darwin 바이너리 5종이 전부 나온다 — 크로스 컴파일러를 따로 설치하지 않는다. 툴체인이 모든 대상의 표준 라이브러리를 이미 갖고 있기 때문이다.

```
$ go tool dist list | wc -l
47
```

47개가 지원되지만 실제로 내는 것은 보통 다섯이다 — `linux/amd64`, `linux/arm64`, `darwin/amd64`(Intel Mac), `darwin/arm64`(Apple Silicon), `windows/amd64`.

## CGO_ENABLED=0

C 코드를 섞지 않겠다는 뜻이고, 그 결과가 **정적 링크**다. 순수 Go 프로그램이라도 명시하는 이유는 기본값이 플랫폼마다 다르기 때문이다 — Linux 에서 빌드하면 기본이 1 이라 glibc 에 동적 링크될 수 있고, 그 바이너리는 더 오래된 배포판에서 안 뜬다. 크로스 컴파일 시에는 C 툴체인이 없어 자동으로 0 이 되지만, **CI 가 어디서 도는지에 따라 갈리므로** 명시가 안전하다.

## 빌드 플래그 — 실측

| 플래그 | 크기 | `Users/adman` 문자열 | `--version` |
|---|---|---|---|
| (없음) | 13,644,288 | **1회** | 산다 |
| `-trimpath` | 13,620,224 | 0회 | 산다 |
| `-trimpath -ldflags "-s -w"` | **9,641,984** | 0회 | **산다** |

- **`-trimpath`** 의 목적은 크기가 아니라 **빌드 머신 경로 제거**다. 안 주면 내 홈 디렉토리 경로가 배포 바이너리 안에 그대로 들어간다(프라이버시). 다른 머신에서 빌드해도 같은 결과가 나오게 하는 재현성 조건이기도 하다.
- **`-s -w`** 는 심볼 테이블과 DWARF 를 뺀다 — 29% 감소. 스택 트레이스의 함수명은 남고 디버거 정보만 사라진다.
- **build info 는 별개 섹션이라 `-s -w` 로 지워지지 않는다.** `--version` 이 릴리스 빌드에서도 그대로 동작한다 → [[build-info]]

## 함정 — `git worktree` 에서는 버전이 안 박힌다

태그된 소스에서 빌드하려고 worktree 를 썼더니 `(devel)` 이 나왔다.

```
$ git worktree add ./wt v0.2.0
$ cd wt && go build -o probe.exe . && ./probe.exe --version
cli-maker (devel) (go1.26.3 windows/amd64)      ← vcs= 줄이 아예 없다

$ ls -la wt/.git
-rw-r--r--  60  wt/.git                          ← 디렉토리가 아니라 파일
$ cat wt/.git
gitdir: C:/Users/adman/projects/cli-maker/.git/worktrees/wt
```

**Go 의 VCS 스탬핑은 `.git` 디렉토리를 찾고, worktree 의 `gitdir:` 포인터 파일은 따라가지 않는다.** git 명령은 멀쩡히 동작하므로(`git describe` → `v0.2.0`) 소스는 정상인데 버전만 사라진다.

정답은 **clone** 이다 — CI 가 하는 방식이기도 하다.

```
$ git clone -b v0.2.0 <repo> rel-src
$ cd rel-src && go build -o probe.exe . && ./probe.exe --version
cli-maker v0.2.0 (go1.26.3 windows/amd64)       ← .git 이 디렉토리라 스탬핑된다
```

세 소스가 세 답을 준다: 작업 트리(HEAD)는 pseudo-version, worktree 는 `(devel)`, clone 은 태그 그대로.

## 아카이브 — Windows 에서 만들 때의 두 함정

**실행 비트가 안 붙는다.** 첫 tar 의 바이너리가 `-rw-r--r--` 였다. Git Bash 의 `chmod +x` 는 NTFS 에서 먹지 않아 `tar --mode=0755` 로 강제해야 했다. 아카이브 만들기 자체는 성공한 것처럼 보이므로 **열어 보지 않으면 모른다** — 받은 유저가 `Permission denied` 를 보고 나서야 드러난다.

**GNU tar 가 `C:` 를 원격 호스트로 읽는다.**

```
tar: Cannot connect to C: resolve failed
```

콜론이 든 경로를 `host:path` 로 해석하는 옛 문법 때문이다. `--force-local` 이 필요하다.

둘 다 **Linux CI 에서 만들면 사라지는 문제**다. 로컬 릴리스가 번거로운 만큼 자동화의 근거가 된다.

## 관습

- Unix 는 `.tar.gz`, Windows 는 `.zip`
- 아카이브 안은 flat — 유저가 풀어 바로 PATH 에 옮긴다
- 바이너리와 함께 `LICENSE`·`README.md`
- 파일명에 버전과 플랫폼: `cli-maker_v0.2.0_linux_amd64.tar.gz`
- `checksums.txt` 하나에 SHA256 을 모아 둔다 (`sha256sum -c checksums.txt`)

## 관련

[[build-info]] · [[module-publishing]] · [[go-toolchain]] · [[bitmasks]]
