# testing-the-filesystem

파일을 읽는 코드를 테스트하려면 **읽을 파일이 있어야** 한다. Go 는 길을 둘 준다 — 임시 디렉토리를 그 자리에서 만들거나, `testdata/` 에 미리 두거나.

## 핵심

```go
path := filepath.Join(t.TempDir(), "m.yaml")
if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
    t.Fatal(err)          // 준비 실패는 판정이 아니므로 Fatal 이다
}
m, err := Load(path)
```

`t.TempDir()` 은 호출마다 새 경로를 만들고 **지우는 코드를 우리가 쓰지 않는다** — 안에서 `t.Cleanup` 에 등록해 두고 테스트가 끝날 때 지운다. 서브테스트에서 부르면 그 서브테스트만의 디렉토리가 된다:

```
t.TempDir() = C:\Users\adman\AppData\Local\Temp\TestProbe3373533672\001
                                                └ 테스트 이름 + 난수      └ 호출마다 증가
```

## `testdata/` 와의 갈림

Go 툴체인은 **`testdata` 라는 이름의 디렉토리를 패키지로 보지 않는다.** 빌드·`vet`·`gofmt` 대상에서 빠지므로 깨진 소스나 이상한 바이트를 넣어도 안전하다.

| | `t.TempDir()` | `testdata/` |
|---|---|---|
| 입력이 있는 곳 | 케이스 옆 (테이블 안) | 별도 파일 |
| 케이스 간 격리 | 자동 (경로가 다르다) | 없음 (공유) |
| 적합 | 짧은 입력, 케이스마다 한 군데만 다른 것 | 크고 고정된 입력, 여러 테스트가 공유 |

cli-maker 는 `t.TempDir()` 을 골랐다. 매니페스트가 6줄이라 케이스 옆에 두는 편이 읽히고, [[table-driven-tests]] 가 세운 규약 — *케이스마다 딱 한 군데만 망가뜨리고 앞 케이스가 다음으로 새지 않게 한다* — 과 같은 이유다.

## 테스트가 실물 표면을 조립한다

`checkGlobal` 은 예약어를 상수로 두지 않고 `root` 에게 물어본다(ADR-0007). 그 대가가 테스트에 그대로 나타난다 — **테스트가 `root` 를 직접 만들어야 한다.**

```go
func newRoot() *cobra.Command {
    root := &cobra.Command{Use: "cli-maker"}
    root.PersistentFlags().String("output", "auto", "")
    root.AddCommand(&cobra.Command{Use: "greet"})
    return root
}
```

부담처럼 보이지만 이게 이득인 자리다. 상수 목록이었다면 테스트는 그 목록을 **다시 적었을** 것이고, 그건 검증이 아니라 복제다. 지금은 `newRoot` 에 `greet` 를 붙인 것만으로 `greet` 가 예약어가 되고, `--output` 을 단 것만으로 param `output` 이 거부된다.

`main.go` 의 루트를 빌려올 수는 없다 — `package main` 은 임포트 대상이 아니다([[packages-and-main]] · [[internal-packages]]). 그래서 테스트는 **판정에 실제로 필요한 표면만** 조립하고, 무엇이 필요한지가 거기서 드러난다.

## OS 가 쓴 문구는 단언하지 않는다

같은 실수(디렉토리를 파일로 읽기)가 플랫폼마다 다른 문장을 낸다.

```
Windows: read C:\…\ddd.yaml: Incorrect function.
Linux:   read /…/ddd.yaml: is a directory
```

이걸 `strings.Contains` 로 단언하면 테스트가 OS 에 묶인다. 대신 둘만 묻는다 — **우리가 붙인 조각**과 **센티널**:

```go
strings.Contains(err.Error(), "생략 (")        // "생략" 은 우리가 붙였다
errors.Is(err, fs.ErrNotExist)                // os.ReadFile 이 감싼 것을 뚫고 찾는다
```

