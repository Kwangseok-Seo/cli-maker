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

## Run vs RunE

- `RunE` 는 에러를 **반환**하면 cobra 가 `Error:` 접두 + 사용법 + 0 아닌 종료까지 처리한다. `Run` 은 반환할 수 없어 `os.Exit` 을 우리가 챙겨야 한다 → [[error-handling]].

## 관련

[[structs]] · [[pointers]] · [[functions-as-values]] · [[go-modules]] · [[closures]] · [[composite-literals]]
