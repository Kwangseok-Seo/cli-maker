# environment-variables

프로세스가 시작될 때 부모에게서 **복사받는** 문자열 키-값 표. Go 에서는 `os` 패키지로 읽는다.

## 핵심

```go
v := os.Getenv("GITHUB_TOKEN")          // 없으면 ""
v, ok := os.LookupEnv("GITHUB_TOKEN")   // ok = "존재하는가"
```

- **복사본이다.** 자식이 `os.Setenv` 를 해도 부모 셸로 돌아가지 않는다 → `cli-maker login` 같은 명령이 셸에 토큰을 심어 줄 수 없는 이유.
- `Getenv` 는 **"미설정"과 "빈 값으로 설정됨"을 구별하지 못한다.** 둘 다 `""` 다.
- `v, ok :=` 는 Go 의 **comma-ok 관용구** — map 조회, 타입 단언, 채널 수신에도 같은 모양으로 나온다. "값 + 그 값이 진짜인가"를 한 번에 준다.

## 왜 비밀을 여기 두는가

매니페스트는 커밋되고 공유되는 데이터, 토큰은 사람마다 다른 비밀이다. 그래서 매니페스트에는 값이 아니라 **어느 환경변수를 볼지의 이름**만 담는다.

```yaml
auth:
  type: bearer
  env: GITHUB_TOKEN   # ← 값이 아니라 포인터
```

## 구별이 값을 하는 지점 — 실측

`export GITHUB_TOKEN=$(pass show gh/token)` 에서 `pass` 가 실패하면 변수는 **존재하되 빈 값**이 된다. 서버는 이 둘을 구별해 주지 않는다:

| 우리가 보낸 것 | GitHub `/user` 응답 |
|---|---|
| 헤더 없음 | 401 `Requires authentication` |
| `Authorization: Bearer` (빈 값) | 401 `Bad credentials` |

그래서 "설정 안 됨 / 설정됐는데 빈 값"을 갈라 말해 주는 건 우리 몫이다 — `ok` 가 없으면 그 문장을 쓸 수 없다.

## 겪은 함정

- **경고 문구에 `GITHUB_TOKEN` 을 문자열 상수로 박았다.** petstore 매니페스트를 돌렸더니 `warning: GITHUB_TOKEN is not set` 이 나왔다 — 유저는 존재하지도 않는 변수를 찾아 헤맨다. 이름은 이미 `Auth.Env` 로 받고 있었다. 하나의 실행기가 모든 매니페스트를 처리한다는 전제(CONTEXT.md 의 **Executor**)가 문자열 하나에 깨진 것.
- 같은 실패의 **거울상**을 30분 뒤에 또 밟았다 — 이번엔 env 파싱 에러에 변수 이름을 *아예 안* 실어서, `time: missing unit in duration "60"` 만 보고는 flag 때문인지 환경변수 때문인지 알 수 없었다([[error-wrapping]]).

## 관련

[[config-precedence]] · [[http-headers]] · [[error-wrapping]] · [[stdout-stderr]] · [[variables]]