같은 이유로 경로 문자열(`\` vs `/`)도 비교하지 않고 `filepath.Join` 으로 만든 값이나 파일명 조각만 본다([[error-wrapping]]).

## 붙기 전과 후를 비교한다

`LoadDir` 이 *새로* 붙인 것만 세려면 붙기 전을 알아야 한다. 값이 필요 없으니 map 을 집합으로 쓴다([[maps]]).

```go
before := map[string]bool{}
for _, c := range root.Commands() {
    before[c.Name()] = true
}
errs := LoadDir(root, dir)
for _, c := range root.Commands() {
    if !before[c.Name()] { got = append(got, c.Name()) }
}
```

`"greet"` 를 이름으로 걸러내지 않는 이유는, 그러면 `newRoot` 을 고칠 때 테스트도 함께 고쳐야 하기 때문이다.

**보고된 에러와 실제로 붙은 명령을 함께 본다.** 하나만 보면 두 실패가 구별되지 않는다 — 에러를 냈는데 등록도 해 버린 경우와, 조용히 아무것도 안 한 경우.

## 겪은 함정

- **`cobra.EnableCommandSorting` 기본값(true)이 순서를 지운다.** `main.go` 는 이걸 끄는데(ADR-0002) 테스트 바이너리는 `main` 을 안 타므로 기본값이 살아 있다. 안 끄면 `root.Commands()` 가 알파벳순으로 재배열해 *"`LoadDir` 이 파일명 순으로 붙였다"* 는 사실을 볼 수 없다. 확인법은 api 이름을 파일명과 **역순**(`aaa.yaml`→`zebra`, `bbb.yaml`→`alpha`)으로 두는 것 — 두 순서가 우연히 같으면 테스트가 아무것도 증명하지 않는다.
- **mutation 스크립트가 대상의 줄바꿈을 바꿨다.** 파이썬 `write_text` 가 텍스트 모드로 열어 `\n` → `\r\n` 으로 저장했고, LF 로 정규화하는 repo(`.gitattributes`)에서 `git status` 에 `M load.go` 가 남았다. 테스트는 전부 통과했으니 **테스트만 보고 있었다면 못 봤다** — 측정이 끝난 뒤 `git status` 를 본 것이 걸러 냈다.
- **`.yml` 이 조용히 사라지는 것을 이 테스트를 쓰다가 발견했다.** `LoadDir` 이 `filepath.Ext(…) != ".yaml"` 로 걸러서, `.yml` 로 저장한 매니페스트는 에러도 명령도 남기지 않았다. **테스트를 쓰는 일이 결함을 찾은 것**이지 테스트가 실패해서 찾은 게 아니다 — 케이스를 세려면 무엇이 무시되는지 적어야 했고, 적다 보니 그게 계약이 아님이 드러났다.

## mutation testing 으로 pinning 을 확인한다

테스트가 통과하는 것과 테스트가 **무언가를 막는 것**은 다르다. `load.go` 를 한 군데씩 지워(= mutant 를 심어) 보면 드러난다 — 테스트가 FAIL 하면 그 mutant 는 **killed**, 그대로 통과하면 **survived** 다.

| 지운 것 | 잡힌 케이스 |
|---|---|
| 확장자 검사 | 매니페스트가 아닌 파일은 무시한다 |
| `Validate` 실패 후 `continue` | 빈 파일은 Validate 가 잡는다 |
| `checkGlobal` 호출 | 이름 충돌 · 예약 명령 이름 · 예약 flag 이름 · `.yaml`↔`.yml` 충돌 **(4개)** |
| `ErrNotExist` 분기 | 디렉토리가 없으면 에러가 아니다 |

**첫 줄은 처음엔 비어 있었다.** `.yml` 을 받도록 고치기 전에는 "`.yml` 이면 무시된다" 케이스가 그 pinning 역할을 겸했는데, `.yml` 을 받게 되자 확장자 검사를 통째로 지워도 **아무 케이스도 깨지지 않았다** — 표에 매니페스트 아닌 파일이 하나도 없었기 때문이다. `zzz.txt` 케이스를 추가해서 다시 채웠다.

pinning 은 **우리가 걸어 둔 자리에만** 있다. 그리고 코드를 고치면 그 자리도 함께 움직인다 — mutation 을 한 번 재고 끝내면 그 사실이 안 보인다.

## 관련

[[go-test]] · [[table-driven-tests]] · [[file-io]] · [[maps]] · [[error-wrapping]] · [[cobra]] · [[internal-packages]] · [[packages-and-main]] · [[struct-tags]]
