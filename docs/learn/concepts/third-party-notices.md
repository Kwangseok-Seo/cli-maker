# third-party-notices

내 프로그램에 남의 코드가 들어갔을 때, 그 저작권자와 라이선스를 배포물에 함께 알리는 것.

## 왜 — 오픈소스는 공짜지만 무조건은 아니다

MIT·BSD·Apache 같은 permissive 라이선스는 대부분 같은 조건을 건다.

> *"위 저작권 고지와 이 허가 고지를 소프트웨어의 모든 사본 또는 상당 부분에 포함해야 한다"*

**"써도 좋다, 다만 누가 만들었는지는 밝혀라"** 다. 안 지키면 조건 위반이고, 형식적으로는 그 코드를 쓸 권리가 없는 상태가 된다.

## 언제 켜지나 — 배포 형태가 정한다

| 배포 형태 | 남의 코드가 어디 있나 | 고지 의무 |
|---|---|---|
| 소스 공개 (GitHub repo) | 없음 — `go.mod` 에 이름만 | **없음** |
| `go install` | 없음 — 유저의 Go 가 각 모듈을 각자의 LICENSE 와 함께 받는다 | **없음** |
| **바이너리 배포** | **정적 링크되어 실행 파일 안에** | **있음** |

Go 는 정적 링크라 `cli-maker.exe` 하나에 cobra·pflag·yaml 의 컴파일된 코드가 들어 있다. 그 파일을 남에게 주는 순간 "남의 저작물이 담긴 사본을 배포"하는 것이 된다.

cli-maker 는 M11 까지 이 문제가 없다가 릴리스에 바이너리를 올린 순간 생겼다. **의무는 코드가 아니라 배포 방식에서 온다.**

## 라이선스별로 요구가 조금씩 다르다

| 라이선스 | 요구 |
|---|---|
| MIT | 저작권 고지 + 허가 고지 전문 |
| BSD-3-Clause | 위 + 이름으로 보증·추천하지 말 것 |
| **Apache-2.0** | 위 + **원저작물에 `NOTICE` 가 있으면 그 내용도 전달** (§4(d)) + 변경 사실 표시 |

Apache 의 NOTICE 조항이 실무에서 잘 빠진다. `gopkg.in/yaml.v3` 가 그 예다 — LICENSE 첫 줄이 *"This project is covered by two different licenses: MIT and Apache"* 이고, libyaml 을 Go 로 옮겨 온 부분이 Apache-2.0(Canonical)이라 `NOTICE` 파일이 따로 있다.

## 목록을 어디서 얻나 — `-deps` 와 `-m all` 은 다르다

```
go list -m all        go.mod 이 아는 전부 (테스트·문서 생성용 포함)
go list -deps .       실제로 링크되는 것만
```

cli-maker 에서 `blackfriday`·`check.v1` 은 앞쪽에만 나온다 — 바이너리에 안 들어가므로 고지 대상이 아니다. **고지는 링크되는 것에 대한 의무**다.

모듈 디렉토리까지 물어보면 라이선스 파일을 그 자리에서 읽을 수 있다:

```
go list -deps -f '{{if not .Standard}}{{.Module.Path}}|{{.Module.Version}}|{{.Module.Dir}}{{end}}' .
```

이름을 손으로 적지 않는 것이 요점이다 → [[internal-packages]] 와 같은 결로, **실물에 물어보면 뒤처지지 않는다.**

## 함정 — 의존성 집합은 GOOS 마다 다르다

```
GOOS=linux    cobra · pflag · yaml.v3
GOOS=windows  cobra · pflag · yaml.v3 · mousetrap
```

cobra 는 `mousetrap` 을 windows 에서만 쓴다(콘솔에서 더블클릭 실행을 감지). `go list -deps` 는 **현재 GOOS 기준으로** 답하므로, 리눅스 CI 에서 한 번 만들어 전 플랫폼 아카이브에 복사하면 windows 쪽만 조용히 틀린다 → [[cross-compilation]]

그래서 고지는 크로스 컴파일 루프 **안에서** 플랫폼마다 만든다.

## `go-licenses` 를 안 쓴 이유

표준적인 도구이고 라이선스 종류를 판정해 준다:

```
$ go-licenses csv .
github.com/spf13/cobra,…,Apache-2.0
gopkg.in/yaml.v3,…,MIT          ← 틀렸다: MIT + Apache-2.0 이고 NOTICE 가 있다
```

**판정하는 도구는 오판할 수 있다.** 파일을 그대로 싣는 쪽은 판정을 하지 않으므로 그 실수를 할 수 없다 — 파일이 있으면 실린다. 27MB 바이너리와 그 자신의 낡은 의존성을 CI 에 들이지 않는 것은 덤이다.

일반화하면: **분류가 목적이면 도구, 전달이 목적이면 복사.** 고지의 목적은 전달이다.

## 겪은 함정

- **문서에 적어 두고 이행하지 않았다.** README 두 곳에 *"바이너리를 배포할 때는 고지가 필요합니다"* 라고 써 두고, 정작 v0.2.0 바이너리를 고지 없이 올렸다. 아는 것과 하는 것은 다르다.
- **누락을 막으려던 단계가 누락을 만들었다.** 한 번 만들어 복사하는 첫 판이 windows 고지를 빠뜨렸다. artifact 를 열어 3개 대 4개를 대조하고서야 보였다.
- **존재 확인만으로는 못 잡는다.** "windows 고지에 mousetrap 이 있나"만 보면 한 파일을 전부에 복사해도 통과한다. **linux 에는 없어야 한다**까지 확인해야 플랫폼별 생성이 실제로 일어났다는 판정이 된다.

## 관련

[[cross-compilation]] · [[go-modules]] · [[github-actions]] · [[module-publishing]]
