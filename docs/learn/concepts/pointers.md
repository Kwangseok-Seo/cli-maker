# pointers

값 자체가 아니라 값의 **주소(참조)**. `&x` = "x 의 주소". 복사본 대신 원본을 공유·수정할 때 쓴다.

## 핵심

- `&cobra.Command{...}` → 그 struct 의 포인터(`*cobra.Command`). cobra 가 명령을 공유·수정해야 해서 포인터를 요구.
- `*T` = "T 의 포인터" 타입, `&v` = "v 의 주소" 값.

## 언제 포인터 vs 값인가 (M10)

struct 필드에서는 판정 기준이 하나 더 있다 — **`nil` 이 필요한가.**

```go
Auth Auth    // 값 — Auth.Type == "" 이 "auth 없음"을 뜻할 수 있다
Body *Body   // 포인터 — "본문 없음"을 뜻할 필드가 없다
```

`nil` 을 가질 수 있는 타입은 포인터·슬라이스·map·채널·함수·인터페이스뿐이다. `bool` 은 `false` 가 "없음"인지 진짜 false 인지 못 나누고 `string` 도 `""` 가 그렇다. 그래서 **"없음"을 표현해야 하는데 sentinel 로 쓸 필드가 없으면 포인터**다 → [[absent-vs-empty]].

값 struct 를 고르는 것이 기본이다. 포인터는 nil 역참조 패닉의 자리를 만들기 때문이다. 다만 메서드는 nil 리시버에서도 부를 수 있으므로, 읽기 전에 갈라 두면 호출자가 매번 검사하지 않아도 된다 → [[methods]]:

```go
func (b *Body) ContentTypeOrDefault() string {
	if b == nil || b.ContentType == "" { return "application/json" }
	return b.ContentType
}
```

## 겪은 함정

- **포인터 필드는 `%+v` 가 펼쳐 주지 않는다.** `fmt` 는 맨 바깥 포인터 하나만 `&{...}` 로 펼치고, struct 안에 중첩된 포인터는 **주소**로 찍는다. `cli-maker parse` 의 출력이 `Body:0x3eeb9d5f4408` 이 되어 디버그 가치가 줄었다.

## 관련

[[structs]] · [[cobra]] · [[absent-vs-empty]] · [[methods]] · [[variables]]
