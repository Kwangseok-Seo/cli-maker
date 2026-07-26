# go-test

테스트는 별도 트리가 아니라 **소스 옆**에 살고, 단언 라이브러리 없이 `if` 로 쓴다. 규약은 놀랄 만큼 적다.

## 규약은 이것뿐

| 무엇 | 규약 |
|---|---|
| 파일 이름 | `*_test.go` — 이 파일들은 일반 빌드에 들어가지 않는다 |
| 함수 | `func TestXxx(t *testing.T)` |
| `Test` 다음 글자 | **소문자면 안 된다** |
| 형제들 | `BenchmarkXxx(*testing.B)` · `FuzzXxx(*testing.F)` · `ExampleXxx()` · `TestMain(*testing.M)` |
| 패키지 | 같은 패키지, 또는 `<패키지>_test` |

**이 목록 밖은 전부 자유다.** 테이블 변수·루프 변수·필드 이름은 우리가 짓는 이름이고 툴체인은 모른다([[table-driven-tests]]).

세 번째 줄은 조용히 건너뛰는 게 아니라 빌드를 막는다:

```
naming_test.go:6:6: Testlowercase has malformed name: first letter after 'Test' must not be lowercase
```

`go test` 가 실행 전에 **vet 검사 일부를 자동으로** 돌리기 때문이다([[go-toolchain]]).

## 단언은 손으로 쓴다

```go
if got != tt.want {
	t.Errorf("BuildURL() = %s, want %s", got, tt.want)
}
```

JUnit·AssertJ 에 해당하는 것이 없다. 대신 관례가 강하다 — 메시지는 `함수(입력) = 실제, want 기대` 순서. 표준 라이브러리가 전부 이 모양이라 따라가면 읽는 쪽이 순서를 다시 안 따져도 된다.

**`%s` 와 `%q` 는 골라 써야 한다.** 개행·공백의 유무가 계약이면 `%q` 여야 보인다.

```
out = Encourage flow.
, want Encourage flow.              ← %s: 다른 건지 같은 건지 안 보인다
out = "Encourage flow.\n", want "Encourage flow."   ← %q
```

## `Errorf` 와 `Fatalf`

- **`Errorf`** — 실패로 기록하고 **계속**. 판정이 여럿이면 이쪽(하나 틀려도 나머지를 본다).
- **`Fatalf`** — 그 서브테스트를 **중단**. 뒤 검사가 앞의 성공을 전제할 때(에러가 `nil` 인지 먼저 가른 뒤 `err.Error()` 를 볼 때 등).

`Fatalf` 는 **그 서브테스트만** 끊는다. 다른 서브테스트는 계속 돈다.

## 어느 패키지에 둘 것인가

```go
package executor        // 같은 패키지 — 비공개 함수도 부를 수 있다
package executor_test   // 외부 시점 — 공개 API 만 보인다
```

이 repo 는 같은 패키지를 쓴다. `applyAuth`·`isTerminal`·`rewrite`·`checkGlobal` 처럼 **밖에 안 내보낸 함수가 판단의 핵심**이라서다. 외부 패키지 방식은 "우리 API 를 남이 쓸 때의 모양"을 강제로 확인시켜 주는 게 장점이다.

## 실행

```
go test ./...                       # 전부
go test ./internal/format -v        # 서브테스트 이름까지
go test -run 'TestBuildURL/query' ./internal/executor
go test ./... -count=1              # 캐시 끄기
go test ./... -cover                # 커버리지
```

소스가 안 바뀐 패키지는 결과를 재사용하고 `(cached)` 로 표시된다.

## 테스트가 주는 도구들

```go
t.Helper()                       // 실패 위치를 헬퍼가 아니라 부른 줄로 찍는다
t.Cleanup(func() { f.Close() })  // 이 (서브)테스트가 끝나면 실행
t.TempDir()                      // 끝나면 통째로 지워지는 디렉토리
t.Setenv(k, v)                   // 끝나면 원래대로 ([[environment-variables]])
t.Context()                      // 테스트가 끝나면 취소되는 컨텍스트 ([[context]])
```

`t.Cleanup` 이 `defer` 보다 나은 자리가 있다 — **헬퍼 함수 안에서 연 자원**. 거기에 `defer` 를 쓰면 헬퍼가 반환하는 순간 닫혀 버린다.

## 겪은 함정

- **스켈레톤이 컴파일되지 않았다.** `m := ...` 를 선언해 놓고 쓰는 자리를 `// TODO` 로 비웠더니 `declared and not used: m`. 지역 변수는 선언만으로 **에러**다([[variables]]). 빈칸을 남긴 테스트 골격은 `_ = BuildURL(...)` 처럼 **일단 쓰이게** 해야 컴파일된다.
- **아무것도 단언하지 않는 테스트가 통과한다.** `_ = BuildURL(...)` 상태에서 `ok` 가 떴다 — 통과가 "확인했다"가 아니라 "아무것도 안 물어봤다"였다. 판정을 붙인 직후 **기대값을 일부러 틀리게 해 실패를 한 번 보는 것**이 가장 값싼 확인이다.
- **보고된 테스트 시간이 동작 시간이 아니다.** 타임아웃 테스트가 `0.30s` 로 찍혔지만 `Execute` 는 50.3ms 에 돌아왔다. 나머지는 `t.Cleanup` 의 `srv.Close()` 가 자고 있는 핸들러를 기다린 시간이다([[httptest]]).

## 관련

[[table-driven-tests]] · [[httptest]] · [[go-toolchain]] · [[variables]] · [[errors-compile-vs-runtime]] · [[environment-variables]]
