package openapi

import (
	"testing"

	"github.com/Kwangseok-Seo/cli-maker/internal/manifest"
)

// convertAll 은 petstore 19개를 전부 옮겨, 이름으로 찾을 수 있게 모아 준다.
func convertAll(t *testing.T) (map[string]*manifest.Command, map[string][]string, []string) {
	t.Helper()

	cmds := map[string]*manifest.Command{}
	warns := map[string][]string{}
	var skipped []string

	for _, op := range loadPetstore(t).Operations() {
		cmd, w, err := toCommand(op)
		if err != nil {
			skipped = append(skipped, op.Operation.OperationID)
			continue
		}
		cmds[cmd.Name] = cmd
		if len(w) > 0 {
			warns[cmd.Name] = w
		}
	}
	return cmds, warns, skipped
}

// petstore 19개가 전부 넘어와야 한다. 못 넘기는 건 deletePet 의 optional header
// param 하나뿐이고, 그건 operation 을 죽이지 않는다.
func TestToCommandCoversPetstore(t *testing.T) {
	cmds, warns, skipped := convertAll(t)

	if len(skipped) != 0 {
		t.Errorf("스킵된 operation = %v, want 없음", skipped)
	}
	if len(cmds) != 19 {
		t.Errorf("옮겨진 명령 수 = %d, want 19", len(cmds))
	}

	// 경고는 딱 한 곳에서만 나와야 한다 — 그 외에서 나오면 뭔가를 조용히 잃고 있다.
	if len(warns) != 1 || len(warns["deletePet"]) != 1 {
		t.Errorf("경고 = %v, want deletePet 하나만", warns)
	}
}

func TestToCommandParams(t *testing.T) {
	cmds, _, _ := convertAll(t)

	tests := []struct {
		cmd  string
		want []manifest.Param
	}{
		// path param 하나. schema.type 이 Param.Type 으로 넘어온다.
		{"getPetById", []manifest.Param{
			{Name: "petId", In: "path", Type: "integer", Required: true},
		}},
		// query param 인데 required 다 — in 과 required 는 서로 독립이라는 확인.
		{"findPetsByStatus", []manifest.Param{
			{Name: "status", In: "query", Type: "string", Required: true},
		}},
		// optional query param 둘. 스펙에 적힌 순서가 그대로 보존돼야 한다
		// (parameters 는 리스트라 map 과 달리 순서가 있다).
		{"loginUser", []manifest.Param{
			{Name: "username", In: "query", Type: "string", Required: false},
			{Name: "password", In: "query", Type: "string", Required: false},
		}},
		// api_key(header, optional)는 빠지고 petId 만 남아야 한다.
		{"deletePet", []manifest.Param{
			{Name: "petId", In: "path", Type: "integer", Required: true},
		}},
		// param 이 아예 없는 operation.
		{"getInventory", nil},
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			got := cmds[tt.cmd]
			if got == nil {
				t.Fatalf("%s 가 안 옮겨졌다", tt.cmd)
			}
			if len(got.Params) != len(tt.want) {
				t.Fatalf("params = %+v, want %+v", got.Params, tt.want)
			}
			for i := range tt.want {
				if got.Params[i] != tt.want[i] {
					t.Errorf("params[%d] = %+v, want %+v", i, got.Params[i], tt.want[i])
				}
			}
		})
	}
}

func TestToCommandMethodAndPath(t *testing.T) {
	cmds, _, _ := convertAll(t)

	for _, tt := range []struct{ name, method, path string }{
		{"getPetById", "GET", "/pet/{petId}"},
		{"addPet", "POST", "/pet"},
		{"deleteOrder", "DELETE", "/store/order/{orderId}"},
		{"updateUser", "PUT", "/user/{username}"},
	} {
		got := cmds[tt.name]
		if got == nil {
			t.Errorf("%s 가 안 옮겨졌다", tt.name)
			continue
		}
		if got.Method != tt.method || got.Path != tt.path {
			t.Errorf("%s = %s %s, want %s %s", tt.name, got.Method, got.Path, tt.method, tt.path)
		}
	}
}

func TestToCommandBody(t *testing.T) {
	cmds, _, _ := convertAll(t)

	tests := []struct {
		cmd  string
		want *manifest.Body // nil = 본문이 없어야 한다
	}{
		// 후보 셋 중 json 을 고른다.
		{"addPet", &manifest.Body{Required: true, ContentType: "application/json"}},
		// 후보가 하나뿐이고 json 이 아니다. required 키가 스펙에 없으므로 false.
		{"uploadFile", &manifest.Body{Required: false, ContentType: "application/octet-stream"}},
		// 본문이 없는 operation — 포인터가 nil 이어야 한다.
		// (여기가 zero value 로 채워지면 --data 가 붙지 말아야 할 명령에 붙는다.)
		{"getPetById", nil},
	}

	for _, tt := range tests {
		t.Run(tt.cmd, func(t *testing.T) {
			got := cmds[tt.cmd]
			if got == nil {
				t.Fatalf("%s 가 안 옮겨졌다", tt.cmd)
			}
			switch {
			case tt.want == nil && got.Body != nil:
				t.Errorf("body = %+v, want nil", got.Body)
			case tt.want != nil && got.Body == nil:
				t.Errorf("body = nil, want %+v", tt.want)
			case tt.want != nil && *got.Body != *tt.want:
				t.Errorf("body = %+v, want %+v", *got.Body, *tt.want)
			}
		})
	}
}

func TestPickContentType(t *testing.T) {
	set := func(keys ...string) map[string]struct{} {
		m := map[string]struct{}{}
		for _, k := range keys {
			m[k] = struct{}{}
		}
		return m
	}

	tests := []struct {
		name        string
		in          map[string]struct{}
		want        string
		wantWarning bool
	}{
		{"비어 있으면 빈 문자열", set(), "", false},
		{"하나뿐이면 그것", set("application/octet-stream"), "application/octet-stream", false},
		{"json 이 있으면 json", set("application/xml", "application/json", "text/plain"), "application/json", false},
		// 사전순이면 cbor 가 앞이지만 json 선호가 이긴다 — 선호가 실제로 동작하는지 본다.
		{"json 선호가 사전순을 이긴다", set("application/cbor", "application/json"), "application/json", false},
		{"json 이 없고 여럿이면 사전순 첫 번째 + 경고", set("application/xml", "application/cbor"), "application/cbor", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, warn := pickContentType(tt.in)
			if got != tt.want {
				t.Errorf("contentType = %q, want %q", got, tt.want)
			}
			if (warn != "") != tt.wantWarning {
				t.Errorf("warning = %q, want warning=%v", warn, tt.wantWarning)
			}
		})
	}
}
