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

type Command struct {
	Name   string  `yaml:"name"`
	Method string  `yaml:"method"`
	Path   string  `yaml:"path"`
	Params []Param `yaml:"params"`
	Body   *Body   `yaml:"body"`
}

type Manifest struct {
	Name     string    `yaml:"name"`
	BaseURL  string    `yaml:"baseUrl"`
	Auth     Auth      `yaml:"auth"`
	Commands []Command `yaml:"commands"`
}
