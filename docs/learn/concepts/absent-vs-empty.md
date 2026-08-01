# absent-vs-empty

"안 줬다"와 "빈 것을 줬다"는 다른 사실이다. 값 하나가 그 둘을 담지 못하면 어딘가에서 조용히 뭉개진다.

## 왜 뭉개지나

Go 의 zero value 는 **"채워지지 않았다"와 "0/`""`/false 를 채웠다"가 같은 비트**다 → [[variables]]. 그래서 값만 보고는 구별할 수 없고, 구별하려면 값 **밖에** 표식이 있어야 한다.

Go 표준 라이브러리의 관용구는 대개 **두 번째 반환값**이다.

```go
v, ok := m[key]          // [[maps]] — 없는 키인가, zero value 가 들어 있는가
v, ok := os.LookupEnv(k) // [[environment-variables]] — 안 set 인가, 빈 문자열인가
v, ok := x.(T)           // [[interfaces]] — 타입 단언
```

`os.Getenv` 하나만 쓰면 그 구별이 사라진다. M5 에서 `LookupEnv` 를 고른 이유가 이것이었다 — 토큰이 **없는 것**과 **빈 것**의 경고 문구가 달라야 했다.

## M10 — 같은 문제가 세 층에 연달아 나왔다

| 층 | "없음" | "빈 것" | 표식 |
|---|---|---|---|
| YAML → struct | `body:` 키가 없다 | `body: {}` | **포인터** — `*Body` 가 `nil` |
| 명령줄 flag | `--data` 를 안 줬다 | `--data ""` | **`Changed()`** — pflag 가 "설정됐는지"를 따로 기록 |
| HTTP 요청 | 본문이 없다 | 길이 0 인 본문 | **`Content-Type` 유무** |

세 답이 다 다르지만 구조는 같다 — *값 밖에 한 비트를 더 둔다.*

### 1층 — 값 struct 는 구별하지 못한다 (실측)

```
(1) body: 키가 없음
    값 struct  Body {Required:false ContentType:}
    포인터     Body <nil>

(2) body: {}
    값 struct  Body {Required:false ContentType:}     ← (1) 과 글자 하나 안 다르다
    포인터     Body &{Required:false ContentType:}
```

`yaml.Unmarshal` 은 없는 키에 **아무것도 하지 않는다**. 그러면 필드에는 struct 가 만들어질 때의 zero value 가 남고, 그건 `{}` 를 파싱한 결과와 같다 → [[serialization]].

`Auth` 는 값 struct 인데 왜 괜찮았나? `Auth.Type == ""` 이 sentinel 노릇을 하기 때문이다. **선례가 성립한 이유가 따로 있었던 것이지 값 struct 라서 괜찮았던 게 아니다** — `Body` 에는 그 자리에 놓을 필드가 없었다 → [[pointers]].

### 2층 — pflag 의 `Changed`

```go
if !cmd.Flags().Changed("data") { return nil, nil }  // 안 준 것
raw, _ := cmd.Flags().GetString("data")              // 준 것 (빈 문자열일 수도)
```

`GetString` 만 쓰면 `--data ""` 가 "안 줬다"와 같아진다 → [[cobra]].

### 3층 — 본문 없는 요청 vs 빈 본문

```
--data '{"a":1}'   body:"{\"a\":1}"  contentLength:7  contentType:"application/json"
--data ''          body:""           contentLength:0  contentType:"application/json"
(--data 없음)      body:""           contentLength:0  contentType:""
```

가운데와 아래가 **본문만 보면 같다**. 서버가 둘을 가르는 표식은 `Content-Type` 의 유무다 — 그래서 본문 없는 요청에 `Content-Type` 을 붙이는 것은 거짓말이 된다 → [[http-headers]] · [[net-http]].

## 겪은 함정

- **한 층을 풀고 다른 층에서 같은 문제를 못 알아봤다.** 1층을 포인터로 푼 직후 2층에서 `raw == ""` 로 판정하는 코드를 썼다. mutation 으로 되돌려 보고서야 같은 문제였음이 드러났다 — `--data ""` 케이스가 유일하게 깨졌다.
- **구별할 수 있음 ≠ 구별할 필요 있음.** 포인터로 바꾸면 nil 을 구별할 수 있다는 것과, 구별해야 할 이유가 있다는 것은 다른 문제다. 이유는 "본문을 안 받는 명령에 `--data` 를 등록하지 않기 위해"였고, 그게 없었다면 값 struct 로 충분했다.

## 관련

[[pointers]] · [[variables]] · [[maps]] · [[environment-variables]] · [[serialization]] · [[cobra]] · [[net-http]] · [[http-headers]]
