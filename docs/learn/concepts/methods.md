# methods

함수 이름 앞에 괄호 하나가 더 붙으면 그 함수는 특정 타입에 **속하게** 된다. 인터페이스를 만족시키는 유일한 수단.

## 핵심

```go
func BuildURL(m *manifest.Manifest, ...) string        // 함수
func (r Raw) Format(dst io.Writer, src io.Reader) error // 메서드
//   ^^^^^^^ 리시버(receiver)
```

- 리시버가 붙으면 `Raw{}.Format(...)` 처럼 **값에서** 부른다.
- 리시버 이름은 관례상 타입 첫 글자 한두 자(`r Raw`, `p Pretty`). `this`/`self` 는 쓰지 않는다.
- **몸통에서 리시버를 안 쓰면 이름을 생략해도 된다** — `func (Raw) Format(...)`. `Raw` 는 필드가 없어 실제로 쓸 게 없다.

## 값 리시버 vs 포인터 리시버

```go
func (p Pretty) Format(...)   // 값 — 복사본을 받는다
func (p *Pretty) Format(...)  // 포인터 — 원본을 받는다
```

- 메서드가 **자기 필드를 바꿔야 하면** 포인터여야 한다. 값 리시버는 복사본을 고치고 버린다.
- cli-maker 의 포맷터는 셋 다 읽기만 하고 크기도 작아(`Pretty` 는 `io.Writer` 필드 하나) 값 리시버로 충분하다.
- 한 타입의 메서드는 **전부 값이거나 전부 포인터**로 맞추는 게 관례다. 섞으면 인터페이스 만족 규칙이 헷갈린다.

## 메서드 집합이 인터페이스 만족을 정한다

```go
type Formatter interface{ Format(io.Writer, io.Reader) error }

var _ Formatter = Raw{}      // OK — Raw 에 Format 이 있다
var _ Formatter = Pretty{}   // OK
```

`implements` 를 쓰는 곳이 **없다**. 메서드가 있으면 그걸로 끝 — 자세한 것은 [[interfaces]].

주의: 포인터 리시버로 만들면 `Pretty{}` 는 만족하지 못하고 `&Pretty{}` 만 만족한다. 값에서 메서드를 부를 땐 Go 가 자동으로 주소를 잡아 주지만([[pointers]]), 인터페이스에 담을 땐 그 편의가 없다.

## 겪은 함정

- 이 프로젝트에서 **M7 전까지 메서드를 한 번도 안 썼다**(`grep '^func ('` 이 0건). 구조체는 [[structs]] 에서 M2 에 나왔지만 전부 데이터 그릇이었고, 동작은 전부 자유 함수였다. 인터페이스가 필요해진 순간에야 메서드가 등장한 셈 — Go 에서 메서드는 "객체를 만들려고"가 아니라 "약속을 만족시키려고" 쓴다.

## 관련

[[interfaces]] · [[structs]] · [[pointers]] · [[functions-as-values]] · [[io-reader-writer]]
