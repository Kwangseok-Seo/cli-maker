package manifest

type Param struct {
	Name     string `yaml:"name"`
	In       string `yaml:"in"`
	Type     string `yaml:"type"`
	Required bool   `yaml:"required"`
}

type Auth struct {
	Type string `yaml:"type"`
	Env  string `yaml:"env"`
}

type Body struct {
	Required    bool   `yaml:"required"`
	ContentType string `yaml:"contentType"`
}

// defaultContentType 은 매니페스트가 contentType 을 적지 않았을 때 붙는 값이다.
// 상수가 여기 하나뿐이라는 것이 요점 — Validate 가 채우면 두 벌이 된다 (ADR-0010).
const defaultContentType = "application/json"

// ContentTypeOrDefault 는 요청에 실제로 붙일 Content-Type 을 돌려준다.
//
// nil 리시버에도 안전하다. Go 에서 포인터 리시버 메서드는 리시버가 nil 이어도 호출되며,
// 그 안에서 필드를 읽을 때만 패닉이 난다 — 읽기 전에 갈라 두면 호출자가 매번
// nil 검사를 늘어놓지 않아도 된다.
//
// 런타임 경로와 생성된 CLI 가 clirun.Body 별칭을 통해 이 한 함수를 공유한다.
func (b *Body) ContentTypeOrDefault() string {
	if b == nil || b.ContentType == "" {
		return defaultContentType
	}
	return b.ContentType
}

// omitempty 가 이 두 칸에만 붙어 있는 것은 의도된 것이다.
//
// 이 struct 를 마샬하는 자리가 둘인데 요구가 반대다 — parse 는 디버그 도구라
// "이 자리가 비었다"를 보여 줘야 하고(오타 난 키는 조용히 버려지므로, 빈 값이
// 유일한 단서다), import 는 유저가 손으로 이어 쓸 초안을 내므로 없는 것을
// 있다고 적으면 안 된다.
//
// 다행히 두 요구가 서로 다른 필드에 걸린다. 노이즈를 내는 것은 Params·Body 이고
// (`params: []`, `body: null`), 오타를 드러내는 것은 Auth 와 Param.Required 다.
// 그래서 앞의 둘에만 붙였다. 실측: import 산출물 180줄 → 161줄, parse 는
// autth/requiredd 오타를 그대로 드러낸다.
type Command struct {
	Name   string  `yaml:"name"`
	Method string  `yaml:"method"`
	Path   string  `yaml:"path"`
	Params []Param `yaml:"params,omitempty"`
	Body   *Body   `yaml:"body,omitempty"`
}

type Manifest struct {
	Name     string    `yaml:"name"`
	BaseURL  string    `yaml:"baseUrl"`
	Auth     Auth      `yaml:"auth"`
	Commands []Command `yaml:"commands"`
}
