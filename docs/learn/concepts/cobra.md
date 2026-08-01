# cobra

CLI 프레임워크(`github.com/spf13/cobra`). 명령을 struct 로 **묘사(declare)** 하면 파싱·도움말·서브커맨드를 대신 처리한다. kubectl·gh 가 사용.

## 핵심

```go
rootCmd := &cobra.Command{
    Use:   "cli-maker",
    Short: "...",
    Run:   func(cmd *cobra.Command, args []string) { ... },
}
rootCmd.Execute()  // os.Args 를 읽어 맞는 Run 호출
```

- `Use`/`Short` 묘사만으로 `--help`·`-h` **자동 생성**.
- `Run` 의 `args []string` = 프로그램명·플래그를 걸러낸 깔끔한 인자 → [[slices-and-args]] 의 손파싱 고통 해소.
- 하위 명령 = 또 하나의 `cobra.Command` + `AddCommand` (아래).

## 하위 명령 (subcommand)

하위 명령은 그저 또 하나의 `cobra.Command`. `AddCommand` 로 루트에 매단다.

```go
greetCmd := &cobra.Command{
    Use:  "greet <이름>",
    Args: cobra.ExactArgs(1), // 인자 개수 검증
    Run: func(cmd *cobra.Command, args []string) {
        fmt.Println("안녕, " + args[0] + "!") // 검증 후라 안전
    },
}
rootCmd.AddCommand(greetCmd)
```

- `Args: cobra.ExactArgs(1)` → 인자 부족 시 **panic 대신** `Error: accepts 1 arg(s), received 0` + 사용법. [[slices-and-args]] 의 panic 을 대체.
- 루트 `--help` 에 자동 등재. `help`·`completion` 하위 명령도 cobra 가 공짜로 추가.

## 명령 트리와 그룹 명령 (M3)

```go
group := &cobra.Command{Use: m.Name} // Run 없음 = 스스로 하는 일이 없는 그룹
group.AddCommand(sub)                // 부모-자식을 잇는 유일한 수단
rootCmd.AddCommand(group)            // cli-maker <api> <command>
```

- cobra 는 명령줄 단어를 왼쪽부터 소비하며 트리를 내려가 **도착한 노드의 `Run`** 을 부른다. `Run` 이 없는 노드에서 멈추면 자식 목록(`--help`)을 보여준다.
- 자식 개수는 매니페스트가 정하니 노드를 루프가 만든다 → [[closures]].

## 자동 정렬 끄기

```go
cobra.EnableCommandSorting = false // 패키지 전역 — 명령 만들기 전에 한 번
```

- 기본값 `true` 면 자식 명령이 **알파벳순**으로 재배열돼 매니페스트에 적은 순서가 사라진다(ADR-0002 의 근거와 충돌).
- 전역 변수라 그룹만 따로 끌 수 없다 — root 목록도 등록 순이 된다(우리 명령이 앞, `help`/`completion` 이 뒤).

## Run vs RunE (M4)

```go
Run:  func(cmd *cobra.Command, args []string)          // 반환값 없음
RunE: func(cmd *cobra.Command, args []string) error    // 에러를 올릴 수 있다
```

`RunE` 가 반환한 에러는 **`rootCmd.Execute()` 의 반환값으로 그대로 올라온다.** cobra 는 그것을 `Error: …` 로 stderr 에 찍어 주지만 **종료 코드는 건드리지 않는다** — 0 아닌 종료는 `main` 이 챙긴다:

```go
if err := rootCmd.Execute(); err != nil { os.Exit(1) }
```

그래서 `RunE` → `Execute()` → `os.Exit(1)` 이 하나의 통로다. cli-maker 는 이 통로 하나로 "본문은 stdout, 진단은 stderr, 실패는 exit 1"을 전부 처리한다 — `os.Stderr` 를 직접 건드리지 않아 [[io-reader-writer]] 로 열어 둔 테스트 가능성이 유지된다 → [[error-handling]].

## SilenceUsage 는 RunE 안에서 켠다 (M4)

에러를 반환하면 cobra 는 기본적으로 **사용법 전체를 덤프**한다. `| head` 로 파이프가 끊겼을 때도 도움말이 쏟아지는 건 소음이다. 그렇다고 명령에 `SilenceUsage = true` 를 박으면 정작 필요한 곳(인자를 틀렸을 때)까지 죽는다. 실측:

| 설정 위치 | flag 누락 | 런타임 에러 |
|---|---|---|
| 없음(기본) | usage 나옴 ✓ | usage 덤프 ✗ |
| `cmd.SilenceUsage = true` (명령에) | usage 사라짐 ✗ | usage 없음 ✓ |
| **`RunE` 첫 줄에서** | usage 나옴 ✓ | usage 없음 ✓ |

flag 파싱은 `RunE` **이전에** 끝나므로, `RunE` 안에서 켜면 그 경로에는 영향이 없다. (`SilenceErrors` 는 별개 — 그건 `Error:` 줄 자체를 지운다.)

## 플래그 (M4)

```go
sub.Flags().String(p.Name, "", p.In+" - "+p.Type)  // 등록 (시작 시점)
sub.MarkFlagRequired(p.Name)
...
cmd.Flags().GetString(p.Name)                       // 읽기 (Run 시점)
```

- `Flags()` 는 그 명령 전용, `PersistentFlags()` 는 자식까지 상속.
- **등록과 읽기는 시점이 다르다.** 등록은 명령 트리를 세울 때, 읽기는 cobra 가 argv 를 파싱한 뒤. 그래서 `c.Params` 를 두 번 돈다.
- 등록은 `&cobra.Command{...}` 리터럴 **밖**에서 — 리터럴은 표현식이라 `for` 를 품을 수 없다 → [[composite-literals]].
- `MarkFlagRequired` 는 파싱 때 막아 줄 뿐 **`--help` 에는 표시되지 않는다.** 필수임을 보이려면 도움말 문자열에 직접 적어야 한다.

## Changed — "안 줬다"와 "빈 값을 줬다" (M10)

```go
cmd.Flags().Changed("data")     // 유저가 이 flag 를 명령줄에 적었는가
cmd.Flags().GetString("data")   // 그 값 (안 적었으면 등록 시 기본값)
```

값만 읽으면 `--data ""` 와 `--data` 를 아예 안 준 것이 같아진다. pflag 는 "설정됐는지"를 값과 **별도로** 기록해 두고 `Changed` 로 내준다 → [[absent-vs-empty]].

## 같은 명령에 같은 이름 flag = 패닉 (M10)

```
panic: withBody flag redefined: data
```

`PersistentFlags()` 로 그룹에 단 flag 는 같은 이름의 자식 flag 가 **가리기만** 한다. 그런데 같은 `Flags()` 에 같은 이름을 두 번 등록하면 pflag 가 패닉한다 — 무관한 명령까지 함께 죽는다(M6 의 중복 param 과 같은 부류).

그래서 cli-maker 는 등록 시점에 `Flags().Lookup(name) != nil` 로 갈라 패닉을 막고, 그 사실을 검증기가 따로 보고한다. **가드만 있으면 유저가 적은 param 이 조용히 사라지고, 검사만 있으면 보고하기 전에 패닉한다** — 둘 다 있어야 "그 파일만 빠지고 이유를 말한다"가 된다 → [[M10]].

## 관련

[[structs]] · [[pointers]] · [[functions-as-values]] · [[go-modules]] · [[closures]] · [[composite-literals]] · [[context]] · [[error-handling]] · [[absent-vs-empty]]
