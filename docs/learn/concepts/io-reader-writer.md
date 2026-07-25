# io-reader-writer

메서드 하나짜리 인터페이스 둘이 Go 의 입출력 전체를 잇는다. 파일·네트워크·메모리 버퍼가 같은 구멍에 꽂힌다.

## 핵심

```go
type Reader interface { Read(p []byte) (n int, err error) }
type Writer interface { Write(p []byte) (n int, err error) }
```

- `os.Stdout`·`bytes.Buffer`·`http.ResponseWriter` = **Writer**
- `resp.Body`·`os.File`·`strings.Reader` = **Reader**
- 둘을 잇는 것이 `io.Copy(dst Writer, src Reader) (int64, error)`.

```go
io.Copy(out, resp.Body)   // 응답을 out 으로 흘려보낸다
```

## "흘려보낸다"는 뜻 — 실측

100KB 를 옮기며 `Write` 가 불린 횟수를 세면:

```
Read 만 있는 소스(= resp.Body):    32768 → 32768 → 32768 → 4096 바이트 (4회)
strings.Reader:                    102400 바이트 (1회)
```

- `io.Copy` 는 **32KB 버퍼 하나**로 "읽고 → 쓰고"를 반복한다. 응답이 500MB 여도 우리가 쓰는 메모리는 32KB.
- 대안인 `io.ReadAll(resp.Body)` 는 전체를 메모리에 쌓는다 — 500MB 응답 = 500MB RAM.
- `strings.Reader` 가 한 번에 간 것은 "나한테 맡기면 한 번에 써 주겠다"는 별도 메서드(`WriteTo`)를 갖고 있어 `io.Copy` 가 그 지름길로 샜기 때문. `resp.Body` 에는 없어서 일반 경로를 탄다.

[[file-io]] 의 `os.ReadFile` 이 통째로 읽는 것과 대비된다 — 매니페스트는 크기를 우리가 알고 작지만, HTTP 응답은 크기를 서버가 정한다.

## 인터페이스로 받으면 갈아 끼울 수 있다

```go
func Execute(..., out io.Writer) error   // os.Stdout 이 아니라 io.Writer
```

- 실제로는 `cmd.OutOrStdout()` 을 넘기지만, 테스트에서는 `&bytes.Buffer{}` 를 끼워 출력을 문자열로 붙잡을 수 있다(M7).
- `os.Stdout` 을 함수 안에 직접 박으면 그 함수는 영원히 화면에만 쓴다.

## 겪은 함정

- 데모를 `strings.Reader` 로 짜서 "32KB 씩 나뉜다"를 보이려다 **한 번에 다 가는** 출력을 얻었다. 지름길 최적화 때문. 인터페이스는 같아도 구현체마다 뒷길이 다르다 — 측정 대상이 실제 대상(`resp.Body`)과 같은 성질인지 먼저 확인해야 한다.

## 관련

[[net-http]] · [[file-io]] · [[functions-as-values]] · [[stdout-stderr]] · [[defer]]
