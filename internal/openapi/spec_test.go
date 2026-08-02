package openapi

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

// 표적은 손으로 줄인 fixture 가 아니라 petstore 실물이다
// (https://petstore3.swagger.io/api/v3/openapi.json, OpenAPI 3.0.4).
// 줄인 fixture 는 우리가 이미 아는 것만 담아서, 정작 모르는 자리 — 우리가 칸을 안 판
// 키, 오타 낸 태그 — 를 잡지 못한다. 부분 디코드는 조용히 실패하므로 표적이 진짜여야 한다.
func loadPetstore(t *testing.T) *Spec {
	t.Helper()

	data, err := os.ReadFile("testdata/petstore.json")
	if err != nil {
		t.Fatal(err)
	}

	// .json 을 yaml.Unmarshal 로 읽는다 — YAML 1.2 가 JSON 의 상위집합이라
	// encoding/json 을 따로 들이지 않아도 된다.
	var s Spec
	if err := yaml.Unmarshal(data, &s); err != nil {
		t.Fatal(err)
	}
	return &s
}

func TestOperations(t *testing.T) {
	ops := loadPetstore(t).Operations()

	if len(ops) != 19 {
		t.Fatalf("operation 수 = %d, want 19", len(ops))
	}

	// 앞 5개만 못 박는다. 전부 적으면 spec 이 조금만 바뀌어도 통째로 깨지는데,
	// 확인하려는 것("경로는 사전순, 같은 경로 안에서는 메서드 선언순")은 5개로 드러난다.
	// operationId 까지 보는 이유는 부분 디코드의 오타가 조용하기 때문이다 —
	// 태그를 틀리면 빈 문자열이 되지 순회가 깨지지는 않는다.
	want := []Op{
		{Path: "/pet", Method: "POST", Operation: &Operation{OperationID: "addPet"}},
		{Path: "/pet", Method: "PUT", Operation: &Operation{OperationID: "updatePet"}},
		{Path: "/pet/findByStatus", Method: "GET", Operation: &Operation{OperationID: "findPetsByStatus"}},
		{Path: "/pet/findByTags", Method: "GET", Operation: &Operation{OperationID: "findPetsByTags"}},
		{Path: "/pet/{petId}", Method: "GET", Operation: &Operation{OperationID: "getPetById"}},
	}

	for i, w := range want {
		got := ops[i]
		if got.Path != w.Path || got.Method != w.Method || got.Operation.OperationID != w.Operation.OperationID {
			t.Errorf("ops[%d] = %s %s (%s), want %s %s (%s)",
				i, got.Method, got.Path, got.Operation.OperationID,
				w.Method, w.Path, w.Operation.OperationID)
		}
	}
}

// map range 는 무작위다 — 정렬을 빼먹으면 이 테스트가 잡는다.
// 가끔이 아니라 거의 매번 잡는다: 같은 프로세스 안에서도 순회마다 순서가 바뀐다.
func TestOperationsIsStable(t *testing.T) {
	s := loadPetstore(t)
	first := s.Operations()

	for try := range 20 {
		got := s.Operations()
		if len(got) != len(first) {
			t.Fatalf("%d번째 순회의 길이 = %d, 첫 순회 = %d", try, len(got), len(first))
		}
		for j := range first {
			if got[j] != first[j] {
				t.Fatalf("%d번째 순회에서 순서가 달라졌다: [%d] %s %s → %s %s",
					try, j, first[j].Method, first[j].Path, got[j].Method, got[j].Path)
			}
		}
	}
}
