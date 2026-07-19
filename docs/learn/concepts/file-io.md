# file-io

파일을 통째로 바이트로 읽어들이기. 매니페스트를 프로그램 안으로 들이는 첫 관문.

## 핵심

```go
data, err := os.ReadFile("apis/example.yaml")  // ([]byte, error)
// data = 파일 전체의 바이트, err = 없으면 nil
```

- `os.ReadFile` 은 파일을 **한 번에** `[]byte` 로 읽는다 — 매니페스트처럼 작은 파일에 적합(스트리밍 불필요).
- 반환이 둘(데이터, 에러) → [[error-handling]] 의 `if err != nil` 로 **즉시** 검사.
- `[]byte` 는 바이트들의 슬라이스 → [[slices-and-args]]. `len(data)` 로 크기, `string(data)` 로 문자열화.

## 겪은 함정

- M2 에선 깨끗이 통과. 다만 읽은 바이트는 아직 "그냥 텍스트"다 — 타입 있는 값(`Manifest`)이 되려면 다음 단계 [[serialization]] 이 필요하다.

## 관련

[[slices-and-args]] · [[error-handling]] · [[serialization]]
