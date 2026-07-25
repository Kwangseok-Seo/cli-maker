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

## 미사용 변수 규칙이 에러 검사를 강제한다 (M4)

```go
req, err := http.NewRequestWithContext(...)
resp, err := client.Do(req)     // ✗ declared and not used: err
```

첫 `err` 을 한 번도 *읽지* 않고 덮어썼다. 표면적으론 변수 잔소리지만 실제 의미는 **"에러를 검사하지 않았다"** 다. 예외가 없는 언어에서 에러는 그냥 반환값이라 무시할 수 있어 보이는데, 이 규칙이 그 구멍을 메운다 — 무심코 흘리면 컴파일 자체가 안 된다. (`_` 로 명시적으로 버리는 건 통과한다 → [[unused-results]].)

## 상태 코드는 우리가 정하는 정책 (M4)

```go
_, err = io.Copy(out, resp.Body)   // 본문을 먼저 내보내고
if err != nil { return err }
if resp.StatusCode >= 400 {        // 그다음에 상태를 판정한다
    return fmt.Errorf("HTTP %s", resp.Status)
}
```

- HTTP 404 는 통신으로는 **성공**이라 `Do` 의 `err` 은 `nil` 이다 → [[net-http]]. 실패로 볼지는 정책.
- 세 정책(exit 0 유지 / exit 1 / exit 1 + stderr 상태줄)을 실제 400 응답으로 대조한 뒤 셋째를 골랐다. 본문은 세 안 모두 stdout 이라 `| jq` 가 안 깨지고, 갈리는 건 종료 코드와 stderr 뿐이었다 → [[M4]].
- **본문 복사가 먼저**인 이유: 400 응답 본문에 `couldn't convert 'abc' to type Long` 같은 진짜 이유가 들어 있다. 상태부터 보고 일찍 반환하면 유저가 그 설명을 잃는다.

## 겪은 함정

- M3: 에러 분기 안에서 `rootCmd.Execute()` 를 부르고 `return` 을 빼먹어 **명령이 두 번 실행**됐다. `if` 블록이 끝나면 다음 문장으로 흘러내리고, 아래에 또 `Execute()` 가 있었기 때문. `os.Exit` 이 있는 분기만 빠져나간다는 감각이 착각의 원인.
- M4: `fmt.Errorf(...)` 를 `return` 없이 써서 에러를 만들고 즉시 버렸다. 400 이 나도 exit 0 이었을 코드 — `go vet` 이 잡았다 → [[unused-results]].

## 관련

[[cobra]] · [[packages-and-main]] · [[go-toolchain]] · [[stdout-stderr]] · [[file-io]] · [[net-http]] · [[unused-results]] · [[defer]]
