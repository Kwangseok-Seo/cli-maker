# interfaces

"무엇을 할 수 있는가"만 적은 약속. 구현하는 쪽이 선언하지 않아도 메서드만 맞으면 자동으로 성립한다.

## 핵심

```go
type Formatter interface {
	Format(dst io.Writer, src io.Reader) error
}
```

- 메서드 목록이 전부. 필드도 구현도 없다.
- `Raw`·`Pretty`·`Compact` 어디에도 `implements Formatter` 가 **없다**. 메서드가 맞으면 그 순간 만족한다([[methods]]).
- 그래서 `io.Writer` 를 정의한 사람과 `bytes.Buffer` 를 만든 사람이 서로 몰라도 둘이 맞물린다.

## 왜 갈아 끼울 수 있나

```go
executor.Execute(..., f format.Formatter, ...)   // 고르는 곳과 쓰는 곳이 분리된다
    → f.Format(out, resp.Body)
```

포맷을 고르는 건 플래그를 읽는 `internal/cli`, 쓰는 건 응답이 도착하는 `internal/executor`. `switch` 로 했다면 executor 가 `"pretty"` 라는 **문자열의 의미**를 알아야 하고 포맷이 늘 때마다 executor 를 열어야 한다. 인터페이스로 두면 executor 는 약속 하나만 안다.

## 인터페이스 값은 (타입, 값) 쌍이다

```go
var w io.Writer = os.Stdout   // 타입=*os.File, 값=Stdout 을 가리키는 포인터
```

그래서 zero value 가 `nil` 이고([[variables]]), 원래 타입을 되찾는 **타입 단언**이 가능하다.

```go
f, ok := w.(*os.File)   // 두 값 형태 — 아니면 ok == false
f := w.(*os.File)       // 한 값 형태 — 아니면 panic
```

**항상 두 값 형태**를 쓴다. [[maps]] 의 `v, ok := m[k]` 와 같은 comma-ok 관용구로, Go 는 "실패할 수 있는 조회"를 전부 이 모양으로 통일한다.

`isTerminal` 이 이걸 쓴다 — `io.Writer` 는 `Write` 만 약속하므로 파일 모드를 물으려면 `*os.File` 껍질을 벗겨야 한다([[bitmasks]]). 테스트가 `bytes.Buffer` 를 넘기면 `ok == false` 라 자연스럽게 "터미널 아님"이 된다.

## 좁게 유지하기 — 설정은 구현이 들고 있는다

`Pretty` 는 경고를 어디로 보낼지 **필드**로 갖는다.

```go
type Pretty struct{ Warn io.Writer }

format.Pretty{}                        // Warn 은 zero value = nil → 조용히 폴백
format.Pretty{Warn: cmd.ErrOrStderr()} // → 폴백할 때 한 줄 남긴다
```

`Format(dst, src, errOut)` 처럼 인자로 만들면 경고할 일이 없는 `Raw` 까지 안 쓰는 인자를 받는다. **인터페이스는 좁게, 구현마다 다른 설정은 그 구현이 필드로.**

## 겪은 함정

- `Warn: cmd.ErrOrStderr()` 를 채우라는 지시에서 막혔다. 원인은 문법이 아니라 **"값이 설정을 들고 다닌다"** 가 낯설어서였다. 같은 타입을 두 모습(`Pretty{}` / `Pretty{Warn: …}`)으로 만들 수 있다는 걸 실행해서 보고 나서야 풀렸다 — `quiet.Warn == nil ? true` / `loud.Warn == nil ? false`.
- 인터페이스를 반환하는 함수에서 실패 시 `nil` 을 돌려주려면 타입이 인터페이스여야 한다. `format.Raw{}` 를 "빈 값" 삼아 돌려주면 호출자는 에러를 무시했을 때 조용히 raw 출력을 얻는다.

## 관련

[[methods]] · [[io-reader-writer]] · [[maps]] · [[variables]] · [[bitmasks]] · [[functions-as-values]]
