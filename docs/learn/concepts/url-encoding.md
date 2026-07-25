# url-encoding

유저가 준 값을 URL 에 안전하게 끼워 넣기. **경로와 쿼리는 규칙이 다르다.**

## 핵심

```go
url.PathEscape("a b/c")   // "a%20b%2Fc"   ← 경로 조각용
q := url.Values{}
q.Set("status", "a b/c")
q.Encode()                // "status=a+b%2Fc"  ← 쿼리용 (+ 정렬·이스케이핑 포함)
```

같은 입력 `a b/c` 인데 공백이 갈린다:

| | 공백 | `/` |
|---|---|---|
| `PathEscape` (경로) | `%20` | `%2F` |
| `Values.Encode` (쿼리) | `+` | `%2F` |

버그가 아니라 **서로 다른 표준**이다. 쿼리는 폼 인코딩 관례(공백 = `+`)를, 경로는 URI 규칙(공백 = `%20`)을 따른다. 양쪽 다 `/` 를 `%2F` 로 막는 것이 중요하다 — 안 그러면 값 하나가 경로를 한 칸 더 파고든다.

## 조립 패턴 (cli-maker)

```go
path := c.Path            // 누적자 — 원본을 두고 여기에 쌓는다
q := url.Values{}         // 누적자 — 루프 밖에서 한 번만
for _, p := range c.Params {
    v := values[p.Name]
    switch p.In {
    case "path":
        path = strings.ReplaceAll(path, "{"+p.Name+"}", url.PathEscape(v))
    case "query":
        if v != "" { q.Set(p.Name, v) }
    }
}
if qs := q.Encode(); qs != "" { path += "?" + qs }
```

- 루프는 **매니페스트의 `c.Params` 기준**으로 돈다. 유저 입력(`values`)만 돌면 그 이름이 path 인지 query 인지(`p.In`) 알 길이 없다 — `In` 은 매니페스트에만 있는 정보다.
- `values[p.Name]` 은 없는 키여도 패닉하지 않고 빈 문자열을 준다(map 의 zero value). 그래서 "안 줬다"와 "빈 값"이 한 규칙으로 처리된다.
- `if qs := ...; qs != ""` — 쿼리가 없을 때 `?` 만 달랑 붙는 것을 막는다.

## 겪은 함정

- `strings.ReplaceAll(path, ...)` 의 **반환값을 버려** 치환이 없던 일이 됐다. 문자열은 불변이라 원본은 안 바뀐다 → [[unused-results]].
- `q := url.Values{}` 를 루프 **안**에 둬서 매 반복마다 새로 태어났다. 누적자는 루프 밖.
- 변수명을 `url` 로 지으면 패키지 `net/url` 이 가려진다. 그 파일에서 `url.Values` 가 필요해지는 순간 터진다 — `reqURL` 처럼 피한다.

## 관련

[[net-http]] · [[unused-results]] · [[slices-and-args]] · [[structs]]
