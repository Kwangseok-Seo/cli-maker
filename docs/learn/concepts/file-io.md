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

## 디렉토리 읽기 (M3)

```go
entries, err := os.ReadDir("apis") // ([]os.DirEntry, error)
for _, e := range entries {
    if filepath.Ext(e.Name()) != ".yaml" {
        continue
    }
    path := filepath.Join("apis", e.Name())
}
```

- `os.ReadDir` 은 **파일명 순으로 정렬된** 목록을 준다 → `--help` 에 API 가 나타나는 순서가 실행마다 같다(무작위면 유저 경험이 흔들린다).
- `e.Name()` 은 **파일명만** 준다(경로 없음). 경로는 `filepath.Join` 으로 — 문자열에 `/` 를 직접 붙이지 않는다. Windows 는 `apis\example.yaml`, Linux 는 `apis/example.yaml` 로 알아서 갈린다.
- 실패하면 `entries` 는 `nil` 이고 `range nil` 은 0회 돈다 → "디렉토리 없음"에 특별 분기가 필요 없다 → [[slices-and-args]].

## 겪은 함정

- M2 에선 깨끗이 통과. 다만 읽은 바이트는 아직 "그냥 텍스트"다 — 타입 있는 값(`Manifest`)이 되려면 다음 단계 [[serialization]] 이 필요하다.
- M3: 디렉토리가 없을 때를 "에러니까 뭔가 해야 한다"고 본 것이 과잉이었다. `nil` + `range` 의 성질을 알면 분기 자체가 사라진다 → [[error-handling]].

## 관련

[[slices-and-args]] · [[error-handling]] · [[serialization]]
