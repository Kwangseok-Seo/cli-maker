package clirun

import (
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"
	"testing"
)

// TestAliasFieldsAreDocumented 는 별칭 doc 에 적어 둔 필드 목록이 실제 필드와
// 갈리지 않게 붙든다.
//
// 왜 목록이 필요한가: 원본 타입이 internal/ 에 있어서 go doc 이 필드를 펼쳐 주지
// 않는다. 밖에서 이 패키지를 보는 사람에게는 doc 에 적힌 것이 전부다.
//
// 왜 테스트가 필요한가: 그 목록은 복제다. manifest 쪽에 필드가 하나 붙으면 doc 은
// 조용히 뒤처지고, 뒤처진 doc 은 코드가 아니라서 컴파일러도 vet 도 잡지 않는다.
// 필드 이름은 reflect 로 실물에서 끌어오고, doc 은 파서로 소스에서 끌어온다 —
// 양쪽 다 손으로 적지 않는다.
func TestAliasFieldsAreDocumented(t *testing.T) {
	doc := aliasDoc(t)

	// 별칭 넷을 zero value 로 만들어 실제 필드를 물어본다.
	for _, v := range []any{Manifest{}, Command{}, Param{}, Auth{}} {
		rt := reflect.TypeOf(v)
		for f := range rt.Fields() {
			if !strings.Contains(doc, f.Name) {
				t.Errorf("%s.%s 가 별칭 doc 에 없다 — clirun.go 의 type 블록 주석을 갱신해야 한다", rt.Name(), f.Name)
			}
		}
	}
}

// aliasDoc 은 clirun.go 의 type 블록에 붙은 doc 주석을 돌려준다.
//
// 파일을 문자열로 훑지 않고 파서에게 묻는다 — 주석이 어디 붙은 것인지는 위치가
// 아니라 구조가 정하기 때문이다.
func aliasDoc(t *testing.T) string {
	t.Helper()

	f, err := parser.ParseFile(token.NewFileSet(), "clirun.go", nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("clirun.go 를 파싱할 수 없다: %v", err)
	}

	for _, decl := range f.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE || gen.Doc == nil {
			continue
		}
		return gen.Doc.Text()
	}

	t.Fatal("clirun.go 에 doc 이 붙은 type 블록이 없다")
	return ""
}
