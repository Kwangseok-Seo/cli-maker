# user-config-dir

설치된 프로그램이 "유저의 것"을 어디에 두는지의 OS 규약과, Go 가 그것을 감싸는 방법.

## 함수 하나가 네 전통을 감싼다

`os.UserConfigDir()` 의 실제 분기 (`$GOROOT/src/os/file.go`):

| GOOS | 보는 환경변수 | 뒤에 붙는 것 |
|---|---|---|
| windows | `%AppData%` | 없음 |
| darwin, ios | `$HOME` | `/Library/Application Support` |
| unix | `$XDG_CONFIG_HOME`, 없으면 `$HOME` | 후자면 `/.config` |
| plan9 | `$home` | `/lib` |

XDG Base Directory Spec(`~/.config`, `~/.local/share`, `~/.cache`)은 **Linux 관습**이지 범용 표준이 아니다. Windows·macOS 는 각자 규약을 갖고 있고, Windows 머신에서 `$XDG_CONFIG_HOME` 은 아예 unset 이다.

`os.UserCacheDir()` 이 같은 구조로 캐시 자리를 준다. Windows 에서 둘이 갈린다 — 설정은 `%AppData%`(Roaming, 도메인 계정을 따라 이동), 캐시는 `%LocalAppData%`(이 PC 전용).

```
os.UserConfigDir()   C:\Users\adman\AppData\Roaming
os.UserCacheDir()    C:\Users\adman\AppData\Local
```

## 세 가지를 안 해 준다

- **디렉토리를 만들지 않는다.** 존재 여부도 확인하지 않고 경로 문자열만 계산한다.
- **앱 이름을 붙여 주지 않는다.** `AppData\Roaming` 까지만 준다. `filepath.Join(cfg, "cli-maker", "apis")` 는 우리 몫이다.
- **실패할 수 있다.** 환경변수가 비면 경로가 아니라 error 다(`%AppData% is not defined`). `%AppData%` 없는 서비스 계정, `$HOME` 없는 컨테이너가 실제로 그렇다.

셋째가 설계에 걸린다. 설정 디렉토리를 못 알아내는 것은 **실패가 아니라 그 자리가 없는 것**이다 — 에러로 올리면 그 기능을 안 쓰는 유저까지 도구가 안 뜬다. cli-maker 는 그 자리를 목록에서 빼고 진행한다.

```go
dirs := []string{"apis"}
if cfg, err := os.UserConfigDir(); err == nil {
    dirs = append(dirs, filepath.Join(cfg, "cli-maker", "apis"))
}
```

## 그래서 이 함수는 환경으로 조종할 수 없다

테스트에서 설정 디렉토리를 가짜로 만들려면 `t.Setenv` 로 환경변수를 바꿔야 하는데, **어느 변수를 바꿔야 하는지가 GOOS 마다 다르다.** Windows 에서 `APPDATA` 를 바꾼 테스트는 Linux CI 에서 아무 효과가 없다. darwin 은 더 나쁘다 — `$HOME` 을 바꿔도 뒤에 `/Library/Application Support` 가 강제로 붙어 **원하는 경로를 정확히 만들 수 없다.**

처방은 층을 가르는 것이다.

```
환경을 읽는 층   main.apiDirs      os.UserConfigDir 을 부른다   → 테스트 대상 아님
목록을 쓰는 층   cli.LoadDirs      []string 을 받아 순회한다     → 여기를 테스트한다
```

검증하고 싶은 계약(*"앞선 디렉토리가 이기고 가려진 쪽은 이유가 남는다"*)은 경로 두 개만 있으면 `t.TempDir` 로 확인된다. `os.UserConfigDir` 이 필요 없다 → [[testing-the-filesystem]].

## os.Executable() 은 다른 축이다

설정 디렉토리가 "유저의 것"이라면 실행 파일 옆은 "설치의 것"이다. 아카이브를 풀어 쓰는 배포에서 바이너리와 데이터가 함께 다니는 형태.

**`go run` 에서는 임시 빌드 산출물을 가리킨다 — 실측:**

```
go run 으로     C:\...\Temp\go-build2193046101\b001\exe\dirs.exe
go build 후     C:\...\scratchpad\dirs\dirs.exe
```

`go run` 은 매번 임시 디렉토리에 빌드해서 실행하므로([[go-toolchain]]), 개발 중 `go run .` 으로 확인하면 이 경로는 영원히 쓸모없는 값이다. symlink 로 설치된 경우엔 `filepath.EvalSymlinks` 가 한 겹 더 필요하다.

## 겪은 함정

- **XDG 를 "표준"으로 알고 접근하면 절반만 맞는다.** 세 OS 중 하나의 관습이다. 직접 `$XDG_CONFIG_HOME` 을 읽는 코드는 Windows 에서 조용히 빈 문자열을 받는다.
- **`os.UserConfigDir()` 이 준 경로를 그대로 쓰면 남의 자리에 쓴다.** `AppData\Roaming` 까지만 오므로 앱 이름을 안 붙이면 다른 프로그램의 설정과 같은 디렉토리를 뒤진다.
- **에러를 위로 올리고 싶어진다.** `if err != nil { return nil, err }` 이 반사적으로 나오는데, 여기서는 그게 "설정 디렉토리 없는 환경 = 도구 못 씀"이 된다. [[error-handling]] 의 기본형이 안 맞는 자리다.

## 관련

[[environment-variables]] · [[config-precedence]] · [[testing-the-filesystem]] · [[go-toolchain]] · [[error-handling]]
