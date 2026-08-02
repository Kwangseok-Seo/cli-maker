# 동적 명령 표면: 매니페스트 하나 = 2단 그룹, apis/ 스캔으로 발견

각 **Manifest** 를 `cli-maker <api> <command>` 형태의 **2단 그룹 명령**으로 노출하고, 매니페스트는 시작 시 `apis/` 디렉토리를 스캔해 발견한다 — 명령 이름 충돌을 구조적으로 불가능하게 만들고 "매니페스트 추가 = 새 API"([ADR-0001](0001-runtime-interpreter-architecture.md))의 약속을 유저 체험에서도 지키기 위해, 타이핑이 한 단어 길어지고 어떤 파일이 로드됐는지가 명령줄에 드러나지 않으며(암묵적 발견) 매 실행마다 모든 매니페스트를 파싱하는 비용을 감수한다.

거부한 대안: **flat 노출**(`cli-maker getPetById`)은 두 API 가 같은 command name 을 쓰는 순간 깨지고 `--help` 에서 소속이 사라진다. **`run` 아래 3단 격리**는 충돌은 막지만 매니페스트를 1급 시민이 아닌 인자로 강등한다 — cli-maker 자체 명령(`greet`, 후속 `generate`)과의 이름 공간 분리는 매니페스트 이름 예약으로 다루면 된다. **`--manifest` 단일 파일 전용**은 "여러 API 를 하나의 CLI 로"라는 그림을 표면에서 지운다(단, 발견 override 로는 유효 — 아래).

범위: M3 은 프로세스 cwd 기준 `./apis` 만 본다. 설치된 바이너리를 위한 발견 경로(환경변수, XDG config 디렉토리, `--manifest` override)와 그 우선순위는 M5(인증·설정)에서 정한다 — **실제로는 M5 에서 다루지 않았고, M12 에서 [ADR-0013](0013-manifest-discovery-paths.md) 이 유저 설정 디렉토리를 더한 합집합으로 정했다. 환경변수와 `--manifest` override 는 여전히 없다.**
