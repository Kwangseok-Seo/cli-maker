package manifest

import (
	"strings"
	"testing"
)

// validManifest 는 검증을 통과하는 최소 매니페스트를 새로 만든다.
//
// 케이스마다 여기서 딱 한 군데만 망가뜨린다 — 그래야 나온 에러가 그 한 군데
// 때문이라고 말할 수 있다. 매번 새로 만드는 이유는 앞 케이스의 손상이 다음
// 케이스로 새지 않게 하기 위해서다.
func validManifest() *Manifest {
	return &Manifest{
		Name:    "gh",
		BaseURL: "https://api.github.com",
		Commands: []Command{{
			Name:   "repo",
			Method: "GET",
			Path:   "/repos/{owner}/{repo}",
			Params: []Param{
				{Name: "owner", In: "path", Required: true},
				{Name: "repo", In: "path", Required: true},
			},
		}},
	}
}

// TestValidate 는 매니페스트 하나만 보고 판정하는 규칙들을 못 박는다.
//
// 확인하는 것은 두 가지다.
//   - 어떤 문제가 잡히는가 (want 의 조각이 메시지에 있는가)
//   - 몇 개가 잡히는가 (len(want) — M6 의 "한 번에 모아서 낸다" 계약)
func TestValidate(t *testing.T) {
	tests := []struct {
		name string
		// damage 는 정상 매니페스트를 망가뜨린다. nil 이면 망가뜨리지 않는다.
		damage func(*Manifest)
		// want 는 에러 메시지에 반드시 들어 있어야 할 조각들.
		// 비어 있으면 "에러가 없어야 한다"는 뜻이고, 개수는 곧 기대하는 문제 수다.
		want []string
	}{
		{
			name:   "정상 매니페스트는 통과한다",
			damage: nil,
			want:   nil,
		},
		{
			// M6 의 핵심 결정: 하나 찾을 때마다 반환하면 유저가 고치러 세 번 왕복한다.
			name: "세 곳이 한꺼번에 비면 세 개를 모아서 낸다",
			damage: func(m *Manifest) {
				m.Name = ""
				m.BaseURL = ""
				m.Commands = nil
			},
			want: []string{
				"name 이 비어 있다",
				"baseURL 이 비어 있다",
				"Command 가 비어 있다",
			},
		},
		{
			name: "같은 이름의 command 가 둘이면 잡는다",
			damage: func(m *Manifest) {
				m.Commands = append(m.Commands, Command{Name: "repo", Method: "GET", Path: "/x"})
			},
			want: []string{`"repo": 이름이 중복이다`},
		},
		{
			// required 가 아니면 유저가 안 줬을 때 빈 문자열이 치환돼 /repos//cobra 가
			// 나간다 — 8a 테이블 마지막에서 본 /pet/ 과 같은 사고.
			name: "path param 이 required 가 아니면 잡는다",
			damage: func(m *Manifest) {
				m.Commands[0].Params[0].Required = false
			},
			want: []string{`path param "owner" 는 required 여야 한다`},
		},
		{
			// 검증기는 BuildURL 과 같은 치환을 흉내 내 아는 자리를 지운다.
			// { 가 남으면 실행 시 그 자리가 그대로 URL 에 나간다는 뜻.
			name: "자리표시자에 대응 param 이 없으면 잡는다",
			damage: func(m *Manifest) {
				m.Commands[0].Params = m.Commands[0].Params[1:] // owner param 을 뺀다
			},
			want: []string{"대응 param 이 없는 자리표시자가 있다"},
		},
		{
			name: "method 는 대문자만 받는다",
			damage: func(m *Manifest) {
				m.Commands[0].Method = "get"
			},
			want: []string{`method "get" 는 지원하지 않는다`},
		},
		{
			// in 오타 하나가 문제 둘로 보고된다: 모르는 in 이라는 것과,
			// 그 결과 {owner} 를 치환할 param 이 없어졌다는 것. 유저에겐 둘 다 사실이다.
			name: "in 오타는 문제 둘로 보고된다",
			damage: func(m *Manifest) {
				m.Commands[0].Params[0].In = "quer"
			},
			want: []string{
				`in "quer" 는 지원하지 않는다`,
				"대응 param 이 없는 자리표시자가 있다",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := validManifest()
			if tt.damage != nil {
				tt.damage(m)
			}

			err := Validate(m)

			if len(tt.want) == 0 {
				if err != nil {
					t.Fatalf("Validate() = %v, want nil", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("Validate() = nil, want 문제 %d개", len(tt.want))
			}

			// errors.Join 이 묶은 에러는 Unwrap() []error 로 다시 펼 수 있다.
			// 그 메서드를 가졌는지만 물으면 되므로 타입 이름 대신 익명 인터페이스로 단언한다.
			joined, ok := err.(interface{ Unwrap() []error })
			if !ok {
				t.Fatalf("errors.Join 이 아닌 에러가 왔다: %T", err)
			}
			if n := len(joined.Unwrap()); n != len(tt.want) {
				t.Errorf("문제 %d개, want %d개\n실제:\n%s", n, len(tt.want), err)
			}

			for _, w := range tt.want {
				if !strings.Contains(err.Error(), w) {
					t.Errorf("메시지에 %q 가 없다\n실제:\n%s", w, err)
				}
			}
		})
	}
}
