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

type Command struct {
	Name   string  `yaml:"name"`
	Method string  `yaml:"method"`
	Path   string  `yaml:"path"`
	Params []Param `yaml:"params"`
}

type Manifest struct {
	Name     string    `yaml:"name"`
	BaseURL  string    `yaml:"baseUrl"`
	Auth     Auth      `yaml:"auth"`
	Commands []Command `yaml:"commands"`
}
