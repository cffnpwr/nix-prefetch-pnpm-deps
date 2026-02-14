package logger

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

// clearCIEnvVars clears CI specific environment variables so that they do not affect your tests.
func clearCIEnvVars(t *testing.T) {
	t.Helper()

	t.Setenv("CI", "")
	t.Setenv("GITHUB_ACTIONS", "")
	t.Setenv("GITLAB_CI", "")
	t.Setenv("TF_BUILD", "")
	t.Setenv("TEAMCITY_VERSION", "")
	t.Setenv("BUILDKITE", "")
	t.Setenv("TRAVIS", "")
}

func Test_isTruthy(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "[正常系] trueはtruthyである", value: "true", want: true},
		{name: "[正常系] TRUEはtruthyである", value: "TRUE", want: true},
		{name: "[正常系] Trueはtruthyである", value: "True", want: true},
		{name: "[正常系] 1はtruthyである", value: "1", want: true},
		{name: "[正常系] yesはtruthyである", value: "yes", want: true},
		{name: "[正常系] YESはtruthyである", value: "YES", want: true},
		{name: "[正常系] falseはtruthyでない", value: "false", want: false},
		{name: "[正常系] 0はtruthyでない", value: "0", want: false},
		{name: "[正常系] 空文字はtruthyでない", value: "", want: false},
		{name: "[正常系] noはtruthyでない", value: "no", want: false},
		{name: "[正常系] 任意の文字列はtruthyでない", value: "abc", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := isTruthy(tt.value)
			if got != tt.want {
				t.Errorf("isTruthy(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func Test_detectCI(t *testing.T) {
	tests := []struct {
		name   string
		envs   map[string]string
		wantCI CI
		wantOK bool
	}{
		{
			name:   "[正常系] 非CI環境の場合は空文字とfalseを返す",
			envs:   map[string]string{},
			wantCI: CI(""),
			wantOK: false,
		},
		{
			name:   "[正常系] GitHub Actionsを検出する",
			envs:   map[string]string{"GITHUB_ACTIONS": "true"},
			wantCI: gitHubActions,
			wantOK: true,
		},
		{
			name:   "[正常系] GitLab CIを検出する",
			envs:   map[string]string{"GITLAB_CI": "true"},
			wantCI: gitLabCI,
			wantOK: true,
		},
		{
			name:   "[正常系] Azure Pipelinesを検出する",
			envs:   map[string]string{"TF_BUILD": "true"},
			wantCI: azurePipelines,
			wantOK: true,
		},
		{
			name:   "[正常系] TeamCityを検出する",
			envs:   map[string]string{"TEAMCITY_VERSION": "2024.1"},
			wantCI: teamCity,
			wantOK: true,
		},
		{
			name:   "[正常系] Buildkiteを検出する",
			envs:   map[string]string{"BUILDKITE": "true"},
			wantCI: buildkite,
			wantOK: true,
		},
		{
			name:   "[正常系] Travis CIを検出する",
			envs:   map[string]string{"TRAVIS": "true"},
			wantCI: travisCI,
			wantOK: true,
		},
		{
			name:   "[正常系] CI環境変数のみの場合はothersを返す",
			envs:   map[string]string{"CI": "true"},
			wantCI: others,
			wantOK: true,
		},
		{
			name:   "[正常系] CI=1でもothersとして検出する",
			envs:   map[string]string{"CI": "1"},
			wantCI: others,
			wantOK: true,
		},
		{
			name: "[正常系] 固有環境変数がCI環境変数より優先される",
			envs: map[string]string{
				"CI":             "true",
				"GITHUB_ACTIONS": "true",
			},
			wantCI: gitHubActions,
			wantOK: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearCIEnvVars(t)
			for k, v := range tt.envs {
				t.Setenv(k, v)
			}

			gotCI, gotOK := detectCI()
			if d := cmp.Diff(tt.wantCI, gotCI); d != "" {
				t.Errorf("detectCI() CI mismatch (-want +got):\n%s", d)
			}
			if gotOK != tt.wantOK {
				t.Errorf("detectCI() isCI = %v, want %v", gotOK, tt.wantOK)
			}
		})
	}
}
