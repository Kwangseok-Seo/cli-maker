# table-driven-tests

케이스를 **슬라이스로 적고** 루프가 한 번 돈다. 새 케이스는 코드가 아니라 데이터로 는다.

## 핵심

```go
tests := []struct {
	name string
	in   int
	want int
}{
	{name: "영", in: 0, want: 0},
	{name: "하나", in: 1, want: 2},
}

for _, tt := range tests {
	t.Run(tt.name, func(t *testing.T) {
		if got := Double(tt.in); got != tt.want {
			t.Errorf("Double(%d) = %d, want %d", tt.in, got, tt.want)
		}
	})
}
```

- `struct { ... }` 는 **이름 없이 그 자리에서 만든 타입**이다. 이 테이블 말고 아무도 안 쓸 타입이라 이름을 안 붙인다([[composite-literals]]).
- `t.Run` 이 각 행을 **서브테스트**로 만든다. 실패하면 어느 행인지 이름으로 나오고, `-run 'TestX/이름'` 으로 하나만 돌릴 수 있다.
- 여러 테스트가 같은 모양을 공유하면 `type formatCase struct{...}` 로 이름을 붙여 빼는 편이 낫다. 익명인 것은 필수가 아니다.

## 이름은 전부 우리가 짓는다

`name`·`in`·`want` 는 규약이 아니라 **관례**다. 툴체인은 이 이름들을 모른다 — 실제로 바꿔서 돌려 봤다:

```go
케이스들 := []struct{ 제목, 넣는것, 나와야하는것 string }{...}
for _, 케이스 := range 케이스들 {
	t.Run(케이스.제목, func(t *testing.T) { ... })   // 통과한다
}
```

이름이 연결된 곳은 `t.Run` 의 첫 인자 **하나뿐**이고, 그건 그냥 "서브테스트 이름으로 쓸 문자열"이다. 그래도 `tests`·`tt`·`got`·`want` 를 쓰는 이유는 표준 라이브러리가 그래서 — 남이 훑고 지나갈 수 있다.

필드 이름은 **무엇과도 충돌하지 않는다**. 항상 `tt.name` 처럼 무언가에 붙어서만 불리므로, 변수 이름과 달리 `byte`·`len` 같은 미리 선언된 이름을 가릴 일이 없다([[identifier-shadowing]]).

## 안쪽 `t` 가 바깥 `t` 를 가린다 — 의도적이다

```go
t.Run(tt.name, func(t *testing.T) { ... })
//                      ^ 이 t 는 서브테스트 전용
```

서브테스트마다 별개의 `*testing.T` 를 받아야 실패가 그 서브테스트에만 귀속된다. 가리는 것이 늘 사고는 아니다 — **모르고** 가리는 것이 사고다([[identifier-shadowing]]).

## 필드에 함수를 담는다

값으로 적기 어려운 케이스는 **어떻게 만들지/망가뜨릴지**를 함수로 적는다([[functions-as-values]]).

```go
damage func(*Manifest)              // 정상 매니페스트를 한 군데만 망가뜨린다
newf   func(warn io.Writer) Formatter  // 경고 받을 곳을 받아 포매터를 만든다
w      func(t *testing.T) io.Writer    // 임시 파일·디바이스를 그 자리에서 연다
```

`damage` 형태의 이점이 크다 — 행마다 입력을 통째로 적으면 **실수로 두 군데가 깨져도 모른다**. 정상에서 한 군데만 바꾸면 "이 규칙이 잡혔다"가 코드에 드러난다. 함수 타입의 zero value 는 `nil` 이라 "망가뜨릴 게 없다"를 `nil` 로 적을 수 있다.

## 겪은 함정

- **기대값을 직관으로 적으면 틀린다.** query param 을 `zebra`, `apple` 순으로 선언했으니 URL 도 그 순서일 거라 생각했지만 `?apple=2&zebra=1` 이다 — `url.Values.Encode` 가 키를 정렬한다([[url-encoding]]). **코드가 옳은데 테스트가 틀리는** 실패는 원인 찾기가 더 오래 걸린다. 테이블을 쓰기 전에 실제 출력을 한 번 재는 편이 싸다.
- **한 행이 문제를 하나만 일으킨다는 보장이 없다.** `in: "quer"` 오타 하나가 검증 에러 **둘**을 만든다(모르는 `in` + 그 결과 치환할 param 이 없어짐). 개수까지 판정하면 이런 게 드러나고, 나중에 중복을 줄일지 결정을 강제한다([[error-wrapping]]).

## 관련

[[go-test]] · [[composite-literals]] · [[functions-as-values]] · [[identifier-shadowing]] · [[closures]] · [[url-encoding]]
