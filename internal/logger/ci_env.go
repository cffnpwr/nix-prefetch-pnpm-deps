package logger

import (
	"os"
	"strings"
)

type CI string

const (
	gitHubActions  CI = "github_actions"
	gitLabCI       CI = "gitlab_ci"
	azurePipelines CI = "azure_pipelines"
	teamCity       CI = "teamcity"
	buildkite      CI = "buildkite"
	travisCI       CI = "travis_ci"
	others         CI = "others"

	ciEnv             = "CI"
	gitHubActionsEnv  = "GITHUB_ACTIONS"
	gitLabCIEnv       = "GITLAB_CI"
	azurePipelinesEnv = "TF_BUILD"
	teamCityEnv       = "TEAMCITY_VERSION"
	buildkiteEnv      = "BUILDKITE"
	travisCIEnv       = "TRAVIS"
)

// isTruthy returns `true` if environment variable has truthy value.
func isTruthy(val string) bool {
	switch strings.ToLower(val) {
	case "true", "1", "yes":
		return true
	default:
		return false
	}
}

func detectCI() (CI, bool) {
	switch {
	case isTruthy(os.Getenv(gitHubActionsEnv)):
		return gitHubActions, true
	case isTruthy(os.Getenv(gitLabCIEnv)):
		return gitLabCI, true
	case isTruthy(os.Getenv(azurePipelinesEnv)):
		return azurePipelines, true
	case os.Getenv(teamCityEnv) != "":
		return teamCity, true
	case isTruthy(os.Getenv(buildkiteEnv)):
		return buildkite, true
	case isTruthy(os.Getenv(travisCIEnv)):
		return travisCI, true
	case isTruthy(os.Getenv(ciEnv)):
		return others, true
	default:
		return CI(""), false
	}
}
