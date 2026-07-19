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

## 관련

[[structs]] · [[pointers]] · [[functions-as-values]] · [[go-modules]]
