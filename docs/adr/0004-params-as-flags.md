# 모든 Param 은 flag 로 노출한다

매니페스트의 모든 **Param** 을 `in:` 과 무관하게 `--이름 값` 플래그로 노출한다 — 인자 순서를 틀려도 잘못된 URL 이 조용히 나가는 일을 구조적으로 불가능하게 만들고 cobra 의 필수 검증(`MarkFlagRequired`)을 그대로 쓰기 위해, URL 을 닮은 짧은 표기(`getPetById 10`)를 포기하고 타이핑이 길어지는 것을 감수한다.

거부한 대안: **path param 만 positional.** 실물 비교에서 path param 이 둘인 명령(`/repos/{owner}/{repo}`)에 순서를 바꿔 넣자 아무 에러 없이 `/repos/cobra/spf13` 이 조립됐다 — flag 방식은 이름이 붙어 있어 같은 실수가 성립하지 않는다. 더구나 positional 의 순서 기준을 `params:` 순으로 할지 `path` 문자열의 `{}` 등장 순으로 할지가 우리 구현의 선택이 되어, 매니페스트 작성자가 둘을 다르게 적으면 같은 방식으로 조용히 뒤바뀐다. 대조한 출력은 [M4 로그](../learn/milestones/M4.md).

부속 결정: flag 는 `type` 과 무관하게 전부 **문자열**로 받는다. 정수 flag 는 미지정을 `0` 으로 돌려주어 `?count=0` 을 명시한 것과 구별할 수 없는데, 빈 문자열을 "유저가 주지 않음"의 신호로 삼는 규약이 `BuildURL` 의 query 생략 규칙을 지탱한다. `type` 의 실제 검증은 후속.

범위: 이 결정은 CLI 표면에 관한 것이고, `in:` 값(path/query/header/body)은 요청 조립에서 계속 쓰인다 — 유저에게 안 보일 뿐 사라지지 않는다.
