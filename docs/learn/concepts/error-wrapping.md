# error-wrapping

에러에 맥락을 덧붙이되 원본을 안에 보관하는 것. `fmt.Errorf` 의 `%w` 가 그 일을 한다.

## `%w` vs `%v`

```go
return 0, fmt.Errorf("%s: %w", timeoutEnv, err)   // 감싼다 — 원본이 안에 남는다
return 0, fmt.Errorf("%s: %v", timeoutEnv, err)   // 납작하게 — 문자열만 남는다
```

둘 다 화면에는 똑같이 보인다. 차이는 **나중에 기계가 안쪽을 알아볼 수 있는가**다.

```go
errors.Is(err, os.ErrNotExist)   // 감싸져 있어도 안쪽까지 뒤진다
```

`main.go` 는 `apis/` 디렉토리가 없을 때 `errors.Is(err, os.ErrNotExist)` 로 "없는 건 괜찮다"를 판별한다 — 중간에 누가 `%v` 로 납작하게 만들었다면 그 판별이 죽는다. `%w` 는 진단 문구를 앞에 붙이면서도 그 가능성을 안 잃는 방법이다.

## `errors.Join` — 여러 개를 나란히 품는다

`%w` 가 하나를 감싼다면, `errors.Join` 은 여러 개를 한 에러로 묶는다.

```go
var errs []error
errs = append(errs, fmt.Errorf("commands[%d]: name 이 비어 있다", i))
return errors.Join(errs...)
```

성질 셋:

1. **전부 nil 이거나 슬라이스가 비면 `nil` 을 반환한다** → 호출자는 `if err != nil` 만 쓰면 되고 "문제 없음"용 분기가 필요 없다.
2. 출력하면 각 에러가 **줄바꿈으로** 이어진다. 그래서 각 에러가 `\n` 으로 끝나면 빈 줄이 하나씩 낀다 — Go 의 에러 문자열이 관례상 개행·문장부호로 끝나지 않는 이유 중 하나.
3. `errors.Is` 가 **모든 가지를** 뒤진다.

검증처럼 문제가 여럿 나올 수 있는 곳에서 값을 한다. 하나 찾을 때마다 반환하면 유저가 고치러 여러 번 왕복한다.

`errors.Join` 이 만든 에러는 `Unwrap() []error` 를 가지므로 **다시 펼 수 있다.** 그 타입은 비공개라 이름을 적을 수 없지만, 필요한 건 이름이 아니라 계약뿐이다:

```go
joined, ok := err.(interface{ Unwrap() []error })
len(joined.Unwrap())   // 문제가 몇 개였나
```

## 센티널이 있을 때와 없을 때

`errors.Is` 로 정확히 판정하려면 **비교할 값**이 있어야 한다.

```go
errors.Is(err, context.DeadlineExceeded)   // ○ 표준 라이브러리가 값 하나로 정해 둔 센티널
strings.Contains(err.Error(), "중복이다")     // △ 우리 검증 에러 — 비교할 값이 없다
```

우리 `Validate` 의 에러는 전부 `errors.New`/`fmt.Errorf` 로 만든 **문자열**이라 서로 구별할 타입도 값도 없다. 그래서 테스트가 메시지 조각으로 판정할 수밖에 없고, 문구를 다듬으면 테스트가 깨진다.

반대쪽은 `context.DeadlineExceeded` 다 — `net/http` 가 `*url.Error` 로 감싸도 `errors.Is` 가 뚫고 찾는다. **센티널을 둘지 말지는 "이 에러를 기계가 분기에 쓸 것인가"로 갈린다.** 사람만 읽는다면 문자열로 충분하다.

## 진단은 출처를 말해야 한다

```
time: missing unit in duration "60"                     ← 어디서 온 값인지 모른다
CLI_MAKER_TIMEOUT: time: missing unit in duration "60"  ← 고칠 곳을 안다
```

`time.ParseDuration` 은 자기가 받은 문자열만 안다. 그게 `--timeout` 에서 왔는지 환경변수에서 왔는지는 **부른 쪽만** 아니까, 그 층에서 이름을 붙여 줘야 한다.

## 에러를 반환하는 함수는 출력하지 않는다

```go
fmt.Fprintln(os.Stderr, "parse:", err)   // ← 이러면
return timeout, err                       //   호출자(cobra)가 또 찍는다
```

실제로 같은 줄이 두 번 나왔다:

```
parse: time: missing unit in duration "60"
Error: time: missing unit in duration "60"
```

**반환이 곧 보고다.** 출력할지·어디에 할지는 호출자가 정한다. 라이브러리성 코드가 `os.Stderr` 에 직접 붙으면 테스트도 어려워진다([[stdout-stderr]]).

## 에러가 있으면 값은 zero

```go
return 0, fmt.Errorf(...)   // ○
return timeout, err         // × — 파싱 실패한 timeout 을 같이 넘긴다
```

호출자가 err 검사를 빠뜨렸을 때 반쯤 유효한 값이 흘러 들어가는 걸 막는 규약이다. 타임아웃에서는 그 값이 `0` 이라 `context.WithTimeout` 이 **즉시 만료**된다.

## 겪은 함정

- 에러 감싸기 줄을 **엉뚱한 층에** 넣었다. flag 층에 넣었더니 `--timeout 5s` 를 정상적으로 친 사람까지 무조건 실패했다(`err` 가 `nil` 이어도 `fmt.Errorf` 는 에러 값을 만든다). 컴파일러가 `declared and not used: duration` 으로 먼저 잡아 줬다 — 미사용 변수 에러를 `_` 로 덮었다면 그 버그가 살아남았을 것이다([[errors-compile-vs-runtime]]).

## 관련

[[error-handling]] · [[environment-variables]] · [[config-precedence]] · [[stdout-stderr]] · [[errors-compile-vs-runtime]]
