# error-handling

Go 는 예외(exception)를 던지지 않는다. 함수가 **에러를 값으로 반환**하고, 호출한 쪽이 `if err != nil` 로 확인한다. 이 관용구가 Go 코드 전반을 지배한다.

## 핵심

```go
if err := rootCmd.Execute(); err != nil {
    os.Exit(1)
}
```

- `if err := f(); err != nil { ... }` — f 실행 → 에러를 `err` 에 받고 즉시 검사. `err` 은 이 `if` 블록 범위에서만 산다.
- 에러는 보통 마지막 반환값: `func f() (T, error)`.
- `os.Exit(1)` — 0 아닌 코드로 종료 = 실패 신호. 종료 코드 0 = 성공.

## 왜 중요한가

- 종료 코드로 스크립트·CI·파이프라인이 성공/실패를 판단한다 (`cmd && next`, `set -e`, `$LASTEXITCODE`).
- 에러를 무시하면 실패해도 0으로 끝나 "성공"으로 오해된다. (M1 의 `greet` 인자 누락이 그 예 — [[cobra]] 가 에러는 출력하지만, 종료 코드는 우리가 챙겨야 했다.)

## 정책은 호출한 쪽이 정한다 (M3)

```go
m, err := manifest.Load(path)
if err != nil {
    fmt.Fprintf(os.Stderr, "cli-maker: %s 생략 (%v)\n", path, err)
    continue // 이 API 만 버리고 나머지는 살린다 (격벽)
}
```

- 하위 패키지(`manifest`, `cli`)는 에러를 **반환만** 한다. 경고를 찍을지 죽을지는 `main` 이 정한다 — 라이브러리가 정책을 결정하면 테스트나 재사용에서 원치 않는 출력이 샌다.
- 정책은 깨진 YAML 로 실물 비교해 골랐다(건너뛰기 / fail fast / 깨진 것도 노출) → [[M3]].
- 건너뛰기의 생명은 **경고 변별력**. `Fprintf` 는 개행을 붙이지 않으니 `\n` 을 빼면 경고들이 한 줄에 붙는다 → [[stdout-stderr]].

## 포장된 에러: errors.Is

```go
if err != nil && !errors.Is(err, os.ErrNotExist) { ... }
```

- `os.ReadDir` 의 에러는 `*os.PathError` 로 **감싸여** 있어 `err == os.ErrNotExist` 값 비교가 실패한다. `errors.Is` 는 그 포장을 풀며 따라 들어간다.
- 덕분에 "디렉토리 없음"(정상 — 동적 명령 0개)과 "권한 없음"(알릴 일)을 구분할 수 있다.

## 겪은 함정

- M3: 에러 분기 안에서 `rootCmd.Execute()` 를 부르고 `return` 을 빼먹어 **명령이 두 번 실행**됐다. `if` 블록이 끝나면 다음 문장으로 흘러내리고, 아래에 또 `Execute()` 가 있었기 때문. `os.Exit` 이 있는 분기만 빠져나간다는 감각이 착각의 원인.

## 관련

[[cobra]] · [[packages-and-main]] · [[go-toolchain]] · [[stdout-stderr]] · [[file-io]]
