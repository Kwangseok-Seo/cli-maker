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
- 하위 명령은 `rootCmd.AddCommand(...)` (다음 단계).

## 관련

[[structs]] · [[pointers]] · [[functions-as-values]] · [[go-modules]]
