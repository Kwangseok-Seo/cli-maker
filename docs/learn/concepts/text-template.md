# text-template

데이터를 넣으면 **텍스트**가 나오는 틀. `{{...}}` 안이 액션, 나머지는 그대로 복사된다.

## 핵심

```go
t := template.Must(template.New("main.go").Parse(tmplText))
t.Execute(w, data)     // w 는 io.Writer — os.Stdout, &bytes.Buffer{}, 파일 다 같다
```

- **넘기는 값은 하나.** 여러 개가 필요하면 struct 로 감싼다.
- `.` 은 "지금 보고 있는 값" — **커서**다. `{{range}}` 안에서 이동한다.
- `$` 는 언제나 **루트**(Execute 에 넘긴 그 값).
- `text/template` vs `html/template`: 후자는 문맥별 자동 이스케이프. **Go 소스를 만들 땐 text 이므로 인용은 우리 몫**이다.

## 이 틀은 Go 를 모른다

가장 중요한 성질이다. 같은 템플릿에 이름만 바꿔 넣어 봤다:

```
Name="owner"      → 59 바이트 생성, Execute 에러 없음 → go/format: 유효한 Go
Name="repo-name"  → 67 바이트 생성, Execute 에러 없음 → go/format: 4:10: expected type, found '-'
```

두 번째도 **`Execute` 는 성공한다.** 템플릿 입장에선 문자열을 이어 붙였을 뿐이다. [[encoding-json]] 이 구조를 알고 깨진 것을 거부하는 것과 정반대다.

그래서 코드 생성기의 표준 관용구는 **2단**이다 — 템플릿으로 만들고, 문법 판정은 파서에게 맡긴다([[go-ast]]).

## 커서(`.`)와 루트(`$`)

```
{{range .Commands}}{{$.Name}} {{.Name}}  (인증: {{$.Auth.Env}})
{{end}}
```
```
gh whoami  (인증: GITHUB_TOKEN)      ← $.Name 은 매니페스트, .Name 은 Command
gh repo  (인증: GITHUB_TOKEN)
gh zen  (인증: GITHUB_TOKEN)
```

[[closures]] 와 대비하면 선명하다. 클로저는 렉시컬 스코프로 바깥 변수의 *칸*을 붙잡지만, 템플릿엔 스코프가 없고 **커서 하나가 이동**할 뿐이다.

## 데이터가 하나뿐이라 감싼다

매니페스트에 "원본 파일 경로"를 곁들이려면 감싸야 한다. 임베딩하면 템플릿이 짧아진다:

```go
type data struct {
	*manifest.Manifest    // .Name, .Commands 가 승격돼 템플릿에서 그대로 보인다
	Source string         // 얹은 것만 필드로
}
```

`{{.Manifest.Name}}` 이 아니라 `{{.Name}}` 으로 쓸 수 있다. text/template 은 리플렉션으로 필드를 찾으므로 **승격된 필드도 본다**(실측 확인).

## 함정 1 — 공백

액션 앞뒤의 개행이 그대로 출력에 나간다. 읽기 좋게 쓰면 빈 줄이 쏟아진다:

```
commands:                    commands:⏎
{{range .Commands}}          ⏎          ← 매 항목마다
  - {{.Name}}                  - whoami⏎
{{end}}                      ⏎
```

`{{- ` 와 ` -}}` 가 **그 방향의 공백을 먹는다**:

```
commands:
{{- range .Commands}}
  - {{.Name}}
{{- end}}
```
→ `commands:` / `  - whoami` / `  - repo` / `  - zen`

공백 제어를 제대로 하면 생성된 Go 소스가 **`gofmt -l` 에 걸리지 않는다**(실측).

## 함정 2 — 인용. 유저 문자열은 예외 없이 `%q`

험한 값 셋을 두 방식으로 흘린 결과:

| 입력 | 날것 `"{{.}}"` | `{{printf "%q" .}}` |
|---|---|---|
| `say"hi` | `"say"hi"` → 컴파일 안 됨 | `"say\"hi"` → 값 보존 |
| `C:\temp` | `"C:\temp"` → **유효한 Go 인데 값이 바뀜** | `"C:\\temp"` → 값 보존 |
| `two⏎lines` | `"two` → 리터럴 미종료 | `"two\nlines"` → 값 보존 |

가운데 줄이 위험하다. **컴파일은 통과하고 값만 조용히 틀린다**(`\t` 가 탭이 된다). 깨지는 쪽은 오히려 안전하다 — 즉시 발견되니까.

`printf "%q"` 는 템플릿에 내장된 함수로, `fmt.Sprintf("%q", v)` 와 같다. Go 리터럴 문법으로 인용해 준다.

## `Must` 를 쓰는 이유

```go
var mainTmpl = template.Must(template.New("main.go").Parse(mainTmplText))
```

패키지 변수라 **시작할 때 한 번** 파싱된다. 여기서 실패하면 *유저 입력* 문제가 아니라 *우리 템플릿의 버그*이므로 panic 이 맞다 — `manifest.Validate` 가 에러를 **돌려주는** 것과 층이 다르다([[error-handling]]).

## 겪은 함정

- **식별자 자리에 유저 문자열을 넣을 뻔했다.** `var {{.Name}} string` 같은 배치는 `repo-name`·`2fa` 에서 깨진다. 그런데 그 제약은 도메인이 아니라 **템플릿 설계에서 나온 것**이었다 — 명령마다 익명 블록 `{ ... }` 을 쓰면 변수명을 만들 필요가 없어져 유저 문자열이 전부 문자열 리터럴 자리에만 놓인다.
- **데모 harness 가 값을 바꿔 오진할 뻔했다.** 파이썬 `re.sub` 이 치환 문자열의 `\t` 를 탭으로 바꿔 놓아, 템플릿이 이상한 값을 낸 것처럼 보였다. **값이 예상과 다르면 대상보다 측정 장치를 먼저 의심**([[httptest]] 의 `0.30s` 와 같은 형태).

## 관련

[[go-ast]] · [[go-embed]] · [[closures]] · [[encoding-json]] · [[error-handling]] · [[struct-tags]]
