# config-precedence

같은 설정값을 여러 곳에서 줄 수 있을 때, 어느 것이 이기는지의 규칙. cli-maker 의 타임아웃은 세 층이다.

## 수명이 순서를 정한다

| 층 | 이 값이 사는 기간 |
|---|---|
| `--timeout 5s` (flag) | **이번 실행 한 번** |
| `CLI_MAKER_TIMEOUT=60s` (env) | 이 셸 세션 / 이 CI job |
| 코드의 `defaultTimeout` | 프로그램에 영구히 |

**좁은 수명이 넓은 수명을 이긴다.** CI 에 60초를 깔아 놨어도 지금 이 한 번만 5초로 잘라 보는 게 정상적인 요구다. 거꾸로면 flag 는 있으나 마나다.

## 각 층은 "값"이 아니라 "값 + 줬는가"를 답해야 한다

```go
if cmd.Flags().Changed(TimeoutFlag) {        // ← 유저가 실제로 쳤는가
    return cmd.Flags().GetDuration(TimeoutFlag)
}
if v, ok := os.LookupEnv(timeoutEnv); ok {   // ← 존재하는가
    ...
}
return defaultTimeout, nil
```

## 기본값은 "안 줬다"를 지운다 — 실측

pflag 에 기본값을 주고 `GetDuration` 만 보면 이렇게 된다:

```
유저가 친 것: []                 → GetDuration=30s   Changed=false
유저가 친 것: [--timeout 30s]    → GetDuration=30s   Changed=true
유저가 친 것: [--timeout 5s]     → GetDuration=5s    Changed=true
```

**1행과 2행이 `GetDuration` 으로는 완전히 같다.** 기본값이 빈 자리를 그럴듯한 값으로 메워서, "이 층엔 답이 없으니 다음 층에 물어보자"는 판단 자체가 불가능해진다. flag 가 최우선인데 flag 층이 항상 답을 갖고 있다고 주장하니 env 는 영영 차례가 오지 않는다. **로드 순서를 바꿔도 소용없다** — 마지막에 flag 를 적용하는 순간 똑같이 덮어써진다.

`Changed(name) bool` 이 [[environment-variables]] 의 `ok` 와 같은 자리에 있는 메서드다.

## 이 프로젝트에서 세 번 반복된 문제

zero value 로는 "미지정"을 표현할 수 없다는 같은 얘기가 형태를 바꿔 세 번 나왔다:

1. **M4** — flag 를 전부 `String` 으로 받고 `""` = 미지정 규약([ADR-0004](../../adr/0004-params-as-flags.md)). `GetInt` 는 미지정을 `0` 으로 줘서 `?count=0` 과 구별되지 않았다.
2. **M5 인증** — `os.Getenv` 대신 `os.LookupEnv` 의 `ok`.
3. **M5 타임아웃** — pflag 기본값 대신 `Changed`.

## 겪은 함정

- 층마다 다른 태도를 취할 뻔했다. `--timeout abc` 는 pflag 가 `RunE` 에 닿기도 전에 exit 1 로 거절하는데, env 파싱 실패를 조용히 기본값으로 떨어뜨리면 `CLI_MAKER_TIMEOUT=60`(단위 없음)을 깐 유저는 60초로 도는 줄 믿는다. **같은 상황에 층마다 다른 답을 주지 않는다** — env 도 에러로 거절한다.
- persistent flag 이름은 **예약어가 된다.** 매니페스트 param 이 `timeout` 이면 로컬 flag 가 루트의 persistent 를 가려서 `Changed` 는 `true` 인데 `GetDuration` 은 `trying to get duration value of flag of type string` 을 낸다. 죽지는 않지만(그 명령만 exit 1) 등록 전에 검사해야 할 목록에 하나 더 늘었다.

## 관련

[[environment-variables]] · [[cobra]] · [[context]] · [[error-handling]]
