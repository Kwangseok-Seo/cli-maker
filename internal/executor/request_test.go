package executor

import (
	"testing"

	"github.com/Kwangseok-Seo/cli-maker/internal/manifest"
)

// TestBuildURL 은 Command 와 유저가 준 값이 어떤 URL 이 되는지를 못 박는다.
//
// BaseURL 은 모든 케이스가 공유한다 — 이 테이블에서 확인하려는 건 path 치환과
// query 조립이지 BaseURL 이 붙는지가 아니다.
func TestBuildURL(t *testing.T) {
	m := &manifest.Manifest{BaseURL: "https://api.example.com/v3"}

	tests := []struct {
		name   string
		cmd    manifest.Command
		values map[string]string
		want   string
	}{
		{
			name:   "path param 하나를 치환한다",
			cmd:    manifest.Command{Path: "/pet/{petId}", Params: []manifest.Param{{Name: "petId", In: "path"}}},
			values: map[string]string{"petId": "10"},
			want:   "https://api.example.com/v3/pet/10",
		},
		{
			// 선언은 zebra 가 먼저인데 결과는 apple 이 먼저다 — url.Values.Encode 가
			// 키를 정렬하기 때문. 매니페스트에 적은 순서는 URL 에 남지 않는다.
			name: "query 는 매니페스트 순서가 아니라 이름 알파벳순으로 붙는다",
			cmd: manifest.Command{Path: "/search", Params: []manifest.Param{
				{Name: "zebra", In: "query"},
				{Name: "apple", In: "query"},
			}},
			values: map[string]string{"zebra": "1", "apple": "2"},
			want:   "https://api.example.com/v3/search?apple=2&zebra=1",
		},
		{
			// 유저가 안 준 flag 는 빈 문자열로 도착한다. 그걸 lang= 로 실어 보내면
			// "빈 값으로 필터해 달라"는 뜻이 돼 버리므로 아예 빼야 한다.
			name: "값이 빈 query param 은 URL 에 실리지 않는다",
			cmd: manifest.Command{Path: "/search", Params: []manifest.Param{
				{Name: "q", In: "query"},
				{Name: "lang", In: "query"},
			}},
			values: map[string]string{"q": "go", "lang": ""},
			want:   "https://api.example.com/v3/search?q=go",
		},
		{
			// path 자리의 슬래시는 경로 구분자로 오해되면 안 되므로 %2F 로 막는다.
			// 공백은 %20 — 바로 아래 query 케이스와 비교할 것.
			name: "path 값의 슬래시와 공백은 %2F, %20 이 된다",
			cmd: manifest.Command{Path: "/repos/{full}", Params: []manifest.Param{
				{Name: "full", In: "path"},
			}},
			values: map[string]string{"full": "spf13/cobra x"},
			want:   "https://api.example.com/v3/repos/spf13%2Fcobra%20x",
		},
		{
			// 같은 공백 한 칸인데 여기선 + 다. path 와 query 는 이스케이프 규칙이
			// 다르고(url.PathEscape vs url.Values.Encode), 그 차이를 이 두 줄이 고정한다.
			name: "query 값의 & 는 %26, 공백은 + 가 된다",
			cmd: manifest.Command{Path: "/search", Params: []manifest.Param{
				{Name: "q", In: "query"},
			}},
			values: map[string]string{"q": "a&b c"},
			want:   "https://api.example.com/v3/search?q=a%26b+c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildURL(m, &tt.cmd, tt.values)
			if got != tt.want {
				t.Errorf("BuildURL() = %s, want %s", got, tt.want)
			}
		})
	}
}
